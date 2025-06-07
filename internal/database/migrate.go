package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Migrate creates required database tables if they do not exist.
func Migrate(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
CREATE TABLE IF NOT EXISTS tasks (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    message_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`)
	return err
}
