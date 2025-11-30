# Message Scheduler Microservice

A production-ready Go microservice template for message scheduling and delivery, built with Fiber framework. Features webhook integration, Redis caching, SQLite persistence, and comprehensive testing.

## Features

- 🚀 **Fast HTTP Framework**: Built with [Fiber v2](https://gofiber.io/)
- 📝 **API Documentation**: Swagger/OpenAPI integration
- ✅ **Request Validation**: Built-in validation middleware
- ⚙️ **Configuration Management**: Environment-based configuration
- 🐳 **Docker Support**: Containerized deployment with multi-stage builds
- 🧪 **Testing**: Comprehensive test suite with Ginkgo/Gomega
- 📊 **Logging**: Structured logging with Logrus
- 🔄 **Message Scheduler**: Background job processing for message delivery
- 💾 **Dual Storage**: SQLite for persistence + Redis for caching
- 🔌 **Webhook Integration**: External message delivery via webhooks

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Go Template Microservice                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────────────────────┐ │
│  │   REST API   │────▶│   Handlers   │────▶│        Services              │ │
│  │  (Fiber v2)  │     │              │     │                              │ │
│  └──────────────┘     └──────────────┘     │  ┌────────────────────────┐  │ │
│                                            │  │   MessageService       │  │ │
│                                            │  │   - Start/Stop Sched.  │  │ │
│                                            │  │   - List Sent Messages │  │ │
│                                            │  └───────────┬────────────┘  │ │
│                                            │              │               │ │
│                                            │  ┌───────────▼────────────┐  │ │
│                                            │  │   MessageScheduler     │  │ │
│                                            │  │   - Batch Processing   │  │ │
│                                            │  │   - Interval Ticking   │  │ │
│                                            │  └───────────┬────────────┘  │ │
│                                            │              │               │ │
│                                            │  ┌───────────▼────────────┐  │ │
│                                            │  │   MessageSender        │  │ │
│                                            │  │   - Webhook Delivery   │  │ │
│                                            │  └───────────┬────────────┘  │ │
│                                            └──────────────┼───────────────┘ │
│                                                           │                 │
│  ┌────────────────────────────────────────────────────────┼───────────────┐ │
│  │                        Repository Layer                │               │ │
│  │  ┌─────────────────────┐    ┌──────────────────────────▼─────────────┐ │ │
│  │  │ MessageRepository   │    │     MessageCacheRepository             │ │ │
│  │  │ - CRUD Operations   │    │     - Cache Sent Messages              │ │ │
│  │  │ - Status Updates    │    │     - TTL Management                   │ │ │
│  │  └──────────┬──────────┘    └──────────────────────────┬─────────────┘ │ │
│  └─────────────┼──────────────────────────────────────────┼───────────────┘ │
│                │                                          │                 │
│       ┌────────▼────────┐                        ┌────────▼────────┐        │
│       │     SQLite      │                        │      Redis      │        │
│       │   (Persistent)  │                        │    (Cache)      │        │
│       └─────────────────┘                        └─────────────────┘        │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
                        ┌─────────────────────────┐
                        │    External Webhook     │
                        │    (Message Delivery)   │
                        └─────────────────────────┘
```

### Message Flow

1. **Message Creation**: Messages are created with `PENDING` status in SQLite
2. **Scheduler Processing**: Background scheduler picks up pending messages in batches
3. **Webhook Delivery**: Messages are sent to external webhook endpoint
4. **Status Update**: On success, message status is updated to `SENT` with external ID
5. **Caching**: Sent messages are cached in Redis for fast retrieval
6. **Retrieval**: List API fetches from cache first, then falls back to database

### Component Responsibilities

| Component | Responsibility |
|-----------|---------------|
| **Handler** | HTTP request/response handling, validation |
| **Service** | Business logic orchestration |
| **Scheduler** | Background job processing with configurable interval |
| **Sender** | External webhook communication |
| **Repository** | Data access abstraction (SQLite + Redis) |

## Project Structure

```
├── cmd/
│   └── api/                  # Application entrypoint
│       ├── main.go           # Main function
│       └── bootstrap.go      # Dependency injection & app setup
├── internal/
│   ├── config/               # Configuration management
│   ├── constants/            # Application constants
│   ├── handlers/             # HTTP handlers
│   ├── middleware/           # Custom middleware (validation)
│   ├── models/               # Domain models (Message, Cache)
│   ├── repository/           # Data access layer
│   │   └── mocks/            # Repository mocks for testing
│   ├── resources/
│   │   ├── request/          # Request DTOs
│   │   └── response/         # Response DTOs
│   ├── router/               # Route definitions
│   └── services/             # Business logic
│       └── mocks/            # Service mocks for testing
├── pkg/
│   ├── redis/                # Redis client wrapper
│   ├── sqlite/               # SQLite client wrapper
│   ├── utils/                # Utility functions
│   └── validator/            # Validation logic
├── docs/                     # Swagger documentation
├── dev/                      # Development tools (docker-compose)
├── Dockerfile                # Multi-stage Docker build
├── Makefile                  # Build automation
└── .env.example              # Environment configuration template
```

## Getting Started

### Prerequisites

- Go 1.23.9 or higher
- Docker & Docker Compose (for Redis)
- Make (optional, for convenience commands)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd go-template-microservice
```

2. Download dependencies:
```bash
make download-deps
# or
go mod download
```

3. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Start Redis (required for caching):
```bash
cd dev && docker-compose up -d
```

### Running the Application

#### Using Make (recommended):
```bash
make run
```

#### Using Go directly:
```bash
go run cmd/api/*
```

#### Using Docker:
```bash
# Build image
make dockerize

# Run container
make docker-run CONTAINER_NAME=my-service
```

The server will start on `http://localhost:8080` (or configured port).

## API Endpoints

### Health Check

```http
GET /health
```

**Response:**
```json
{
  "status": "success",
  "timestamp": 1732972800000,
  "data": {
    "service": "up"
  }
}
```

### Start Message Scheduler

```http
POST /messages/start
```

Starts the background scheduler that processes pending messages.

**Response:**
```json
{
  "status": "success",
  "timestamp": 1732972800000,
  "data": {
    "state": "started"
  }
}
```

### Stop Message Scheduler

```http
POST /messages/stop
```

Stops the background scheduler gracefully.

**Response:**
```json
{
  "status": "success",
  "timestamp": 1732972800000,
  "data": {
    "state": "stopped"
  }
}
```

### List Sent Messages

```http
GET /messages/sent?limit=10
```

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | int | 10 | Maximum number of messages to retrieve |

**Response:**
```json
{
  "status": "success",
  "timestamp": 1732972800000,
  "data": [
    {
      "message_id": 1,
      "external_message_id": "ext-abc123",
      "to": "+905551234567",
      "content": "Hello World",
      "sent_at": "2025-11-30 12:30:00"
    }
  ]
}
```

## Configuration

All configuration is done via environment variables. See `.env.example` for all available options:

### Server Configuration
| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_HTTP_PORT` | HTTP server port | `8080` |
| `SERVER_ENVIRONMENT` | Environment (local/development/staging/production) | `local` |
| `SERVER_LOG_LEVEL` | Log level (DEBUG/INFO/WARN/ERROR) | `INFO` |
| `SERVER_READ_TIMEOUT` | HTTP read timeout in seconds | `5` |
| `SERVER_WRITE_TIMEOUT` | HTTP write timeout in seconds | `10` |

### HTTP Client Configuration
| Variable | Description | Default |
|----------|-------------|---------|
| `HTTP_CLIENT_TIMEOUT` | HTTP client timeout in seconds | `30` |
| `HTTP_CLIENT_KEEP_ALIVE` | Keep-alive duration in seconds | `30` |
| `HTTP_CLIENT_IDLE_CONN_TIMEOUT` | Idle connection timeout in seconds | `90` |
| `HTTP_CLIENT_MAX_IDLE_CONNS` | Maximum idle connections | `100` |
| `HTTP_CLIENT_MAX_IDLE_CONNS_PER_HOST` | Max idle connections per host | `10` |

### Webhook Configuration
| Variable | Description | Default |
|----------|-------------|---------|
| `WEBHOOK_CONFIG_URL` | External webhook URL for message delivery | `http://localhost:9000/webhook` |
| `WEBHOOK_CONFIG_AUTH_KEY` | Authentication key for webhook | - |

### Scheduler Configuration
| Variable | Description | Default |
|----------|-------------|---------|
| `SCHEDULER_INTERVAL_IN_SECONDS` | Interval between scheduler runs | `120` |
| `SCHEDULER_BATCH_SIZE` | Number of messages to process per batch | `2` |

### Database Configuration
| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_NAME` | SQLite database name | `message` |

### Redis Configuration
| Variable | Description | Default |
|----------|-------------|---------|
| `REDIS_HOST` | Redis server host | `localhost` |
| `REDIS_PORT` | Redis server port | `6379` |
| `REDIS_PASSWORD` | Redis password | - |
| `REDIS_DB` | Redis database number | `0` |
| `REDIS_TTL_IN_SECONDS` | Cache TTL in seconds | `3600` |

## API Documentation (Swagger)

The API documentation is automatically generated using Swagger and available at:

- **Swagger UI**: `http://localhost:8080/documentation/`
- **JSON Spec**: `http://localhost:8080/documentation/document.json`

### Generating/Updating Documentation

```bash
make docs
```

This command uses `swag` to parse Go annotations and generate OpenAPI documentation.

## Development

### Available Make Commands

| Command | Description |
|---------|-------------|
| `make run` | Run the application locally |
| `make download-deps` | Download Go dependencies |
| `make docs` | Generate Swagger documentation |
| `make test` | Run tests with coverage report |
| `make create-mocks` | Generate mocks for testing |
| `make install-mockgen` | Install mockgen tool |
| `make dockerize` | Build Docker image |
| `make docker-run` | Run Docker container |

### Running Tests

```bash
# Run all tests with coverage
make test

# Run specific package tests
go test -v ./internal/repository/...
go test -v ./internal/services/...
```

This generates:
- `coverage.out`: Coverage data file
- `coverage.html`: HTML coverage report (open in browser)

### Test Architecture

The project uses **Ginkgo/Gomega** for BDD-style testing:

- **Repository Tests**: Test data access layer with real SQLite and Redis
- **Service Tests**: Test business logic with mocks and integration tests
- **E2E Tests**: Full flow testing with mock webhook server

### Generating Mocks

```bash
make create-mocks
```

This generates mocks for:
- `IRepository`
- `MessageRepository`
- `MessageCacheRepository`
- `MessageSenderService`

## Docker Deployment

### Building the Image

```bash
docker build -t go-template-microservice .
```

### Running with Docker Compose

For local development with all dependencies:

```bash
# Start Redis
cd dev && docker-compose up -d

# Run the application
docker run -p 8080:8080 --env-file .env --network host go-template-microservice
```

### Production Deployment

The Dockerfile uses multi-stage builds:
1. **Builder stage**: Compiles the Go binary with CGO enabled for SQLite
2. **Runner stage**: Minimal Alpine image with the binary

Features:
- Non-root user for security
- CA certificates for HTTPS calls
- Minimal image size

## Error Handling

The application provides standardized error responses:

### Success Response
```json
{
  "status": "success",
  "timestamp": 1732972800000,
  "data": { ... }
}
```

### Error Response
```json
{
  "status": "error",
  "timestamp": 1732972800000,
  "error": {
    "code": "UNEXPECTED_ERROR",
    "message": "Error description"
  }
}
```

### Validation Error Response
```json
{
  "status": "error",
  "timestamp": 1732972800000,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      {
        "field": "limit",
        "message": "must be greater than 0"
      }
    ]
  }
}
```
