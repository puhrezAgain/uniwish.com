/*
uniwish.com/interal/worker/base

centralized scraper worker base structs and interfacees
*/
package worker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"uniwish.com/internal/api/repository"
)

var ErrNoWork = errors.New("no job available")

type WorkerRepo interface {
	BeginSession(context.Context) (WorkerSession, error)
}
type DefaultWorkerRepo struct {
	db repository.TransactionCreator
}

func (wr *DefaultWorkerRepo) BeginSession(ctx context.Context) (WorkerSession, error) {
	tx, err := wr.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &DefaultWorkerSession{
		repository.NewPostgresScrapeRequestRepository(wr.db),
		repository.NewProductRepository(wr.db),
		tx,
	}, nil

}

type WorkerSession interface {
	repository.ScrapeRequestRepository
	repository.ProductRepository
	repository.Transaction
}

type DefaultWorkerSession struct {
	repository.ScrapeRequestRepository
	repository.ProductRepository
	*sql.Tx
}

func NewWorkerRepo(db repository.TransactionCreator) DefaultWorkerRepo {
	return DefaultWorkerRepo{db}
}

type JobError struct {
	// used to represent errors that shouldn't be considered worker critical
	// It indicates the worker is healthy and should continue running.
	JobID uuid.UUID
	Err   error
}

func (e JobError) Error() string {
	return fmt.Sprintf("job %s: %v", e.JobID, e.Err)
}

func (e JobError) Unwrap() error {
	return e.Err
}
