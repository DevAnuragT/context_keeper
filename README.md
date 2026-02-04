# ContextKeeper Go Backend

A GitHub repository memory system that preserves engineering context from PRs, issues, and commits. **✅ COMPLETE IMPLEMENTATION**

## Architecture

- **Go Backend**: API orchestration, GitHub integration, data storage
- **Python AI Service**: Context processing and requirement clarification  
- **Frontend**: Web dashboard for context queries

## Features

✅ **Completed Features:**
- GitHub OAuth authentication with JWT tokens
- Repository data ingestion (PRs, issues, commits) with 50/50/100 limits
- Background job processing with goroutine-based workers
- AI-powered context restoration with 30s timeout
- Requirement clarification and context queries
- Property-based testing for all 10 correctness properties
- Structured logging with JSON output
- Health checks and monitoring endpoints
- Docker deployment configuration
- Comprehensive test coverage (unit, integration, system, property-based)

## Development

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Docker (optional)

### Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
# Server Configuration
PORT=8080
ENVIRONMENT=development
LOG_LEVEL=info

# Database Configuration
DATABASE_URL=postgres://localhost/contextkeeper?sslmode=disable

# JWT Configuration
JWT_SECRET=your-secret-key-here

# GitHub OAuth Configuration
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
GITHUB_REDIRECT_URL=http://localhost:8080/api/auth/github

# AI Service Configuration
AI_SERVICE_URL=http://localhost:8000
AI_SERVICE_TIMEOUT=30

# CORS Configuration
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
```

### Running Locally

```bash
# Install dependencies
go mod tidy

# Run with environment variables
source .env
go run cmd/server/main.go

# Or with Docker (development)
docker-compose -f docker-compose.dev.yml up --build

# Or with Docker (production)
docker-compose up -d
```

### Testing

```bash
# Run all tests
go test ./...

# Run with property-based tests
go test ./... -v

# Run integration tests (requires database)
go test ./internal/server

# Run system tests (requires database)
go test ./test
```

### API Endpoints

#### Authentication
- `POST /api/auth/github` - GitHub OAuth callback

#### Repository Management
- `GET /api/repos` - List ingested repositories (🔒 Auth required)
- `POST /api/repos/ingest` - Trigger repository ingestion (🔒 Auth required)
- `GET /api/repos/{id}/status` - Get ingestion status (🔒 Auth required)

#### Context Queries
- `POST /api/context/query` - Query repository context (🔒 Auth required)

#### Monitoring
- `GET /health` - Basic health check
- `GET /ready` - Readiness check with database connectivity
- `GET /metrics` - Application metrics

## Project Structure

```
cmd/
  server/           # Application entry point
internal/
  config/           # Configuration management with validation
  database/         # Database migrations
  handlers/         # HTTP handlers with middleware
  logger/           # Structured logging
  middleware/       # Authentication and CORS middleware
  models/           # Data models with JSONB support
  repository/       # Database operations with PostgreSQL
  server/           # HTTP server with security headers
  services/         # Business logic services
    auth.go         # GitHub OAuth + JWT authentication
    github.go       # GitHub API client with rate limiting
    job.go          # Background job processing
    context.go      # AI service integration
test/               # System tests
secrets/            # Docker secrets (examples)
.env.example        # Environment configuration template
docker-compose.yml  # Production Docker setup
docker-compose.dev.yml # Development Docker setup
DOCKER.md          # Docker deployment guide
```

## Implementation Status

✅ **All Tasks Completed:**
- [x] Project structure and core interfaces
- [x] Database schema and repository layer
- [x] GitHub OAuth authentication service
- [x] GitHub API client with rate limiting
- [x] Background job processing service
- [x] AI service integration
- [x] HTTP API endpoints and middleware
- [x] Integration and service wiring
- [x] Docker deployment configuration
- [x] Structured logging and monitoring
- [x] Comprehensive testing (10 property-based tests)

## Property-Based Testing

The system includes 10 correctness properties validated through property-based testing:

1. **OAuth Scope Consistency** - Validates GitHub OAuth scopes
2. **JWT Authentication Round Trip** - Validates JWT token lifecycle
3. **Repository Data Extraction Limits** - Validates 50/50/100 data limits
4. **Repository Metadata Field Extraction** - Validates required fields
5. **Structured Array Serialization** - Validates JSONB serialization
6. **Ingestion Job Lifecycle** - Validates job state transitions
7. **AI Context Payload Filtering** - Validates 10/10/20 filtering limits
8. **AI Service Timeout Enforcement** - Validates 30s timeout handling
9. **Backend API Authentication Enforcement** - Validates JWT middleware
10. **Fail-Fast Error Handling** - Validates error response structure

## Deployment

See [DOCKER.md](DOCKER.md) for detailed deployment instructions.

### Quick Start with Docker

```bash
# Development
docker-compose -f docker-compose.dev.yml up

# Production (configure secrets first)
docker-compose up -d
```

## Next Steps for You

See the **"What You Need to Do"** section below for required setup steps.