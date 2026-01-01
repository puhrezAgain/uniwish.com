/*
uniwish.com/interal/api/services/scrape_request_test

tests for scrape request service
*/
package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	apiErrors "uniwish.com/internal/api/errors"
	"uniwish.com/internal/api/repository"
	"uniwish.com/internal/scrapers"
)

var fakeId uuid.UUID = uuid.New()

type FakeRepo struct {
	id uuid.UUID
}

func (r *FakeRepo) Insert(_ context.Context, _ string) (uuid.UUID, error) {
	return fakeId, nil
}
func (r *FakeRepo) Dequeue(_ context.Context) (*repository.ScrapeRequest, error) {
	return nil, nil
}

func (r *FakeRepo) MarkDone(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (r *FakeRepo) MarkFailed(_ context.Context, _ uuid.UUID) error {
	return nil
}

var FakeRegistry = scrapers.NewScraperRegistry(map[string]scrapers.ScraperFactory{
	"store.com": func() scrapers.Scraper {
		return nil
	},
},
)

func TestScrapeRequestService(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedError  error
		expectedResult uuid.UUID
	}{
		{
			name:           "no_url",
			url:            "",
			expectedError:  apiErrors.ErrInputInvalid,
			expectedResult: uuid.Nil,
		},
		{
			name:           "parse_error",
			url:            "http://[::1",
			expectedError:  apiErrors.ErrInputInvalid,
			expectedResult: uuid.Nil,
		},
		{
			name:           "no_scheme",
			url:            "example.com",
			expectedError:  apiErrors.ErrInputInvalid,
			expectedResult: uuid.Nil,
		},
		{
			name:           "no_host",
			url:            "http://",
			expectedError:  apiErrors.ErrInputInvalid,
			expectedResult: uuid.Nil,
		},
		{
			name:           "unsupported_host",
			url:            "http://whatever.com",
			expectedError:  apiErrors.ErrStoreUnsupported,
			expectedResult: uuid.Nil,
		},
		{
			name:           "healthy",
			url:            "http://store.com",
			expectedError:  nil,
			expectedResult: fakeId,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			srv := NewScrapeRequestService(&FakeRepo{}, FakeRegistry)
			r, e := srv.Request(context.Background(), tt.url)

			if !errors.Is(e, tt.expectedError) {
				t.Fatalf("expected error %v, received %v", tt.expectedError, e)
			}

			if r != tt.expectedResult {
				t.Fatalf("exepected result %s, received %s", tt.expectedResult, r)
			}
		})
	}
}
