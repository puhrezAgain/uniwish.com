/*
uniwish.com/internal/scrapers/zara/scraper_test

contains test related to scraping zara pages
*/
package zara

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type FakeZaraScraper struct {
	ZaraScraper
}

func TestZaraScraper_LoadFile(t *testing.T) {
	html, _ := os.ReadFile("testdata/zara_product.html")

	if !bytes.Contains(html, []byte("application/ld+json")) {
		t.Fatal("expected zara file to have ld+json tag")
	}
}

func TestZaraScraper_Scrape(t *testing.T) {
	scraper := &FakeZaraScraper{}
	page, _ := os.ReadFile("testdata/zara_product.html")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(page)
	}))
	defer ts.Close()
	scraper.client = ts.Client()

	expectedSku := "469506351-800-32"
	expectedName := "JEANS TRF WIDE LEG TIRO ALTO"
	productRecord, err := scraper.Scrape(context.Background(), ts.URL)

	if err != nil {
		t.Fatalf("expected product, received err: %v", err)
	}

	if productRecord == nil {
		t.Fatal("product is nil")
	}

	if productRecord.Product.URL != ts.URL {
		t.Fatalf("url expected to be %s, received %s", ts.URL, productRecord.Product.URL)
	}

	if productRecord.Product.Name != expectedName {
		t.Fatalf("name expected to be %s, received %s", expectedName, productRecord.Product.Name)
	}

	if productRecord.Product.SKU != expectedSku {
		t.Fatalf("name expected to be %s, received %s", expectedSku, productRecord.Product.SKU)
	}
	if len(*productRecord.Offers) != 24 {
		t.Fatalf("expected json length 24, receieved %d", len(*productRecord.Offers))
	}

	for _, o := range *productRecord.Offers {
		if o.ProductID != productRecord.Product.ID {
			t.Fatalf("product offer's parent %v doesn't patch parent %v", o.ProductID, productRecord.Product.ID)
		}
		if o.Price == 0 {
			t.Fatalf("price nil")
		}
	}
}

func TestZaraScraper_ExtractJLJSON_Success(t *testing.T) {
	scraper := &FakeZaraScraper{}
	page, _ := os.Open("testdata/zara_product.html")
	ldjson, err := scraper.extractProductsFromPage(page)

	if err != nil {
		t.Fatalf("expected json, received err: %v", err)
	}

	if ldjson == nil {
		t.Fatal("json is nil")
	}

	if len(ldjson) != 24 {
		t.Fatalf("expected json length 24, receieved %d", len(ldjson))
	}
}
