# Cortext

AI-powered developer context intelligence that aggregates knowledge from GitHub, Slack, and Discord directly into your IDE through MCP-compatible tools.

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+
- PostgreSQL 14+
- Docker (optional)

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/context_keeper.git
cd context_keeper

# Install Go dependencies
go mod download

# Install Node dependencies
npm install

# Set up environment variables
cp .env.example .env
# Edit .env with your configuration

# Run database migrations
make migrate

# Start the server
make run
```

### Docker Deployment

```bash
# Development
docker-compose up

# With MCP server
docker-compose --profile with-mcp up

# With Slack bot  
docker-compose --profile with-slack up

# Full stack
docker-compose --profile with-mcp --profile with-slack up

# Production
docker-compose -f docker-compose.yml up -d
```

See [Docker Guide](docs/DOCKER.md) for detailed instructions.

## 📚 Documentation

- [Setup Guide](docs/SETUP_GUIDE.md) - Detailed setup instructions
- [Architecture](docs/ARCHITECTURE.md) - System architecture and design
- [Project Structure](docs/PROJECT_STRUCTURE.md) - Code organization guide
- [Security](SECURITY.md) - Security policies and best practices
- [Docker Deployment](docs/DOCKER.md) - Docker deployment guide
- [Slack Integration](docs/SLACK_INTEGRATION.md) - Slack integration setup
- [Discord Integration](docs/DISCORD_INTEGRATION.md) - Discord integration setup
- [Deployment](deployment/README.md) - Kubernetes deployment guide

## 🏗️ Architecture

```
context_keeper/
├── cmd/                    # Application entrypoints
│   └── server/            # Main server application
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   ├── database/         # Database migrations
│   ├── handlers/         # HTTP request handlers
│   ├── middleware/       # HTTP middleware
│   ├── models/           # Data models
│   ├── repository/       # Data access layer
│   ├── server/           # Server setup
│   └── services/         # Business logic
│       └── connectors/   # Platform connectors (GitHub, Slack, Discord)
├── src/                   # TypeScript/Node.js code
│   ├── mcp/              # MCP server implementation
│   ├── slack/            # Slack bot
│   └── services/         # Node services
├── web/                   # Frontend web application
├── deployment/            # Kubernetes manifests
├── docs/                  # Documentation
├── scripts/               # Utility scripts
└── test/                  # Integration tests
```

## 🔧 Development

### Running Tests

```bash
# Go tests
make test

# Node tests
npm test

# Integration tests
make test-integration
```

### Building

```bash
# Build Go binary
make build

# Build Docker image
docker build -t mcp-context-engine .
```

## 🔌 Integrations

### GitHub
- Pull requests, issues, commits, and code reviews
- File-level change context
- Automatic decision extraction

### Slack
- Channel discussions and threads
- Direct messages
- Engineering decision tracking

### Discord
- Server conversations
- Channel message history
- Community insights

## 🛠️ MCP Tools

The engine provides 5 MCP-compatible tools:

1. **search_project_knowledge** - Search across all project knowledge
2. **get_context_for_file** - Get comprehensive file context
3. **get_decision_history** - Track architectural decisions
4. **list_recent_architecture_discussions** - Recent discussions
5. **explain_why_code_exists** - Code reasoning and history

## 🔒 Security

- End-to-end encryption (AES-256-GCM)
- Complete tenant isolation
- OAuth 2.0 authentication
- Role-based access control
- Audit logging
- SOC 2 compliant

See [SECURITY.md](SECURITY.md) for details.

For questions or issues:
- Check the [documentation](docs/)
- Review existing issues
