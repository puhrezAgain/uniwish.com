/*
uniwish.com/interal/api/repository/product

db logic for products and prices
*/
package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"uniwish.com/internal/domain"
)

type ProductRepository interface {
	UpsertProduct(context.Context, domain.ProductSnapshot) (uuid.UUID, error)
	InsertPrice(context.Context, uuid.UUID, float32, string) error
}

type DefaultProductRepository struct {
	db DB
}

func NewProductRepository(db DB) ProductRepository {
	return &DefaultProductRepository{db: db}
}

func (pr *DefaultProductRepository) UpsertProduct(ctx context.Context, product domain.ProductSnapshot) (uuid.UUID, error) {
	_, err := pr.db.ExecContext(ctx,
		`
	INSERT INTO products
	(id, store, store_product_id, name, image_url, url)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (store, store_product_id)
	DO UPDATE SET
		name = EXCLUDED.name
		update_now = now()
	RETURNING id
	`, product.ID, product.Store, product.StoreProductID, product.Name, product.ImageURL, product.URL)

	if err != nil {
		return uuid.Nil, fmt.Errorf("product upsert error %ws", err)
	}
	return product.ID, nil
}
func (pr *DefaultProductRepository) InsertPrice(ctx context.Context, productID uuid.UUID, price float32, currency string) error {
	_, err := pr.db.ExecContext(ctx,
		`
	INSERT INTO prices (id, product_id, price, currency, scraped_at)
	VALUES ($1, $2, $3, $4, now())
	`, uuid.New(), productID, price, currency)

	if err != nil {
		return fmt.Errorf("price insert error %w", err)
	}

	return nil
}
