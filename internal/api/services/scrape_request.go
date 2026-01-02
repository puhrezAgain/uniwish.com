/*
uniwish.com/internat/api/services/scape_request

contains logic for handling our scape_requests
*/
package services

import (
	"context"
	goErrors "errors"

	"github.com/google/uuid"
	"uniwish.com/internal/api/errors"
	"uniwish.com/internal/api/repository"
	"uniwish.com/internal/scrapers"
)

type ScrapeRequester interface {
	Request(ctx context.Context, rawUrl string) (uuid.UUID, error)
}
type ScrapeRequestService struct {
	repo     repository.ScrapeRequestRepository
	registry scrapers.Registry
}

func NewScrapeRequestService(sr repository.ScrapeRequestRepository, registry scrapers.Registry) *ScrapeRequestService {
	return &ScrapeRequestService{repo: sr, registry: registry}
}

func (s *ScrapeRequestService) Request(ctx context.Context, rawUrl string) (uuid.UUID, error) {
	// validates the rawUrl and if ok, inserts a scrape request to the db
	if rawUrl == "" {
		return uuid.Nil, errors.ErrInputInvalid
	}

	err := s.registry.ValidateUrl(rawUrl)

	switch {
	case goErrors.Is(err, scrapers.ErrInvalidURL):
		return uuid.Nil, errors.ErrInputInvalid
	case goErrors.Is(err, scrapers.ErrNoScraper):
		return uuid.Nil, errors.ErrStoreUnsupported
	default:
		return s.repo.Insert(ctx, rawUrl)
	}
}
