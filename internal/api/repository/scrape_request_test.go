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
	"uniwish.com/internal/testutil"
)

func TestScrapeRequestRepo_Insert(t *testing.T) {
	testutil.RequireIntegration(t)
	tx, err := testDB.BeginTx(context.Background(), nil)

	if err != nil {
		t.Fatalf("transaction failed %v", err)
	}
	defer tx.Rollback()

	repo := NewPostgresScrapeRequestRepository(tx)
	id, err := repo.Insert(context.Background(), "whatever")

	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	if id == uuid.Nil {
		t.Fatal("expected non-nil id")
	}
}

func TestScrapeRequestRepo_PersistsFields(t *testing.T) {
	testutil.RequireIntegration(t)
	tx, err := testDB.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("transaction failed %v", err)
	}
	defer tx.Rollback()

	repo := NewPostgresScrapeRequestRepository(tx)
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

	err = tx.QueryRow(`
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

func TestScrapeRequestRepo_DequeueEmpty(t *testing.T) {
	testutil.RequireIntegration(t)
	tx, err := testDB.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("transaction failed %v", err)
	}
	defer tx.Rollback()

	repo := NewPostgresScrapeRequestRepository(tx)
	job, err := repo.Dequeue(context.Background())

	if err != nil {
		t.Fatalf("expected err to be nil, received %v", err)
	}

	if job != nil {
		t.Fatalf("expected job to be nil, received %v", job)
	}
}

func TestScrapeRequestRepo_DequeueProcesses(t *testing.T) {
	testutil.RequireIntegration(t)
	tx, err := testDB.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("transaction failed %v", err)
	}
	defer tx.Rollback()

	repo := NewPostgresScrapeRequestRepository(tx)
	expectedUrl := "whatever"
	ctx := context.Background()

	_, err = repo.Insert(ctx, expectedUrl)

	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	job, err := repo.Dequeue(context.Background())
	if err != nil {
		t.Fatalf("expected err to be nil, received %v", err)
	}

	if job == nil {
		t.Fatalf("expected job to be nonnil")
	}

	if job.ID == uuid.Nil {
		t.Fatalf("expected job.id to be nonnil")
	}

	if job.Status != "processing" {
		t.Fatalf("expected job status to be processing, received %s", job.Status)
	}
}

func TestScrapeRequestRepo_MarkDone(t *testing.T) {
	testutil.RequireIntegration(t)
	tx, err := testDB.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("transaction failed %v", err)
	}
	defer tx.Rollback()

	repo := NewPostgresScrapeRequestRepository(tx)
	expectedUrl := "whatever"
	ctx := context.Background()

	_, err = repo.Insert(ctx, expectedUrl)

	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	job, err := repo.Dequeue(context.Background())
	if err != nil {
		t.Fatalf("dequeue failed, %v", err)
	}
	err = repo.MarkDone(ctx, job.ID)
	if err != nil {
		t.Fatalf("mark done failed, %v", err)
	}
	var status string

	err = tx.QueryRow(`
	SELECT status
	FROM scrape_requests 
	WHERE id = $1
	`, job.ID).Scan(&status)

	if status != "done" {
		t.Fatalf("expected status to be 'done', received, %s", status)
	}
}

func TestScrapeRequestRepo_MarkFailed(t *testing.T) {
	testutil.RequireIntegration(t)
	tx, err := testDB.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("transaction failed %v", err)
	}
	defer tx.Rollback()

	repo := NewPostgresScrapeRequestRepository(tx)
	expectedUrl := "whatever"
	ctx := context.Background()

	_, err = repo.Insert(ctx, expectedUrl)

	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	job, err := repo.Dequeue(context.Background())
	if err != nil {
		t.Fatalf("dequeue failed, %v", err)
	}
	err = repo.MarkFailed(ctx, job.ID)
	if err != nil {
		t.Fatalf("mark failed failed, %v", err)
	}
	var status string

	err = tx.QueryRow(`
	SELECT status
	FROM scrape_requests 
	WHERE id = $1
	`, job.ID).Scan(&status)

	if status != "failed" {
		t.Fatalf("expected status to be 'failed', received, %s", status)
	}
}
