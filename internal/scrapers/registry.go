/*
uniwish.com/internal/scrapers/registry

contains logic related to http interaction with zara pages
*/

package scrapers

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"uniwish.com/internal/scrapers/zara"
)

var ErrNoScraper = errors.New("no scraper available")
var ErrInvalidURL = errors.New("invalid url")

type Registry interface {
	AssertSupports(string) error
	NewScraperFor(string) (Scraper, error)
}

type ScraperFactory func() Scraper
type ScraperRegistry struct {
	byHost map[string]ScraperFactory
}

func (sr *ScraperRegistry) AssertSupports(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ErrInvalidURL
	}

	for h := range sr.byHost {
		if strings.Contains(rawURL, h) {
			return nil
		}
	}
	return ErrNoScraper
}

func (sr *ScraperRegistry) NewScraperFor(rawURL string) (Scraper, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, ErrInvalidURL
	}

	for scraperHost, scraperFactort := range sr.byHost {
		if strings.Contains(parsed.Host, scraperHost) {
			return scraperFactort(), nil
		}
	}
	return nil, ErrNoScraper
}

func NewScraperRegistry(byHostMap map[string]ScraperFactory) *ScraperRegistry {
	return &ScraperRegistry{byHost: byHostMap}
}

var DefaultScraperRegistry = NewScraperRegistry(map[string]ScraperFactory{
	"zara.com": func() Scraper {
		return zara.NewZaraScraper(10 * time.Second)
	},
},
)
