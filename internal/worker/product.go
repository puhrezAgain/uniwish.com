/*
uniwish.com/interal/api/repository/product

db logic for products and prices
*/
package worker

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"uniwish.com/internal/api/repository"
	"uniwish.com/internal/domain"
)

type ProductWriter interface {
	UpsertProduct(context.Context, domain.ProductSnapshot) (uuid.UUID, error)
	InsertPrice(context.Context, []domain.Offer) error
}

type DefaultProductWriter struct {
	db repository.DB
}

func NewProductWriter(db repository.DB) ProductWriter {
	return &DefaultProductWriter{db: db}
}

func (pr *DefaultProductWriter) UpsertProduct(ctx context.Context, product domain.ProductSnapshot) (uuid.UUID, error) {
	_, err := pr.db.ExecContext(ctx,
		`
	INSERT INTO products
	(id, store, store_product_id, name, image_url, url)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (store, store_product_id)
	DO UPDATE SET
		name = EXCLUDED.name,
		updated_at = now()
	RETURNING id
	`, product.ID, product.Store, product.SKU, product.Name, product.ImageURL, product.URL)

	if err != nil {
		return uuid.Nil, fmt.Errorf("product upsert error: %ws", err)
	}
	return product.ID, nil
}
func (pr *DefaultProductWriter) InsertPrice(ctx context.Context, offers []domain.Offer) error {
	offersCount := len(offers)
	if offersCount == 0 {
		return nil
	}

	ids := make([]string, 0, offersCount)
	productIds := make([]string, 0, offersCount)
	prices := make([]float64, 0, offersCount)
	currencies := make([]string, 0, offersCount)

	for _, o := range offers {
		ids = append(ids, o.ID.String())
		productIds = append(productIds, o.ProductID.String())
		prices = append(prices, o.Price)
		currencies = append(currencies, o.Currency)
	}
	_, err := pr.db.ExecContext(ctx,
		`
	INSERT INTO prices (id, product_id, price, currency)
	SELECT * FROM UNNEST($1::uuid[], $2::uuid[], $3::float[], $4::text[])
	`, pq.Array(ids), pq.Array(productIds), pq.Array(prices), pq.Array(currencies))

	if err != nil {
		return fmt.Errorf("price insert error: %w", err)
	}

	return nil
}
