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

type ScrapeRequestRepository interface {
	Insert(ctx context.Context, url string) (uuid.UUID, error)
}

type PostgresScrapeRequestRepository struct {
	db *sql.DB
}

func NewPostgresScrapeRequestRepository(db *sql.DB) *PostgresScrapeRequestRepository {
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
