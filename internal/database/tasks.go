package database

import (
	"context"
	"database/sql"
	"fmt"

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

// ListTasks returns tasks for a user with pagination and filtering
func ListTasks(ctx context.Context, db *pgxpool.Pool, userID int64, filter *TaskFilter) ([]models.Task, error) {
	query := `SELECT id, user_id, name, type, priority, status, payload,
        result, error_message, message_id, worker_id, started_at, completed_at,
        created_at, updated_at FROM tasks WHERE user_id=$1`
	
	args := []interface{}{userID}
	argIndex := 2
	
	if filter.Status != "" {
		query += fmt.Sprintf(" AND status=$%d", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}
	
	if filter.Type != "" {
		query += fmt.Sprintf(" AND type=$%d", argIndex)
		args = append(args, filter.Type)
		argIndex++
	}
	
	if filter.Priority != "" {
		query += fmt.Sprintf(" AND priority=$%d", argIndex)
		args = append(args, filter.Priority)
		argIndex++
	}
	
	query += " ORDER BY created_at DESC"
	
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}
	
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
	}
	
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	tasks := []models.Task{}
	for rows.Next() {
		var t models.Task
		var startedAt, completedAt sql.NullTime
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Type, &t.Priority, &t.Status,
			&t.Payload, &t.Result, &t.Error, &t.MessageID, &t.WorkerID,
			&startedAt, &completedAt, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		if startedAt.Valid {
			t.StartedAt = startedAt.Time
		}
		if completedAt.Valid {
			t.CompletedAt = completedAt.Time
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// GetTask returns a single task by ID
func GetTask(ctx context.Context, db *pgxpool.Pool, taskID, userID int64) (*models.Task, error) {
	var t models.Task
	var startedAt, completedAt sql.NullTime
	
	err := db.QueryRow(ctx, `
		SELECT id, user_id, name, type, priority, status, payload,
		result, error_message, message_id, worker_id, started_at, completed_at,
		created_at, updated_at FROM tasks WHERE id=$1 AND user_id=$2`,
		taskID, userID).Scan(
		&t.ID, &t.UserID, &t.Name, &t.Type, &t.Priority, &t.Status,
		&t.Payload, &t.Result, &t.Error, &t.MessageID, &t.WorkerID,
		&startedAt, &completedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	if startedAt.Valid {
		t.StartedAt = startedAt.Time
	}
	if completedAt.Valid {
		t.CompletedAt = completedAt.Time
	}
	
	return &t, nil
}

// UpdateTaskStatus updates a task's status and related fields
func UpdateTaskStatus(ctx context.Context, db *pgxpool.Pool, taskID int64, status string, messageID string) error {
	_, err := db.Exec(ctx, `
		UPDATE tasks SET status=$1, message_id=$2, updated_at=CURRENT_TIMESTAMP
		WHERE id=$3`, status, messageID, taskID)
	return err
}

// UpdateTaskProgress updates a task when worker starts processing
func UpdateTaskProgress(ctx context.Context, db *pgxpool.Pool, messageID, workerID string) error {
	_, err := db.Exec(ctx, `
		UPDATE tasks SET status='processing', worker_id=$1, started_at=CURRENT_TIMESTAMP
		WHERE message_id=$2`, workerID, messageID)
	return err
}

// CompleteTask marks a task as completed with result
func CompleteTask(ctx context.Context, db *pgxpool.Pool, messageID string, result []byte) error {
	_, err := db.Exec(ctx, `
		UPDATE tasks SET status='completed', result=$1, completed_at=CURRENT_TIMESTAMP
		WHERE message_id=$2`, result, messageID)
	return err
}

// FailTask marks a task as failed with error message
func FailTask(ctx context.Context, db *pgxpool.Pool, messageID string, errorMsg string) error {
	_, err := db.Exec(ctx, `
		UPDATE tasks SET status='failed', error_message=$1, completed_at=CURRENT_TIMESTAMP
		WHERE message_id=$2`, errorMsg, messageID)
	return err
}

// CancelTask cancels a pending/queued task
func CancelTask(ctx context.Context, db *pgxpool.Pool, taskID, userID int64) error {
	result, err := db.Exec(ctx, `
		UPDATE tasks SET status='cancelled', completed_at=CURRENT_TIMESTAMP
		WHERE id=$1 AND user_id=$2 AND status IN ('pending', 'queued')`,
		taskID, userID)
	
	if err != nil {
		return err
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("task not found or cannot be cancelled")
	}
	
	return nil
}

// GetTaskStats returns task statistics for a user
func GetTaskStats(ctx context.Context, db *pgxpool.Pool, userID int64) (*TaskStats, error) {
	var stats TaskStats
	
	err := db.QueryRow(ctx, `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'pending') as pending,
			COUNT(*) FILTER (WHERE status = 'queued') as queued,
			COUNT(*) FILTER (WHERE status = 'processing') as processing,
			COUNT(*) FILTER (WHERE status = 'completed') as completed,
			COUNT(*) FILTER (WHERE status = 'failed') as failed,
			COUNT(*) FILTER (WHERE status = 'cancelled') as cancelled
		FROM tasks WHERE user_id=$1`, userID).Scan(
		&stats.Total, &stats.Pending, &stats.Queued, &stats.Processing,
		&stats.Completed, &stats.Failed, &stats.Cancelled,
	)
	
	return &stats, err
}

// TaskFilter contains filtering options for listing tasks
type TaskFilter struct {
	Status   string
	Type     string
	Priority string
	Limit    int
	Offset   int
}

// TaskStats contains task statistics
type TaskStats struct {
	Total      int `json:"total"`
	Pending    int `json:"pending"`
	Queued     int `json:"queued"`
	Processing int `json:"processing"`
	Completed  int `json:"completed"`
	Failed     int `json:"failed"`
	Cancelled  int `json:"cancelled"`
}
