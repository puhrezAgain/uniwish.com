/*
uniwish.com/internat/api/services/scape_request

contains logic for handling our scape_requests
*/
package services

import (
	"context"
	"net/url"

	"uniwish.com/internal/api/errors"
)

type ScrapeRequestService struct{}

func NewScrapeRequestService() *ScrapeRequestService {
	return &ScrapeRequestService{}
}

func (s *ScrapeRequestService) Create(ctx context.Context, rawUrl string) (string, error) {
	if rawUrl == "" {
		return "", errors.ErrInputInvalid
	}

	url, err := url.Parse(rawUrl)
	if err != nil || url.Scheme == "" || url.Host == "" {
		return "", errors.ErrInputInvalid
	}

	if url.Host != "store.com" { // TODO: when scrapers defined, change this to ensure host maps to a scraper
		return "", errors.ErrStoreUnsupported
	}

	// TODO: insert a scrape request

	return "fakeid", nil // TODO: when db defined, change this to return created / existing item id
}
