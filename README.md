# Task Queue System

This repository contains an experimental implementation of a task queue system written in Go. It demonstrates a REST API using Gin, PostgreSQL access via pgx, and a simple integration with AWS SQS for message queuing.

The project layout follows a typical Go application structure with code organized under the `internal/` directory. The `cmd/server` entrypoint starts the HTTP server and exposes minimal endpoints for enqueuing tasks and health checks.

This is a work in progress and only a subset of the planned features are implemented.

## Quick start

```sh
PORT=8080 DATABASE_URL=postgres://user:pass@localhost:5432/db \
AWS_SQS_QUEUE_URL=https://sqs.region.amazonaws.com/account/queue \
go run ./cmd/server
```

## Legacy Calculator

The repository previously held a simple Reverse Polish Notation calculator which remains in `main.go` for reference.
