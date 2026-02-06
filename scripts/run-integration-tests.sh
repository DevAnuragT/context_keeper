#!/bin/bash

# Integration Test Runner for MCP Context Engine
# This script runs all integration, performance, and security tests

set -e

echo "========================================="
echo "MCP Context Engine - Integration Tests"
echo "========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if test database is available
echo "Checking test database..."
if ! psql -h localhost -U test -d contextkeeper_test -c "SELECT 1" > /dev/null 2>&1; then
    echo -e "${YELLOW}Warning: Test database not available. Some tests may fail.${NC}"
    echo "To set up test database, run: make setup-test-db"
fi

# Run Go backend tests
echo ""
echo "========================================="
echo "Running Go Backend Tests"
echo "========================================="
cd "$(dirname "$0")/.."

echo ""
echo "1. Running unit tests..."
go test ./internal/... -v -short

echo ""
echo "2. Running end-to-end integration tests..."
go test ./test/e2e_integration_test.go -v -timeout 5m

echo ""
echo "3. Running performance tests..."
go test ./test/performance_test.go -v -timeout 10m

echo ""
echo "4. Running security tests..."
go test ./test/security_test.go -v -timeout 5m

# Run TypeScript/Node.js tests
echo ""
echo "========================================="
echo "Running TypeScript/Node.js Tests"
echo "========================================="

echo ""
echo "5. Running MCP server tests..."
npm test -- src/mcp/

echo ""
echo "6. Running web integration tests..."
npm test -- tests/web/

# Generate test report
echo ""
echo "========================================="
echo "Test Summary"
echo "========================================="

# Count test results
GO_TESTS=$(go test ./... -v 2>&1 | grep -c "PASS:" || true)
TS_TESTS=$(npm test 2>&1 | grep -oP '\d+(?= passed)' | head -1 || echo "0")

echo ""
echo -e "${GREEN}✓ Go Tests Passed: $GO_TESTS${NC}"
echo -e "${GREEN}✓ TypeScript Tests Passed: $TS_TESTS${NC}"
echo ""

# Check for failures
if go test ./... > /dev/null 2>&1 && npm test > /dev/null 2>&1; then
    echo -e "${GREEN}=========================================${NC}"
    echo -e "${GREEN}All Tests Passed!${NC}"
    echo -e "${GREEN}=========================================${NC}"
    exit 0
else
    echo -e "${RED}=========================================${NC}"
    echo -e "${RED}Some Tests Failed${NC}"
    echo -e "${RED}=========================================${NC}"
    exit 1
fi
