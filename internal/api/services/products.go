/*
uniwish.com/internal/api/services/products

contains logic of application's product service
*/
package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	apiErrors "uniwish.com/internal/api/errors"
	"uniwish.com/internal/api/repository"
)

type ProductListItemResponse struct {
	Name      string  `json:"name"`
	Store     string  `json:"store"`
	ImageURL  string  `json:"image_url"`
	LastPrice float64 `json:"last_price"`
	Currency  string  `json:"currency"`
}

type ProductListResponse struct {
	Products []ProductListItemResponse `json:"products"`
}

type ProductItemResponse struct {
	Name     string `json:"name"`
	Store    string `json:"store"`
	ImageURL string `json:"image_url"`
}

type OfferListResponse struct {
	Price        float64   `json:"price"`
	Currency     string    `json:"currency"`
	Availability string    `json:"availability"`
	UpdatedAt    time.Time `json:"updated_at"`
}
type ProductDetailResponse struct {
	Product ProductItemResponse `json:"product"`
	Offers  []OfferListResponse `json:"offers"`
}

type ProductReaderService interface {
	Get(context.Context, uuid.UUID) (*ProductDetailResponse, error)
	List(context.Context) (*ProductListResponse, error)
}
type DefaultProductReaderService struct {
	repo repository.ProductReader
}

func (s *DefaultProductReaderService) List(ctx context.Context) (*ProductListResponse, error) {
	list, err := s.repo.ListProducts(ctx)
	if err != nil {
		return nil, err
	}

	plr := &ProductListResponse{
		Products: make([]ProductListItemResponse, 0, len(list)),
	}

	for _, product := range list {
		pli := ProductListItemResponse{
			Name:      product.Name,
			Store:     product.Store,
			ImageURL:  product.ImageURL,
			LastPrice: product.LastPrice,
			Currency:  product.Currency,
		}
		plr.Products = append(plr.Products, pli)
	}

	return plr, nil
}

func (s *DefaultProductReaderService) Get(ctx context.Context, uuid uuid.UUID) (*ProductDetailResponse, error) {
	product, err := s.repo.GetProduct(ctx, uuid)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apiErrors.ErrNoProductFound
		}
		return nil, err
	}

	if product == nil {
		return nil, apiErrors.ErrNoProductFound
	}

	pir := ProductItemResponse{
		Name:     product.Name,
		Store:    product.Store,
		ImageURL: product.ImageURL,
	}

	offers := make([]OfferListResponse, 0, len(product.Offers))

	for _, o := range product.Offers {
		olr := OfferListResponse{
			Price:        o.Price,
			Currency:     o.Currency,
			Availability: o.Availability,
			UpdatedAt:    o.UpdatedAt,
		}
		offers = append(offers, olr)
	}
	pdr := &ProductDetailResponse{
		Product: pir,
		Offers:  offers,
	}
	return pdr, nil
}

func NewDefaultProductReaderService(repo repository.ProductReader) ProductReaderService {
	return &DefaultProductReaderService{repo: repo}
}
