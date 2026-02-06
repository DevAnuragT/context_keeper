# Testing Guide - MCP Context Engine

This document describes the comprehensive testing strategy for the MCP Context Engine project, including integration tests, performance tests, and security validation.

## Overview

The testing suite validates:
- **End-to-end user journeys** from signup to context queries
- **Multi-platform data ingestion** and processing
- **Performance and scalability** under load
- **Security and compliance** requirements
- **Tenant isolation** and data security

## Test Structure

```
test/
├── e2e_integration_test.go    # End-to-end integration tests
├── performance_test.go         # Performance and load tests
├── security_test.go            # Security and compliance tests
└── system_test.go              # System-level tests

tests/web/
├── auth.integration.test.ts              # Authentication flow tests
├── dashboard.integration.test.ts         # Dashboard functionality tests
├── integration-wizard.integration.test.ts # Integration wizard tests
└── e2e-user-journey.integration.test.ts  # Complete user journey tests
```

## Running Tests

### Quick Start

Run all tests:
```bash
./scripts/run-integration-tests.sh
```

### Individual Test Suites

**Go Backend Tests:**
```bash
# Unit tests only
go test ./internal/... -v -short

# End-to-end integration tests
go test ./test/e2e_integration_test.go -v

# Performance tests
go test ./test/performance_test.go -v

# Security tests
go test ./test/security_test.go -v
```

**TypeScript/Node.js Tests:**
```bash
# All tests
npm test

# Specific test files
npm test tests/web/auth.integration.test.ts
npm test tests/web/e2e-user-journey.integration.test.ts

# Watch mode
npm run test:watch
```

## Test Categories

### 1. End-to-End Integration Tests

**File:** `test/e2e_integration_test.go`

Tests the complete user journey:
1. User signup and email verification
2. User login and JWT token generation
3. Project workspace creation
4. Platform integration connection (GitHub, Slack, Discord)
5. Data ingestion and processing
6. Context queries via MCP tools
7. File context retrieval
8. Decision history queries

**Key Tests:**
- `TestEndToEndUserJourney` - Complete user flow
- `TestMultiPlatformIngestion` - Multi-platform data ingestion
- `TestTenantIsolation` - Tenant data isolation
- `TestSecurityBoundaries` - Authentication and authorization

**Run:**
```bash
go test ./test/e2e_integration_test.go -v -timeout 5m
```

### 2. Performance and Load Tests

**File:** `test/performance_test.go`

Validates performance requirements:
- Response times under 1 second for MCP tool queries
- Caching improves performance by 30%+
- System handles 50+ concurrent users
- Sustained load of 20 req/s for 30 seconds
- Graceful degradation under stress

**Key Tests:**
- `TestResponseTimeRequirements` - Response time validation
- `TestLoadHandling` - Concurrent and sustained load
- `TestGracefulDegradation` - Stress testing
- `TestHorizontalScaling` - Multi-instance deployment

**Run:**
```bash
go test ./test/performance_test.go -v -timeout 10m
```

**Performance Metrics:**
- P50 response time: < 500ms
- P95 response time: < 1s
- P99 response time: < 2s
- Success rate: > 99%
- Throughput: > 10 req/s

### 3. Security and Compliance Tests

**File:** `test/security_test.go`

Validates security controls:
- Password strength requirements
- JWT token security and expiration
- OAuth CSRF protection
- Session management
- Data encryption
- Role-based access control
- Tenant isolation
- Input validation (SQL injection, XSS, path traversal)
- Error handling
- GDPR compliance

**Key Tests:**
- `TestAuthenticationFlows` - Auth mechanisms
- `TestDataEncryption` - Sensitive data protection
- `TestAuthorizationControls` - RBAC validation
- `TestTenantIsolationSecurity` - Tenant data isolation
- `TestInputValidation` - Injection prevention
- `TestErrorHandling` - Secure error handling
- `TestComplianceRequirements` - GDPR compliance

**Run:**
```bash
go test ./test/security_test.go -v -timeout 5m
```

### 4. Web Integration Tests

**Files:** `tests/web/*.integration.test.ts`

Tests web platform functionality:
- Authentication flows (signup, login, OAuth, password reset)
- Dashboard operations (project CRUD, metrics, activity)
- Integration wizard (platform connection, source selection)
- Complete user journeys

**Run:**
```bash
npm test tests/web/
```

## Test Requirements

### Prerequisites

1. **Test Database:**
   ```bash
   createdb contextkeeper_test
   psql contextkeeper_test < schema.sql
   ```

2. **Environment Variables:**
   ```bash
   cp .env.example .env.test
   # Edit .env.test with test configuration
   ```

3. **Dependencies:**
   ```bash
   # Go dependencies
   go mod download

   # Node.js dependencies
   npm install
   ```

### Test Data

Tests use mock data and in-memory databases where possible. For integration tests requiring real databases:

```bash
# Setup test database
make setup-test-db

# Seed test data
make seed-test-data

# Clean test data
make clean-test-data
```

## Continuous Integration

### GitHub Actions

Tests run automatically on:
- Pull requests
- Pushes to main branch
- Nightly builds

**Workflow:** `.github/workflows/test.yml`

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run tests
        run: ./scripts/run-integration-tests.sh
```

## Test Coverage

Target coverage: **80%+**

Generate coverage reports:

```bash
# Go coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# TypeScript coverage
npm test -- --coverage
```

## Debugging Tests

### Verbose Output

```bash
# Go tests with verbose output
go test -v ./test/...

# TypeScript tests with verbose output
npm test -- --verbose
```

### Running Single Tests

```bash
# Go - run specific test
go test -v -run TestEndToEndUserJourney ./test/

# TypeScript - run specific test
npm test -- -t "should complete user signup"
```

### Test Timeouts

Increase timeout for slow tests:

```bash
go test -timeout 30m ./test/performance_test.go
```

## Best Practices

1. **Isolation:** Each test should be independent and not rely on other tests
2. **Cleanup:** Always clean up test data after tests complete
3. **Mocking:** Use mocks for external services (GitHub API, Slack API, etc.)
4. **Assertions:** Use clear, descriptive assertion messages
5. **Performance:** Keep unit tests fast (< 1s), integration tests reasonable (< 5m)

## Troubleshooting

### Common Issues

**Database Connection Errors:**
```bash
# Check database is running
psql -h localhost -U test -d contextkeeper_test -c "SELECT 1"

# Reset test database
dropdb contextkeeper_test
createdb contextkeeper_test
```

**Port Conflicts:**
```bash
# Check if port is in use
lsof -i :8080

# Kill process using port
kill -9 <PID>
```

**Test Timeouts:**
- Increase timeout with `-timeout` flag
- Check for deadlocks or infinite loops
- Verify external services are responding

## Success Criteria

All tests must pass before merging to main:

- ✅ All unit tests pass
- ✅ All integration tests pass
- ✅ Performance requirements met
- ✅ Security tests pass
- ✅ Code coverage > 80%
- ✅ No critical security vulnerabilities

## Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Vitest Documentation](https://vitest.dev/)
- [Property-Based Testing Guide](../docs/PROPERTY_BASED_TESTING.md)
- [Security Testing Guide](../docs/SECURITY_TESTING.md)
