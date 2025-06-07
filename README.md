# Task Queue System

A comprehensive web application for managing task queues with AWS SQS integration, real-time updates, and multi-language worker support.

## Features

- **Google OAuth Authentication**: Secure user authentication with Google OAuth2
- **Task Management**: Create, list, filter, and cancel tasks with priority levels
- **Real-time Updates**: WebSocket integration for live task status updates
- **Multi-language Workers**: Python and Node.js worker examples included
- **Priority Queue**: Support for high, medium, and low priority tasks
- **HTMX Frontend**: Dynamic UI without complex JavaScript frameworks
- **Docker Support**: Fully containerized deployment

## Technology Stack

- **Backend**: Go with Gin framework
- **Database**: PostgreSQL
- **Frontend**: HTMX for dynamic interactions
- **Message Queue**: AWS SQS
- **Authentication**: Google OAuth2 with JWT
- **Real-time**: WebSocket
- **Workers**: Python & Node.js examples
- **Deployment**: Docker & Docker Compose

## Quick Start

### Prerequisites

- Docker and Docker Compose
- AWS Account with SQS configured
- Google OAuth2 credentials

### Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd webapp_htmx
```

2. Copy environment variables:
```bash
cp .env.example .env
```

3. Edit `.env` file with your credentials:
- Google OAuth credentials
- AWS credentials and SQS queue URL
- JWT secret key

4. Start the application:
```bash
docker-compose up
```

The application will be available at `http://localhost:8080`

## Configuration

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `PORT` | Server port (default: 8080) | No |
| `DATABASE_URL` | PostgreSQL connection string | Yes |
| `JWT_SECRET` | Secret key for JWT tokens | Yes |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID | Yes |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret | Yes |
| `GOOGLE_REDIRECT_URL` | OAuth callback URL | Yes |
| `AWS_REGION` | AWS region for SQS | Yes |
| `AWS_ACCESS_KEY_ID` | AWS access key | Yes |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key | Yes |
| `AWS_SQS_QUEUE_URL` | SQS queue URL | Yes |

## Architecture

### Components

1. **Web Server** (Go/Gin)
   - Handles HTTP requests
   - Manages authentication
   - Serves HTMX frontend
   - WebSocket connections

2. **Database** (PostgreSQL)
   - Stores users and tasks
   - Automatic migrations on startup

3. **Message Queue** (AWS SQS)
   - Distributes tasks to workers
   - Supports priority-based processing

4. **Workers** (Python/Node.js)
   - Poll SQS for tasks
   - Process tasks based on type
   - Update task status in database

### Task Types

- **Email**: Email processing simulation
- **Data**: Data manipulation operations
- **File**: File operation simulation
- **API**: External API calls
- **Script**: Script execution simulation
- **Report**: Report generation simulation

## API Endpoints

### Authentication
- `GET /auth/google` - Initiate Google OAuth
- `GET /auth/google/callback` - OAuth callback
- `GET /logout` - Logout user
- `GET /api/user` - Get current user info

### Tasks
- `POST /api/tasks` - Create new task
- `GET /api/tasks` - List tasks (with filters)
- `GET /api/tasks/:id` - Get task details
- `DELETE /api/tasks/:id` - Cancel task
- `GET /api/tasks/stats` - Get task statistics

### WebSocket
- `GET /ws` - WebSocket connection for real-time updates

## Development

### Local Development

1. Install Go 1.21+
2. Install dependencies:
```bash
go mod download
```

3. Run migrations:
```bash
# Set DATABASE_URL environment variable
go run cmd/server/main.go
```

4. Start server:
```bash
go run cmd/server/main.go
```

### Adding New Task Types

1. Add handler in worker files:
   - `workers/python/worker.py`
   - `workers/nodejs/worker.js`

2. Update task type in frontend:
   - `web/templates/dashboard.html`

### Running Tests

```bash
go test ./...
```

## Deployment

### Production Considerations

1. **Security**:
   - Use strong JWT secret
   - Enable HTTPS
   - Restrict CORS origins
   - Use environment-specific OAuth redirect URLs

2. **Database**:
   - Use managed PostgreSQL service
   - Enable SSL connections
   - Regular backups

3. **SQS**:
   - Use appropriate visibility timeout
   - Configure dead letter queues
   - Monitor queue metrics

4. **Workers**:
   - Scale based on queue depth
   - Implement health checks
   - Use container orchestration (K8s, ECS)

### Docker Deployment

Build and run with Docker Compose:
```bash
docker-compose up --build
```

Scale workers:
```bash
docker-compose up --scale python-worker=5 --scale nodejs-worker=5
```

## Monitoring

- Health check endpoint: `GET /healthz`
- Task statistics: `GET /api/tasks/stats`
- SQS metrics via AWS CloudWatch
- Application logs in container stdout

## Troubleshooting

### Common Issues

1. **Database connection failed**:
   - Check DATABASE_URL format
   - Ensure PostgreSQL is running
   - Verify network connectivity

2. **OAuth not working**:
   - Verify redirect URL matches Google console
   - Check client ID and secret
   - Ensure cookies are enabled

3. **Workers not processing tasks**:
   - Check AWS credentials
   - Verify SQS queue exists
   - Check worker logs for errors

## Project Structure

```
/
├── cmd/server/main.go           # Application entry point
├── internal/
│   ├── auth/                    # OAuth & JWT handling
│   ├── config/                  # Configuration management
│   ├── database/                # DB connection & migrations
│   ├── handlers/                # HTTP handlers
│   ├── middleware/              # Auth, CORS, logging middleware
│   ├── models/                  # Data models
│   ├── queue/                   # SQS integration
│   └── websocket/               # WebSocket hub & clients
├── pkg/
│   └── logger/                  # Logging utilities
├── web/
│   ├── templates/               # HTML templates
│   └── static/                  # CSS, JS assets
├── workers/
│   ├── python/                  # Python worker
│   └── nodejs/                  # Node.js worker
├── docker-compose.yml           # Docker composition
└── README.md                    # This file
```

## Contributing

1. Fork the repository
2. Create feature branch
3. Commit changes
4. Push to branch
5. Create Pull Request

## License

This project is licensed under the MIT License.