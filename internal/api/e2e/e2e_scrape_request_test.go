/*
uniwish.com/interal/e2e/e2e_scrape_request_test

End to end testing for our scrape request endpoint path
*/
package e2e

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func requireE2E(t *testing.T) {
	if os.Getenv("E2E_TESTS") == "" {
		t.Skip("e2e tests disabled")
	} else if os.Getenv("API_BASE_URL") == "" {
		t.Fatal("API_BASE required for e2e")
	} else if os.Getenv("DATABASE_URL") == "" {
		t.Fatal("DATABASE_URL required for e2e")
	}
}

func TestCreateScrapeRequest_EndToEnd(t *testing.T) {
	requireE2E(t)
	storeUrl := "http://store.com/product/123"

	req, err := http.NewRequest(
		http.MethodPost,
		os.Getenv("API_BASE_URL")+"/scrape-requests",
		bytes.NewBufferString(fmt.Sprintf(`{"url": "%s"}`, storeUrl)),
	)

	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Do(req)

	if err != nil {
		t.Fatalf("http request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected status 202, received %d", resp.StatusCode)
	}

	var response struct {
		ID     uuid.UUID `json:"id"`
		Status string    `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}

	if response.ID == uuid.Nil {
		t.Fatal("expected non-nil id")
	}

	if response.Status != "pending" {
		t.Fatalf("expected status 'pending', received %s", response.Status)
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))

	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var (
		storedUrl    string
		storedStatus string
	)

	err = db.QueryRowContext(ctx, `
	SELECT url, status
	FROM scrape_requests
	WHERE id = $1
	`, response.ID).Scan(&storedUrl, &storedStatus)

	if err != nil {
		t.Fatalf("db row not found, %v", err)
	}

	if storedUrl != storeUrl {
		t.Fatalf("expected url '%s', received %s", storeUrl, storedUrl)
	}

	if storedStatus != "pending" {
		t.Fatalf("expected status 'pending', received %s", storedStatus)
	}
}

func TestCreateScrapeRequest_InvalidJSON(t *testing.T) {
	requireE2E(t)

	baseURL := os.Getenv("API_BASE_URL")
	req, err := http.NewRequest(
		http.MethodPost,
		baseURL+"/scrape-requests",
		bytes.NewBufferString(`{"url":`),
	)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateScrapeRequest_StoreUnsupported(t *testing.T) {
	requireE2E(t)

	baseURL := os.Getenv("API_BASE_URL")
	req, err := http.NewRequest(
		http.MethodPost,
		baseURL+"/scrape-requests",
		bytes.NewBufferString(`{"url": "http://unsupportedstore.com"}`),
	)

	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}
}
