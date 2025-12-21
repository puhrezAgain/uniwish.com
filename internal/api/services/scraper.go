/*
uniwish.com/internal/api/services/scraper

contains logic related to scrapers and scraping
*/
package services

import (
	"context"
	"net/url"

	"uniwish.com/internal/api/errors"
)

type PlaceholderProduct struct {
	// TODO: placeholder struct
	URL string
}
type BaseScraper interface {
	Scrape(ctx context.Context, url string) (*PlaceholderProduct, error)
}

type Scraper struct{}

func NewScraper(URL string) (BaseScraper, error) {
	parsed, err := url.Parse(URL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.ErrInputInvalid
	}

	if parsed.Host != "store.com" { // TODO: when scrapers defined, change this to ensure host maps to a scraper
		return nil, errors.ErrStoreUnsupported
	}

	return &Scraper{}, nil
}

func (s *Scraper) Scrape(ctx context.Context, url string) (*PlaceholderProduct, error) {
	return nil, nil
}
