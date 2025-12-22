package domain

import "github.com/google/uuid"

type ProductSnapshot struct {
	ID             uuid.UUID
	Name           string
	Store          string
	StoreProductID string
	URL            string
	ImageURL       string
	Price          float32
	Currency       string
}
