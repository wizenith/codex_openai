package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InsertTask stores a new task in the database and returns its ID.
func InsertTask(ctx context.Context, db *pgxpool.Pool, t *Task) (int64, error) {
	row := db.QueryRow(ctx, `
        INSERT INTO tasks (name, payload, status, message_id)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `, t.Name, t.Payload, t.Status, t.MessageID)
	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// ListTasks returns all tasks in the database ordered by creation time descending.
func ListTasks(ctx context.Context, db *pgxpool.Pool) ([]Task, error) {
	rows, err := db.Query(ctx, `
        SELECT id, name, payload, status, message_id, created_at, updated_at
        FROM tasks ORDER BY created_at DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Name, &t.Payload, &t.Status, &t.MessageID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// GetTask retrieves a task by ID.
func GetTask(ctx context.Context, db *pgxpool.Pool, id int64) (*Task, error) {
	row := db.QueryRow(ctx, `
        SELECT id, name, payload, status, message_id, created_at, updated_at
        FROM tasks WHERE id=$1
    `, id)
	var t Task
	if err := row.Scan(&t.ID, &t.Name, &t.Payload, &t.Status, &t.MessageID, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, err
	}
	return &t, nil
}
