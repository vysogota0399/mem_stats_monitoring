package interfaces

import (
	"context"
	"database/sql"
)

type IDB interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	Ping() error
	Close() error
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}
