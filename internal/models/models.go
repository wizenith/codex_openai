package models

import "time"

// User represents an authenticated user.
type User struct {
	ID        int64     `db:"id"`
	GoogleID  string    `db:"google_id"`
	Email     string    `db:"email"`
	Name      string    `db:"name"`
	Picture   string    `db:"picture"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Task represents a queued task.
type Task struct {
	ID          int64     `db:"id"`
	UserID      int64     `db:"user_id"`
	Name        string    `db:"name"`
	Type        string    `db:"type"`
	Priority    string    `db:"priority"`
	Status      string    `db:"status"`
	Payload     []byte    `db:"payload"`
	Result      []byte    `db:"result"`
	Error       string    `db:"error_message"`
	MessageID   string    `db:"message_id"`
	WorkerID    string    `db:"worker_id"`
	StartedAt   time.Time `db:"started_at"`
	CompletedAt time.Time `db:"completed_at"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
