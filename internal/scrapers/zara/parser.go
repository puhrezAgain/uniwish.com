/*
uniwish.com/internal/scrapers/zara/parser

contains logic related to scraping zara pages
*/

package zara

import (
	"io"

	"github.com/google/uuid"
	"uniwish.com/internal/domain"
)

func parseProduct(r io.Reader) (domain.ProductSnapshot, error) {
	return domain.ProductSnapshot{
		ID:             uuid.New(),
		Store:          "zara",
		StoreProductID: "12345",
		Name:           "Zara jacket",
		ImageURL:       "http://zara.com/img.jpeg",
		Price:          45.32,
		Currency:       "euro",
	}, nil
}
