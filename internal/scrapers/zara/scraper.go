/*
uniwish.com/internal/scrapers/zara/scraper

contains logic related to http interaction with zara pages
*/

package zara

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/html"
	"uniwish.com/internal/domain"
)

type ZaraProductJSON struct {
	Type     string `json:"@type"`
	Name     string `json:"name"`
	SKU      string `json:"sku"`
	ImageURL string `json:"image"`
	Size     string `json:"size"`
	Color    string `json:"color"`
	Offer    struct {
		Price        string `json:"price"`
		Currency     string `json:"priceCurrency"`
		Availability string `json:"availability"`
	} `json:"offers"`
}

type ZaraScraper struct {
	client *http.Client
}

func NewZaraScraper(timeout time.Duration) *ZaraScraper {
	return &ZaraScraper{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (s *ZaraScraper) Fetch(ctx context.Context, URL string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, URL, nil)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected http error")
	}

	return resp.Body, nil
}

func (s *ZaraScraper) Scrape(ctx context.Context, URL string) (*domain.ProductRecord, error) {
	pageBody, err := s.Fetch(ctx, URL)
	if err != nil {
		return nil, err
	}
	defer pageBody.Close()

	product, offers, err := s.ParseProduct(pageBody)
	product.URL = URL
	return &domain.ProductRecord{Product: product, Offers: offers}, nil
}

func (s *ZaraScraper) ParseProduct(page io.Reader) (*domain.ProductSnapshot, *[]domain.Offer, error) {
	products, err := s.extractProductsFromPage(page)
	if err != nil {
		return nil, nil, err
	}
	productId := uuid.New()
	productSnapshot := domain.ProductSnapshot{
		ID:       productId,
		Name:     products[0].Name,
		SKU:      products[0].SKU,
		Store:    "zara",
		ImageURL: products[0].ImageURL,
	}

	var offers []domain.Offer
	for _, productOffering := range products {
		price, err := strconv.ParseFloat(productOffering.Offer.Price, 64)
		if err != nil {
			return nil, nil, err
		}
		availability := strings.TrimPrefix(productOffering.Offer.Availability, "https://schema.org/")
		offers = append(offers,
			domain.Offer{
				ID:           uuid.New(),
				ProductID:    productId,
				Price:        price,
				Currency:     productOffering.Offer.Currency,
				Size:         productOffering.Size,
				Color:        productOffering.Color,
				Availability: availability,
			},
		)
	}
	return &productSnapshot, &offers, nil
}

func (s *ZaraScraper) extractProductsFromPage(page io.Reader) ([]ZaraProductJSON, error) {
	parsed, err := html.Parse(page)
	if err != nil {
		return nil, err
	}

	var walk func(*html.Node) []ZaraProductJSON
	walk = func(n *html.Node) []ZaraProductJSON {
		// TODO for now return first product ld+json
		if results, ok := s.parseProductLDJSON(n); ok {
			return results
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if results := walk(c); results != nil {
				return results
			}
		}

		return nil
	}

	products := walk(parsed)

	if products == nil {
		return nil, errors.New("no ld+json product found")
	}

	return products, nil
}

func (*ZaraScraper) parseProductLDJSON(n *html.Node) ([]ZaraProductJSON, bool) {
	if n.Type == html.ElementNode && n.Data == "script" {
		for _, attr := range n.Attr {
			if attr.Key == "type" && attr.Val == "application/ld+json" {
				if n.FirstChild != nil {
					var ps []ZaraProductJSON
					data := []byte(n.FirstChild.Data)
					err := json.Unmarshal(data, &ps)

					if err == nil && len(ps) > 0 && ps[0].Type == "Product" {
						return ps, true
					}
				}
				break
			}
		}
	}
	return nil, false
}
