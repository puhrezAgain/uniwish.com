/*
uniwish.com/internal/api/repository/products

contains logic of application's product read layer
*/
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type ProductListItem struct {
	Name      string
	Store     string
	ImageURL  string
	LastPrice float64
	Currency  string
}

type ProductDetail struct {
	Name     string
	Store    string
	ImageURL string
	Offers   []OfferListItem
}

type OfferListItem struct {
	Price        float64
	Currency     string
	Availability string
	UpdatedAt    time.Time
}

type ProductReader interface {
	ListProducts(ctx context.Context) ([]ProductListItem, error)
	GetProduct(ctx context.Context, id uuid.UUID) (*ProductDetail, error)
}

type DefaultProductReader struct {
	db DB
}

func NewDefaultProductReader(db DB) ProductReader {
	return &DefaultProductReader{db: db}
}

func (p *DefaultProductReader) ListProducts(ctx context.Context) ([]ProductListItem, error) {
	// TODO: add product_id, scraped_at, currency index to DB

	rows, err := p.db.QueryContext(
		ctx,
		`
		SELECT p.name, p.store, p.image_url, pr.price, pr.currency
		FROM products p 
		LEFT JOIN LATERAL (
		SELECT price, currency
		FROM prices 
		WHERE product_id = p.id
		ORDER BY scraped_at DESC
		LIMIT 1
		) pr ON TRUE
		`,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var products []ProductListItem

	for rows.Next() {
		var p ProductListItem
		if err := rows.Scan(
			&p.Name, &p.Store,
			&p.ImageURL, &p.LastPrice, &p.Currency); err != nil {
			return nil, err
		}

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}

func (p *DefaultProductReader) GetProduct(ctx context.Context, id uuid.UUID) (*ProductDetail, error) {
	// TODO: use sql to grab basic info for product and each offer
	rows, err := p.db.QueryContext(ctx,
		`
		SELECT p.name, p.store, p.image_url, pr.price, pr.currency, pr.scraped_at
		FROM products p JOIN prices pr ON p.id = pr.product_id
		WHERE p.id = $1
		ORDER BY pr.scraped_at ASC
		`, id,
	)

	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	defer rows.Close()

	var pd *ProductDetail
	for rows.Next() {
		var (
			name      string
			store     string
			image_url string
			offer     OfferListItem
		)
		if err := rows.Scan(
			&name, &store, &image_url, &offer.Price,
			&offer.Currency, &offer.UpdatedAt); err != nil {
			return nil, err
		}

		if pd == nil {
			pd = &ProductDetail{Name: name, Store: store, ImageURL: image_url, Offers: make([]OfferListItem, 0)}
		}

		pd.Offers = append(pd.Offers, offer)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if pd == nil {
		return nil, sql.ErrNoRows
	}
	return pd, nil
}
