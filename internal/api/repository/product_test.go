/*
uniwish.com/internal/api/repository/product_test

testing for product reader repo
*/
package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"uniwish.com/internal/testutil"
)

func NewFakeProductDetail() *ProductDetail {
	return &ProductDetail{
		Name:     "fake product",
		Store:    "fakestore",
		ImageURL: "fakestore.com/fakeproduct.jpeg",
		Offers: []OfferListItem{
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
	}
}
func insertFakeProductDetailItem(productId uuid.UUID, pd *ProductDetail, ctx context.Context, db DB) error {
	_, err := db.ExecContext(
		ctx,
		`
		INSERT INTO products
		(id, store, store_product_id, name, image_url, url)
		VALUES ($1, $2, $3, $4, $5, $6)
		`,
		productId, pd.Store, pd.Store+"Id", pd.Name, pd.ImageURL, pd.Store+".com/test",
	)
	if err != nil {
		return err
	}

	for _, o := range pd.Offers {
		_, err := db.ExecContext(
			ctx,
			`
			INSERT INTO prices
			(id, product_id, price, currency, scraped_at)
			VALUES ($2, $1, $3, $4, $5)
			`, productId, uuid.New(), o.Price, o.Currency, o.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
func TestProductReader_ListProducts(t *testing.T) {
	testutil.RequireIntegration(t)
	testutil.TruncateTables(t, testDB)
	t.Cleanup(func() {
		testutil.TruncateTables(t, testDB)
	})
	expectedProductDetail := NewFakeProductDetail()

	insertFakeProductDetailItem(uuid.New(), expectedProductDetail, context.Background(), testDB)

	pr := &DefaultProductReader{db: testDB}

	result, err := pr.ListProducts(context.Background())

	if err != nil {
		t.Fatalf("error not nil: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	if len(result) != 1 {
		t.Fatalf("not 1 product, count: %d", len(result))
	}
	if result[0].LastPrice != expectedProductDetail.Offers[0].Price {
		t.Fatalf("expected product last price %f, received: %f",
			expectedProductDetail.Offers[0].Price, result[0].LastPrice)

	}

}

func TestProductReader_GetProduct(t *testing.T) {
	testutil.RequireIntegration(t)
	testutil.TruncateTables(t, testDB)
	t.Cleanup(func() {
		testutil.TruncateTables(t, testDB)
	})

	productId := uuid.New()

	expectedProductDetail := NewFakeProductDetail()
	insertFakeProductDetailItem(productId, expectedProductDetail, context.Background(), testDB)

	pr := &DefaultProductReader{db: testDB}
	result, err := pr.GetProduct(context.Background(), productId)

	if err != nil {
		t.Fatalf("error not nil: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	if result.Name != expectedProductDetail.Name {
		t.Fatalf("result name %s different than expected: %s", result.Name, expectedProductDetail.Name)
	}
	if result.Store != expectedProductDetail.Store {
		t.Fatalf("result store %s different than expected: %s", result.Name, expectedProductDetail.Store)
	}

	if len(result.Offers) != 2 {
		t.Fatalf("Expected 2 offers, received %d", len(result.Offers))
	}
}
