/*
uniwish.com/internat/api/services/item

contains logic for handling our items
*/
package services

import (
	"context"
	"net/url"

	"uniwish.com/internal/api/errors"
)

type ItemService struct{}

func NewItemsService() *ItemService {
	return &ItemService{}
}

func (s *ItemService) Create(ctx context.Context, rawUrl string) (string, error) {
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

	return "fakeid", nil // TODO: when db defined, change this to return created / existing item id
}
