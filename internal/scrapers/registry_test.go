/*
uniwish.com/internal/scrapers/registry_test

contains logic related to http interaction with zara pages
*/
package scrapers

import (
	"errors"
	"testing"
)

func TestRegistry_ValidateUrl(t *testing.T) {
	registry := NewScraperRegistry(map[string]ScraperFactory{
		"zara.com": func() Scraper { return nil },
	})

	tests := []struct {
		url         string
		expectedErr error
	}{
		{"https://zara.com/item", nil},
		{"https://www.zara.com/item", nil},
		{"https://evilzara.com", ErrNoScraper},
		{"https://zara.somethingelse.com", ErrNoScraper},
		{"notaurl", ErrInvalidURL},
	}

	for _, tt := range tests {
		resultErr := registry.ValidateUrl(tt.url)
		if !errors.Is(resultErr, tt.expectedErr) {
			t.Fatalf("url=%s expected %v, received: %v", tt.url, tt.expectedErr, resultErr)
		}
	}
}
