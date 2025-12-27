/*
uniwish.com/internal/api/services/scraper

contains logic related to scrapers and scraping
*/
package services

import (
	"net/url"
	"strings"
	"time"

	"uniwish.com/internal/api/errors"
	"uniwish.com/internal/scrapers"
	"uniwish.com/internal/scrapers/zara"
)

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
