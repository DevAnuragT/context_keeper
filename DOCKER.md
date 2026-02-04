# Docker Deployment Guide

This guide explains how to deploy the ContextKeeper Go Backend using Docker.

## Development Setup

For local development with Docker:

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your configuration values.

3. Start the development environment:
   ```bash
   docker-compose -f docker-compose.dev.yml up --build
   ```

This will start:
- PostgreSQL database on port 5432
- Backend API on port 8080

## Production Setup

For production deployment:

1. Create secret files:
   ```bash
   cp secrets/postgres_password.txt.example secrets/postgres_password.txt
   cp secrets/jwt_secret.txt.example secrets/jwt_secret.txt
   cp secrets/github_client_secret.txt.example secrets/github_client_secret.txt
   ```

2. Edit the secret files with your actual secrets.

3. Set environment variables:
   ```bash
   export GITHUB_CLIENT_ID=your-github-client-id
   export GITHUB_REDIRECT_URL=https://your-domain.com/api/auth/github
   export ALLOWED_ORIGINS=https://your-frontend-domain.com
   ```

4. Start the production environment:
   ```bash
   docker-compose up -d
   ```

## Environment Variables

### Required
- `GITHUB_CLIENT_ID`: GitHub OAuth application client ID
- `GITHUB_CLIENT_SECRET` or `GITHUB_CLIENT_SECRET_FILE`: GitHub OAuth client secret
- `JWT_SECRET` or `JWT_SECRET_FILE`: Secret key for JWT token signing

### Optional
- `PORT`: Server port (default: 8080)
- `DATABASE_URL`: PostgreSQL connection string
- `AI_SERVICE_URL`: AI service endpoint (default: http://localhost:8000)
- `AI_SERVICE_TIMEOUT`: AI service timeout in seconds (default: 30)
- `ALLOWED_ORIGINS`: Comma-separated list of allowed CORS origins
- `ENVIRONMENT`: Environment mode (development/production)
- `LOG_LEVEL`: Logging level (debug/info/warn/error)

## Health Checks

The application provides a health check endpoint at `/health` that returns:
```json
{"status": "ok"}
```

## Security

- The application runs as a non-root user in the container
- Secrets are managed using Docker secrets in production
- Security headers are automatically added to all responses
- CORS is configured with explicit origin allowlists

## Scaling

The application is stateless and can be scaled horizontally:

```bash
docker-compose up --scale backend=3
```

Use a load balancer (nginx, HAProxy, etc.) to distribute traffic across instances.

## Monitoring

- Application logs are written to stdout/stderr
- Health check endpoint for monitoring systems
- Resource limits are configured in docker-compose.yml

## Troubleshooting

1. **Database connection issues**: Ensure PostgreSQL is healthy before starting the backend
2. **Permission errors**: Check that the application user has correct permissions
3. **Secret loading**: Verify secret files exist and are readable
4. **CORS errors**: Check ALLOWED_ORIGINS environment variable