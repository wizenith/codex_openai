package database

import (
	"context"
	"embed"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Connect creates a new pgx connection pool.
func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := runMigrations(ctx, pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return pool, nil
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	files, err := migrations.ReadDir("migrations")
	if err != nil {
		return err
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		content, err := migrations.ReadFile("migrations/" + file.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", file.Name(), err)
		}

		if _, err := pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file.Name(), err)
		}
	}

	return nil
}

// User database operations
func CreateUser(ctx context.Context, db *pgxpool.Pool, googleID, email, name, picture string) (int64, error) {
	var userID int64
	err := db.QueryRow(ctx, `
		INSERT INTO users (google_id, email, name, picture)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (google_id) DO UPDATE
		SET email = EXCLUDED.email,
		    name = EXCLUDED.name,
		    picture = EXCLUDED.picture,
		    updated_at = CURRENT_TIMESTAMP
		RETURNING id
	`, googleID, email, name, picture).Scan(&userID)
	return userID, err
}

func GetUserByGoogleID(ctx context.Context, db *pgxpool.Pool, googleID string) (int64, error) {
	var userID int64
	err := db.QueryRow(ctx, "SELECT id FROM users WHERE google_id = $1", googleID).Scan(&userID)
	if err == pgx.ErrNoRows {
		return 0, nil
	}
	return userID, err
}
