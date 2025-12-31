/*
uniwish.com/internal/scrapers/zara/scraper

contains logic related to http interaction with zara pages
*/

package scrapers

import (
	"context"
	"errors"
	"io"

	"uniwish.com/internal/domain"
)

type Scraper interface {
	Scrape(context.Context, string) (*domain.ProductRecord, error)
	Fetch(context.Context, string) (io.ReadCloser, error)
	ParseProduct(io.Reader) (*domain.ProductSnapshot, *[]domain.Offer, error)
}

type DefaultScraper struct{}

func NewDefaultScraper() *DefaultScraper {
	return &DefaultScraper{}
}

func (s *DefaultScraper) Fetch(_ context.Context, _ string) (io.ReadCloser, error) {
	return nil, nil
}

func (s *DefaultScraper) Scrape(_ context.Context, _ string) (*domain.ProductRecord, error) {
	return nil, errors.ErrUnsupported
}

func (s *DefaultScraper) ParseProduct(_ io.Reader) (*domain.ProductSnapshot, *[]domain.Offer, error) {
	return nil, nil, nil
}
