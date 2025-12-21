/*
uniwish.com/interal/api/services/scraper

contains logic related to scrapers and scraping
*/
package services

import "context"

type Scraper interface {
	Scrape(ctx context.Context, url string) (any, error) // TODO: change any to product type
}

type ScraperRegistry interface {
	Get(host string) (Scraper, bool)
}
