# ContextKeeper MCP + Slack Bot Deployment Guide

This directory contains deployment configurations for the ContextKeeper MCP + Slack Bot system.

## Overview

The ContextKeeper MCP + Slack Bot provides:
- **MCP Server**: Model Context Protocol server for AI assistant integration
- **Slack Bot**: Team collaboration commands for repository context queries
- **Health Monitoring**: Built-in health checks and monitoring endpoints

## Deployment Options

### 1. Docker Compose (Recommended for Development)

#### Prerequisites
- Docker and Docker Compose installed
- ContextKeeper Go Backend running or available
- Required secret files in `./secrets/` directory

#### Quick Start
```bash
# 1. Create secrets directory and files
mkdir -p secrets
echo "your-slack-signing-secret" > secrets/slack_signing_secret.txt
echo "your-github-client-id" > secrets/github_client_id.txt
echo "your-github-client-secret" > secrets/github_client_secret.txt
echo "your-jwt-secret" > secrets/jwt_secret.txt
echo "your-postgres-password" > secrets/postgres_password.txt

# 2. Start the services
docker-compose -f docker-compose.mcp-slack.yml up -d

# 3. Check health
curl http://localhost:8080/health
```

#### Configuration
Environment variables can be configured in the `docker-compose.mcp-slack.yml` file:

- **MCP_GO_BACKEND_URL**: URL of the Go backend service
- **SLACK_SIGNING_SECRET_FILE**: Path to Slack signing secret file
- **DEMO_MODE**: Enable/disable demo mode for testing
- **LOG_LEVEL**: Logging level (debug, info, warn, error)

### 2. Kubernetes (Recommended for Production)

#### Prerequisites
- Kubernetes cluster (1.19+)
- kubectl configured
- Ingress controller (nginx recommended)
- cert-manager for TLS certificates (optional)

#### Deployment Steps

1. **Create namespace:**
```bash
kubectl create namespace contextkeeper
```

2. **Configure secrets:**
```bash
# Copy the template and edit with your actual secrets
cp deployment/kubernetes/secrets.yaml.example deployment/kubernetes/secrets.yaml
# Edit deployment/kubernetes/secrets.yaml with your actual secrets
# Then apply:
kubectl apply -f deployment/kubernetes/secrets.yaml
```

3. **Deploy the application:**
```bash
kubectl apply -f deployment/kubernetes/mcp-slack-deployment.yaml
```

4. **Verify deployment:**
```bash
kubectl get pods -n contextkeeper
kubectl get services -n contextkeeper
```

5. **Check health:**
```bash
kubectl port-forward -n contextkeeper svc/contextkeeper-mcp-slack 8080:8080
curl http://localhost:8080/health
```

#### Scaling
```bash
# Scale to 3 replicas
kubectl scale deployment contextkeeper-mcp-slack -n contextkeeper --replicas=3
```

### 3. Standalone Docker

#### Build and run:
```bash
# Build the image
docker build -f Dockerfile.mcp-slack -t contextkeeper/mcp-slack:latest .

# Run with environment variables
docker run -d \
  --name contextkeeper-mcp-slack \
  -p 3001:3001 \
  -p 3002:3002 \
  -p 8080:8080 \
  -e MCP_GO_BACKEND_URL=http://your-backend:3000 \
  -e SLACK_SIGNING_SECRET=your-slack-secret \
  contextkeeper/mcp-slack:latest
```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `NODE_ENV` | Node.js environment | `production` | No |
| `LOG_LEVEL` | Logging level | `info` | No |
| `MCP_PORT` | MCP server port | `3001` | No |
| `MCP_GO_BACKEND_URL` | Go backend URL | - | Yes |
| `MCP_TIMEOUT` | Request timeout (ms) | `30000` | No |
| `SLACK_PORT` | Slack bot port | `3002` | No |
| `SLACK_SIGNING_SECRET` | Slack signing secret | - | Yes |
| `SLACK_MCP_SERVER_URL` | MCP server URL | `http://localhost:3001` | No |
| `HEALTH_PORT` | Health server port | `8080` | No |
| `DEMO_MODE` | Enable demo mode | `false` | No |

### Secret Files

When using Docker Compose, secrets are loaded from files:
- `secrets/slack_signing_secret.txt`
- `secrets/github_client_id.txt`
- `secrets/github_client_secret.txt`
- `secrets/jwt_secret.txt`
- `secrets/postgres_password.txt`

## Monitoring and Health Checks

### Health Endpoints

- **`GET /health`**: Basic health check
- **`GET /status`**: Detailed system status
- **`GET /ready`**: Kubernetes readiness probe
- **`POST /restart`**: Graceful restart trigger

### Example Health Response
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00.000Z",
  "version": "1.0.0",
  "uptime": 3600,
  "components": {
    "mcpServer": {
      "status": "healthy",
      "port": 3001
    },
    "slackBot": {
      "status": "healthy",
      "port": 3002
    }
  }
}
```

### Graceful Restart

The application supports graceful restarts:

```bash
# Via HTTP endpoint
curl -X POST http://localhost:8080/restart

# Via signal (Docker/Kubernetes)
kill -USR1 <pid>

# Via kubectl (Kubernetes)
kubectl rollout restart deployment/contextkeeper-mcp-slack -n contextkeeper
```

## Troubleshooting

### Common Issues

1. **MCP Server not responding**
   - Check Go backend connectivity
   - Verify MCP_GO_BACKEND_URL configuration
   - Check network policies in Kubernetes

2. **Slack commands failing**
   - Verify SLACK_SIGNING_SECRET is correct
   - Check Slack app configuration
   - Review application logs

3. **Health checks failing**
   - Check if all components are running
   - Verify port configurations
   - Review resource limits

### Logs

```bash
# Docker Compose
docker-compose -f docker-compose.mcp-slack.yml logs -f

# Kubernetes
kubectl logs -f deployment/contextkeeper-mcp-slack -n contextkeeper

# Docker
docker logs -f contextkeeper-mcp-slack
```

### Debug Mode

Enable debug logging:
```bash
# Set LOG_LEVEL=debug in your environment
export LOG_LEVEL=debug
```

## Security Considerations

1. **Secrets Management**: Use proper secret management (Kubernetes secrets, Docker secrets, etc.)
2. **Network Security**: Configure appropriate network policies
3. **TLS**: Enable TLS for production deployments
4. **Resource Limits**: Set appropriate CPU and memory limits
5. **Non-root User**: Application runs as non-root user (UID 1001)

## Performance Tuning

### Resource Recommendations

| Environment | CPU | Memory | Replicas |
|-------------|-----|--------|----------|
| Development | 250m | 256Mi | 1 |
| Staging | 500m | 512Mi | 2 |
| Production | 1000m | 1Gi | 3+ |

### Scaling Considerations

- MCP server is stateless and can be scaled horizontally
- Slack bot handles concurrent requests efficiently
- Consider load balancing for high-traffic scenarios

## Integration with ContextKeeper Backend

The MCP + Slack Bot requires the ContextKeeper Go Backend to be running and accessible. Ensure:

1. Backend is deployed and healthy
2. Network connectivity between services
3. Proper authentication configuration
4. Database is accessible to backend

For backend deployment instructions, see the main ContextKeeper documentation.