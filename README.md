# Task Queue System

This repository contains an experimental implementation of a task queue system written in Go. It demonstrates a REST API using Gin, PostgreSQL access via pgx, and a simple integration with AWS SQS for message queuing.

The project layout follows a typical Go application structure with code organized under the `internal/` directory. The `cmd/server` entrypoint starts the HTTP server and exposes endpoints for creating and retrieving tasks along with a health check. The health endpoint verifies database and SQS connectivity.

This is a work in progress and only a subset of the planned features are implemented.

## Quick start

```sh
PORT=8080 DATABASE_URL=postgres://user:pass@localhost:5432/db \
AWS_SQS_QUEUE_URL=https://sqs.region.amazonaws.com/account/queue \
go run ./cmd/server
```

The server exposes the following endpoints (API version 1):

- `POST /api/v1/tasks` – enqueue a new task (parameters: `name`, `payload`)
- `GET /api/v1/tasks` – list all tasks
- `GET /api/v1/tasks/:id` – retrieve a single task by ID

Database tables are created automatically on startup.

## Legacy Calculator

The repository previously held a simple Reverse Polish Notation calculator which remains in `main.go` for reference.
