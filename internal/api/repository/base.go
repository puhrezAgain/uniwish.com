/*
uniwish.com/interal/api/repository/base

centralizes DB interfaces
*/
package repository

import (
	"context"
	"database/sql"
)

type DB interface {
	// allows us to monkeypatch DB connection
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type TransactionCreator interface {
	DB
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type Transaction interface {
	// allows us to monkeypatch transactions
	Rollback() error
	Commit() error
}
