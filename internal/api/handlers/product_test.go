/*
uniwish.com/interal/api/handlers/product_tests

tests for product endpoint
*/
package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"uniwish.com/internal/api/repository"
	"uniwish.com/internal/api/services"
)

type SuccessfulRepo struct{}

func (p *SuccessfulRepo) ListProducts(ctx context.Context) ([]repository.ProductListItem, error) {
	return []repository.ProductListItem{
		{
			Name:      "fake product 1",
			Store:     "Fake store",
			ImageURL:  "whatever.com/i.img",
			LastPrice: 34.2,
			Currency:  "EUR",
		},
		{
			Name:      "fake product 2",
			Store:     "Fake store",
			ImageURL:  "whatever.com/i.img",
			LastPrice: 35.2,
			Currency:  "EUR",
		},
	}, nil
}

func (p *SuccessfulRepo) GetProduct(ctx context.Context, id uuid.UUID) (*repository.ProductDetail, error) {
	return &repository.ProductDetail{
		Name:     "fake product",
		Store:    "fakestore",
		ImageURL: "fakestore.com/fakeproduct.jpeg",
		Offers: []repository.OfferListItem{
			{
				Price:     34.42,
				Currency:  "EUR",
				UpdatedAt: time.Now(),
			},
			{
				Price:     42.32,
				Currency:  "EUR",
				UpdatedAt: time.Now().Add(-3 * time.Hour),
			},
		},
	}, nil
}

type EmptyRepo struct{}

func (p *EmptyRepo) ListProducts(ctx context.Context) ([]repository.ProductListItem, error) {
	return make([]repository.ProductListItem, 0), nil
}

func (p *EmptyRepo) GetProduct(ctx context.Context, id uuid.UUID) (*repository.ProductDetail, error) {
	return nil, sql.ErrNoRows
}

type FaultyRepo struct{}

func (p *FaultyRepo) ListProducts(ctx context.Context) ([]repository.ProductListItem, error) {
	return nil, sql.ErrConnDone
}

func (p *FaultyRepo) GetProduct(ctx context.Context, id uuid.UUID) (*repository.ProductDetail, error) {
	return nil, sql.ErrConnDone
}

func TestListProducts(t *testing.T) {
	tests := []struct {
		name               string
		repo               repository.ProductReader
		expectedStatusCode int
	}{
		{name: "success", expectedStatusCode: 200, repo: &SuccessfulRepo{}},
		{name: "not_found", expectedStatusCode: 200, repo: &EmptyRepo{}},
		{name: "internal_error", expectedStatusCode: 500, repo: &FaultyRepo{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			hdlr := &DefaultProductHandler{
				service: services.NewDefaultProductReaderService(tt.repo),
			}

			req := httptest.NewRequest(http.MethodGet, "/products", nil)
			rr := httptest.NewRecorder()

			hdlr.ListProducts(rr, req)

			if rr.Code != tt.expectedStatusCode {
				t.Fatalf("expected status %d, got %d", tt.expectedStatusCode, rr.Code)

				if rr.Code == http.StatusOK {
					var resp services.ProductListResponse
					if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
						t.Fatalf("invalid json: %v", err)
					}
					if len(resp.Products) != 2 {
						t.Fatalf("execpted 2 productd, received: %d", len(resp.Products))
					}
				}

			}
		})
	}
}
func TestGetProduct(t *testing.T) {
	productId := uuid.New()
	tests := []struct {
		name               string
		repo               repository.ProductReader
		id                 string
		expectedStatusCode int
	}{
		{name: "success", expectedStatusCode: 200, id: productId.String(), repo: &SuccessfulRepo{}},
		{name: "not_found", expectedStatusCode: 404, id: uuid.NewString(), repo: &EmptyRepo{}},
		{name: "invalid_id", expectedStatusCode: 400, id: "whatever", repo: &EmptyRepo{}},
		{name: "internal_error", expectedStatusCode: 500, id: uuid.NewString(), repo: &FaultyRepo{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			hdlr := &DefaultProductHandler{
				service: services.NewDefaultProductReaderService(tt.repo),
			}
			mux := http.NewServeMux()
			mux.HandleFunc("GET /products/{id}", hdlr.GetProduct)

			req := httptest.NewRequest(http.MethodGet, "/products/"+tt.id, nil)
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatusCode {
				t.Fatalf("expected status %d, got %d", tt.expectedStatusCode, rr.Code)

				if rr.Code == http.StatusOK {
					var resp services.ProductDetailResponse
					if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
						t.Fatalf("invalid json: %v", err)
					}
					if len(resp.Offers) != 2 {
						t.Fatalf("expeceted 2 offers, recevied: %d", len(resp.Offers))
					}
				}
			}
		})
	}
}
