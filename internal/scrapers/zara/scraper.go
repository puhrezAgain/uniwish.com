/*
uniwish.com/internal/scrapers/zara/scraper

contains logic related to http interaction with zara pages
*/

package zara

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"uniwish.com/internal/domain"
)

type Scraper struct {
	client *http.Client
}

func New(timeout time.Duration) *Scraper {
	return &Scraper{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (s *Scraper) Scrape(ctx context.Context, URL string) (domain.ProductSnapshot, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, URL, nil)
	if err != nil {
		return domain.ProductSnapshot{}, fmt.Errorf("request error: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return domain.ProductSnapshot{}, fmt.Errorf("request error: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.ProductSnapshot{}, fmt.Errorf("Unexpected http error")
	}

	return parseProduct(resp.Body)
}
