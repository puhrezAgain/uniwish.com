package domain

import "github.com/google/uuid"

type Offer struct {
	ID           uuid.UUID
	ProductID    uuid.UUID
	Price        float64
	Currency     string
	Size         string
	Color        string
	Availability string
}

type ProductSnapshot struct {
	ID       uuid.UUID
	URL      string
	Name     string
	Store    string
	SKU      string
	ImageURL string
}

type ProductRecord struct {
	Product *ProductSnapshot
	Offers  *[]Offer
}
