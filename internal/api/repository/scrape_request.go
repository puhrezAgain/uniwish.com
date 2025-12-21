/*
uniwish.com/interal/api/repository/scrape_request

centralizes DB operations with scrape requests
*/
package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type DB interface {
	// allows us to monkeypatch DB connection
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type Transaction interface {
	// allows us to monkeypatch transactions
	Rollback() error
	Commit() error
}

type ScrapeRequest struct {
	ID     uuid.UUID
	Status string
	URL    string
}

type ScrapeRequestRepository interface {
	Insert(ctx context.Context, url string) (uuid.UUID, error)
	Dequeue(ctx context.Context) (*ScrapeRequest, error)
	MarkDone(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID) error
}

type PostgresScrapeRequestRepository struct {
	db DB
}

func NewPostgresScrapeRequestRepository(db DB) ScrapeRequestRepository {
	return &PostgresScrapeRequestRepository{db: db}
}

func (r *PostgresScrapeRequestRepository) Insert(ctx context.Context, url string) (uuid.UUID, error) {
	id := uuid.New()
	_, err := r.db.ExecContext(
		ctx,
		`
		INSERT INTO scrape_requests (id, url, status)
		VALUES ($1, $2, 'pending')
		`,
		id, url,
	)

	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (r *PostgresScrapeRequestRepository) Dequeue(ctx context.Context) (*ScrapeRequest, error) {
	// To prevent race conditions between read and update, dequeue must be called with a transaction backed repo.

	scrapeRequest := ScrapeRequest{}

	err := r.db.QueryRowContext(
		ctx,
		`
		SELECT id, status, url
		FROM scrape_requests
		WHERE status = 'pending'
		ORDER BY created_at
		FOR UPDATE SKIP LOCKED
		LIMIT 1
		`,
	).Scan(&scrapeRequest.ID, &scrapeRequest.Status, &scrapeRequest.URL)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	} else if err == sql.ErrNoRows {
		// no work to do, lets return
		return nil, nil
	}

	_, err = r.db.ExecContext(
		ctx,
		`
		UPDATE scrape_requests
		SET status = 'processing'
		WHERE id = $1
		`, scrapeRequest.ID,
	)

	if err != nil {
		return nil, err
	}
	scrapeRequest.Status = "processing"
	return &scrapeRequest, nil
}

func (r *PostgresScrapeRequestRepository) MarkDone(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(
		ctx,
		`
		UPDATE scrape_requests
		SET status = 'done'
		WHERE id = $1
		`, id,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresScrapeRequestRepository) MarkFailed(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(
		ctx,
		`
		UPDATE scrape_requests
		SET status = 'failed'
		WHERE id = $1
		`, id,
	)

	if err != nil {
		return err
	}

	return nil
}
