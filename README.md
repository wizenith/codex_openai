# Task Queue System

This repository contains an experimental implementation of a task queue system written in Go. It demonstrates a REST API using Gin, PostgreSQL access via pgx, and a simple integration with AWS SQS for message queuing.

The project layout follows a typical Go application structure with code organized under the `internal/` directory. The `cmd/server` entrypoint starts the HTTP server and exposes endpoints for creating and listing tasks as well as a simple health check.

This is a work in progress and only a subset of the planned features are implemented.

## Quick start

```sh
cp .env.example .env
go run ./cmd/server
```

Then open `http://localhost:8080/` to view the dashboard UI.

## Legacy Calculator

The repository previously held a simple Reverse Polish Notation calculator which remains in `main.go` for reference.
