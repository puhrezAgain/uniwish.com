/*
uniwish.com/internat/api/services/scape_request

contains logic for handling our scape_requests
*/
package services

import (
	"context"

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
	_, err := NewScraper(rawUrl)

	if err != nil {
		return uuid.Nil, err
	}

	return s.repo.Insert(ctx, rawUrl)
}
