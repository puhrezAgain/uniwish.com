/*
uniwish.com/internat/api/services/scape_request

contains logic for handling our scape_requests
*/
package services

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"uniwish.com/internal/api/errors"
	"uniwish.com/internal/api/repository"
	"uniwish.com/internal/scrapers"
	"uniwish.com/internal/scrapers/zara"
)

type ScrapeRequester interface {
	Request(ctx context.Context, rawUrl string) (uuid.UUID, error)
}
type ScrapeRequestService struct {
	repo repository.ScrapeRequestRepository
}

func NewScrapeRequestService(r repository.ScrapeRequestRepository) *ScrapeRequestService {
	return &ScrapeRequestService{repo: r}
}

func NewScraper(URL string) (scrapers.Scraper, error) {
	parsed, err := url.Parse(URL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.ErrInputInvalid
	}

	switch {
	case strings.Contains(parsed.Host, "zara.com"):
		// TODO should timeout be configuration based?
		return zara.NewZaraScraper(10 * time.Second), nil
	case parsed.Host == "store.com":
		// TODO, perhaps dependency inject this map to make monkey patching trivial?
		return scrapers.NewDefaultScraper(), nil
	default:
		return nil, errors.ErrStoreUnsupported
	}
}

func (s *ScrapeRequestService) Request(ctx context.Context, rawUrl string) (uuid.UUID, error) {
	// validates the rawUrl and if ok, inserts a scrape request to the db
	if rawUrl == "" {
		return uuid.Nil, errors.ErrInputInvalid
	}
	_, err := NewScraper(rawUrl)

	if err != nil {
		return uuid.Nil, err
	}

	return s.repo.Insert(ctx, rawUrl)
}
