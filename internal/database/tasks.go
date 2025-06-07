package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"taskqueue/internal/models"
)

// CreateTask inserts a new task row and returns the filled task.
func CreateTask(ctx context.Context, db *pgxpool.Pool, t *models.Task) error {
	query := `INSERT INTO tasks (user_id, name, type, priority, status, payload)
              VALUES ($1,$2,$3,$4,$5,$6)
              RETURNING id, created_at, updated_at`
	return db.QueryRow(ctx, query,
		t.UserID, t.Name, t.Type, t.Priority, t.Status, t.Payload,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

// ListTasks returns up to limit tasks for a user ordered by creation time.
func ListTasks(ctx context.Context, db *pgxpool.Pool, userID int64, limit int) ([]models.Task, error) {
	rows, err := db.Query(ctx, `SELECT id, user_id, name, type, priority, status, payload,
        result, error_message, message_id, worker_id, started_at, completed_at,
        created_at, updated_at FROM tasks WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2`,
		userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := []models.Task{}
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Type, &t.Priority, &t.Status,
			&t.Payload, &t.Result, &t.Error, &t.MessageID, &t.WorkerID,
			&t.StartedAt, &t.CompletedAt, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}
