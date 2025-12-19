/*
uniwish.com/internat/api/services/scape_request

contains logic for handling our scape_requests
*/
package services

import (
	"context"
	"net/url"

	"github.com/google/uuid"
	"uniwish.com/internal/api/errors"
	"uniwish.com/internal/api/repository"
)

type ScrapeRequestService struct {
	repo repository.ScrapeRequestRepository
}

func NewScrapeRequestService(r repository.ScrapeRequestRepository) *ScrapeRequestService {
	return &ScrapeRequestService{repo: r}
}

func (s *ScrapeRequestService) Request(ctx context.Context, rawUrl string) (uuid.UUID, error) {
	if rawUrl == "" {
		return uuid.Nil, errors.ErrInputInvalid
	}

	parsed, err := url.Parse(rawUrl)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return uuid.Nil, errors.ErrInputInvalid
	}

	if parsed.Host != "store.com" { // TODO: when scrapers defined, change this to ensure host maps to a scraper
		return uuid.Nil, errors.ErrStoreUnsupported
	}

	return s.repo.Insert(ctx, rawUrl)
}
