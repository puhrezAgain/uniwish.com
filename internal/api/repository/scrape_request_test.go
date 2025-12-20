/*
uniwish.com/interal/api/repository/scrape_request_test

testing for scrape request
*/
package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestScrapeRequestRepo_Insert(t *testing.T) {
	requireIntegration(t)

	repo := NewPostgresScrapeRequestRepository(testDB)

	id, err := repo.Insert(context.Background(), "whatever")

	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	if id == uuid.Nil {
		t.Fatal("expected non-nil id")
	}
}

func TestScrapeRequestRepo_PersistsFields(t *testing.T) {
	requireIntegration(t)

	repo := NewPostgresScrapeRequestRepository(testDB)
	expectedUrl := "whatever"
	ctx := context.Background()

	id, err := repo.Insert(ctx, expectedUrl)

	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	var (
		storedUrl  string
		status     string
		created_at time.Time
	)

	err = testDB.QueryRow(`
	SELECT url, status, created_at
	FROM scrape_requests 
	WHERE id = $1
	`, id).Scan(&storedUrl, &status, &created_at)

	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if storedUrl != expectedUrl {
		t.Fatalf("expected url %s, got %s", expectedUrl, storedUrl)
	}

	if status != "pending" {
		t.Fatalf("expected default status should be 'pending', received %s", status)
	}

	if created_at.IsZero() {
		t.Fatalf("expected created_at to be set")
	}
}
