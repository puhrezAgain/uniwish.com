/*
uniwish.com/internal/scrapers/registry

centralizes scraper declaration and instantiation
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
	ValidateUrl(string) error
	NewScraperFor(string) (Scraper, error)
}

type ScraperFactory func() Scraper
type ScraperRegistry struct {
	byHost map[string]ScraperFactory
}

func (sr *ScraperRegistry) getHost(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", ErrInvalidURL
	}
	host := parsed.Hostname()
	for hostKey := range sr.byHost {
		if host == hostKey || strings.HasSuffix(host, "."+hostKey) {
			return hostKey, nil
		}
	}
	return "", ErrNoScraper
}

func (sr *ScraperRegistry) ValidateUrl(rawURL string) error {
	_, err := sr.getHost(rawURL)

	if err != nil {
		return err
	}

	return nil
}

func (sr *ScraperRegistry) NewScraperFor(rawURL string) (Scraper, error) {
	host, err := sr.getHost(rawURL)

	if err != nil {
		return nil, err
	}

	factory, ok := sr.byHost[host]
	if !ok {
		return nil, ErrNoScraper
	}
	return factory(), nil

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
