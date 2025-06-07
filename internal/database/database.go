package database

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect creates a new pgx connection pool.
func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, dsn)
}
