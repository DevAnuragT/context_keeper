# ContextKeeper MCP + Slack Bot Testing Guide

Complete guide for testing all features of the MCP server and Slack bot system.

## Table of Contents

1. [System Overview](#system-overview)
2. [MCP Server Features](#mcp-server-features)
3. [Slack Bot Features](#slack-bot-features)
4. [Terminal Tests](#terminal-tests)
5. [Manual Tests](#manual-tests)
6. [Integration Tests](#integration-tests)
7. [Property-Based Tests](#property-based-tests)

---

## System Overview

**Running Services:**
- **Database**: PostgreSQL with pgvector on port 5434
- **Go Backend**: REST API on http://localhost:8080
- **MCP Server**: Model Context Protocol server on http://localhost:3003
- **Slack Bot**: Slack integration on http://localhost:3002
- **Frontend**: Web UI on http://localhost:3000
- **Health Monitor**: System health on http://localhost:8081

---

## MCP Server Features

### Feature 1: MCP Protocol Compliance
**Description**: Implements standard MCP HTTP transport protocol for AI assistant integration.

**Capabilities**:
- List available resources (repositories, context, timelines)
- List available tools (query_repository_context, get_onboarding_summary)
- Execute tool calls with parameter validation
- Get resource content by URI
- Standard error handling and responses

### Feature 2: Repository Context Queries
**Description**: Natural language queries about repository information.

**Capabilities**:
- Query repository metadata (PRs, issues, commits)
- Search across repository content
- Get clarified understanding of queries
- Receive structured context responses
- Get follow-up question suggestions

### Feature 3: Onboarding Summaries
**Description**: Generate comprehensive onboarding guides for new team members.

**Capabilities**:
- Repository overview and purpose
- Architecture and component breakdown
- Development setup instructions
- Coding standards and guidelines
- Testing and deployment processes
- Getting started checklist

### Feature 4: Resource Management
**Description**: Expose repository data as MCP resources.

**Capabilities**:
- List all available repositories
- Get repository metadata
- Access repository context
- View repository timeline
- Retrieve specific resource content

---

## Slack Bot Features

### Feature 1: /context Command
**Description**: Query repository context through Slack.

**Usage**: `/context <query> [repo:<repository>]`

**Capabilities**:
- Natural language context queries
- Optional repository specification
- Formatted Slack responses
- Query metadata display
- Response truncation for long results

### Feature 2: /onboard Command
**Description**: Get onboarding summary for repositories.

**Usage**: `/onboard <repository>`

**Capabilities**:
- Comprehensive onboarding guides
- Repository-specific information
- Getting started checklists
- Key areas and action items
- Formatted for new team members

### Feature 3: /recent Command
**Description**: View recent repository activity.

**Usage**: `/recent <repository> [days:<number>]`

**Capabilities**:
- Recent commits, PRs, and issues
- Configurable time period (1-30 days)
- Chronological activity timeline
- Activity summaries
- Contributor information

### Feature 4: /status Command
**Description**: Check system health and status.

**Usage**: `/status`

**Capabilities**:
- MCP server health
- Go backend connectivity
- Repository statistics
- Circuit breaker status
- Overall system health

### Feature 5: Error Handling & Resilience
**Description**: Graceful error handling and retry mechanisms.

**Capabilities**:
- Exponential backoff retries
- Circuit breaker protection
- User-friendly error messages
- Timeout handling
- Partial data responses

---

## Terminal Tests

### 1. Health Check Tests

#### Test MCP Server Health
```bash
curl http://localhost:3003/health
```

**Expected Output**:
```json
{
  "status": "healthy",
  "timestamp": "2026-02-10T19:00:00.000Z",
  "version": "1.0.0"
}
```

#### Test Slack Bot Health
```bash
curl http://localhost:3002/health
```

**Expected Output**:
```json
{
  "status": "healthy",
  "timestamp": "2026-02-10T19:00:00.000Z",
  "version": "1.0.0",
  "service": "slack-bot"
}
```

#### Test Go Backend Health
```bash
curl http://localhost:8080/health
```

**Expected Output**:
```json
{
  "status": "healthy",
  "timestamp": "2026-02-10T19:00:00.000Z"
}
```

#### Test System Health Monitor
```bash
curl http://localhost:8081/health
```

**Expected Output**:
```json
{
  "status": "healthy",
  "timestamp": "2026-02-10T19:00:00.000Z",
  "uptime": 1234,
  "components": {
    "mcpServer": { "status": "healthy", "port": 3003 },
    "slackBot": { "status": "healthy", "port": 3002 }
  }
}
```

---

### 2. MCP Protocol Tests

#### List Available Resources
```bash
curl -X POST http://localhost:3003/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "method": "list_resources",
    "id": 1
  }'
```

**Expected Output**:
```json
{
  "result": {
    "resources": [
      {
        "uri": "contextkeeper://repositories",
        "name": "Available Repositories",
        "description": "List of all repositories in the system",
        "mimeType": "application/json"
      }
    ]
  },
  "id": 1
}
```

#### List Available Tools
```bash
curl -X POST http://localhost:3003/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "method": "list_tools",
    "id": 2
  }'
```

**Expected Output**:
```json
{
  "result": {
    "tools": [
      {
        "name": "query_repository_context",
        "description": "Query repository for specific context information using natural language",
        "inputSchema": {
          "type": "object",
          "properties": {
            "query": { "type": "string", "description": "..." },
            "repositoryId": { "type": "string", "description": "..." },
            "limit": { "type": "number", "default": 10 }
          },
          "required": ["query"]
        }
      },
      {
        "name": "get_onboarding_summary",
        "description": "Generate comprehensive onboarding summary for new team members",
        "inputSchema": { "..." }
      }
    ]
  },
  "id": 2
}
```

#### Call query_repository_context Tool
```bash
curl -X POST http://localhost:3003/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "method": "call_tool",
    "params": {
      "name": "query_repository_context",
      "arguments": {
        "query": "What are the main components of this project?",
        "limit": 5
      }
    },
    "id": 3
  }'
```

**Expected Output**:
```json
{
  "result": {
    "content": [{
      "type": "text",
      "text": "## Query Understanding\n...\n\n## Context Information\n..."
    }],
    "isError": false
  },
  "id": 3
}
```

#### Call get_onboarding_summary Tool
```bash
curl -X POST http://localhost:3003/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "method": "call_tool",
    "params": {
      "name": "get_onboarding_summary",
      "arguments": {
        "repositoryId": "1"
      }
    },
    "id": 4
  }'
```

**Expected Output**:
```json
{
  "result": {
    "content": [{
      "type": "text",
      "text": "# Onboarding Guide: ...\n\n## Project Overview\n..."
    }],
    "isError": false
  },
  "id": 4
}
```

#### Get Resource Content
```bash
curl -X POST http://localhost:3003/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "method": "get_resource",
    "params": {
      "uri": "contextkeeper://repositories"
    },
    "id": 5
  }'
```

**Expected Output**:
```json
{
  "result": {
    "contents": [{
      "uri": "contextkeeper://repositories",
      "mimeType": "application/json",
      "text": "[{\"id\":1,\"name\":\"...\"}]"
    }]
  },
  "id": 5
}
```

---

### 3. Error Handling Tests

#### Test Invalid MCP Method
```bash
curl -X POST http://localhost:3003/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "method": "invalid_method",
    "id": 6
  }'
```

**Expected Output**:
```json
{
  "error": {
    "code": -32601,
    "message": "Method not found",
    "data": { "details": "Unknown method: invalid_method" }
  },
  "id": 6
}
```

#### Test Missing Required Parameter
```bash
curl -X POST http://localhost:3003/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "method": "call_tool",
    "params": {
      "name": "query_repository_context",
      "arguments": {}
    },
    "id": 7
  }'
```

**Expected Output**:
```json
{
  "error": {
    "code": -32603,
    "message": "Internal error",
    "data": { "details": "Missing or invalid required parameter: query" }
  },
  "id": 7
}
```

#### Test Malformed Request
```bash
curl -X POST http://localhost:3003/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "invalid": "request"
  }'
```

**Expected Output**:
```json
{
  "error": {
    "code": -32600,
    "message": "Invalid Request",
    "data": { "details": "Missing required field: method" }
  }
}
```

---

### 4. Go Backend Integration Tests

#### List Repositories
```bash
curl http://localhost:8080/api/repositories
```

**Expected Output**:
```json
{
  "repositories": [
    {
      "id": 1,
      "name": "example-repo",
      "full_name": "owner/example-repo",
      "owner": "owner",
      "description": "Example repository",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-15T00:00:00Z"
    }
  ]
}
```

#### Query Repository Context
```bash
curl -X POST http://localhost:8080/api/query \
  -H "Content-Type: application/json" \
  -d '{
    "repo_id": 1,
    "query": "What is this project about?",
    "mode": "query"
  }'
```

**Expected Output**:
```json
{
  "clarified_goal": "Understanding the project purpose and overview",
  "context": "This project is...",
  "tasks": [],
  "questions": ["What are the main features?", "..."]
}
```

---

### 5. Performance Tests

#### Measure MCP Response Time
```bash
time curl -X POST http://localhost:3003/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "method": "list_tools",
    "id": 1
  }'
```

**Expected**: Response time < 100ms

#### Measure Context Query Time
```bash
time curl -X POST http://localhost:3003/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "method": "call_tool",
    "params": {
      "name": "query_repository_context",
      "arguments": {
        "query": "What are the main components?"
      }
    },
    "id": 1
  }'
```

**Expected**: Response time < 2 seconds

---

## Manual Tests

### 1. Slack Bot Command Tests

#### Test /context Command

**Steps**:
1. Open Slack workspace where ContextKeeper bot is installed
2. In any channel or DM, type: `/context What are the main components?`
3. Press Enter

**Expected Result**:
- Bot responds with context information
- Response includes query understanding
- Response includes relevant repository information
- Response is formatted with markdown
- Response includes metadata (repository, query)

**Variations to Test**:
- `/context How does authentication work?`
- `/context Show me recent changes`
- `/context What is the project structure? repo:my-app`
- `/context` (no query - should show help)

---

#### Test /onboard Command

**Steps**:
1. In Slack, type: `/onboard my-repo`
2. Press Enter

**Expected Result**:
- Bot responds with comprehensive onboarding guide
- Includes project overview
- Includes architecture information
- Includes getting started checklist
- Formatted for readability

**Variations to Test**:
- `/onboard` (no repository - should show help or use default)
- `/onboard non-existent-repo` (should show error)

---

#### Test /recent Command

**Steps**:
1. In Slack, type: `/recent my-repo`
2. Press Enter

**Expected Result**:
- Bot responds with recent activity
- Shows commits, PRs, issues from last 7 days
- Chronologically ordered
- Includes contributor information

**Variations to Test**:
- `/recent my-repo days:14`
- `/recent my-repo days:3`
- `/recent days:7` (should use default repo if configured)
- `/recent` (should show help)

---

#### Test /status Command

**Steps**:
1. In Slack, type: `/status`
2. Press Enter

**Expected Result**:
- Bot responds with system status
- Shows MCP server health
- Shows Go backend health
- Shows repository statistics
- Shows overall system status
- Includes timestamp

---

### 2. Error Handling Tests

#### Test Backend Unavailable

**Steps**:
1. Stop the Go backend: `kill <backend-pid>`
2. In Slack, type: `/context What is this project?`
3. Observe error message

**Expected Result**:
- User-friendly error message
- Explains backend is unavailable
- Suggests retry or contacting admin
- No technical stack traces visible

**Cleanup**: Restart Go backend

---

#### Test Invalid Repository

**Steps**:
1. In Slack, type: `/onboard invalid-repo-12345`
2. Press Enter

**Expected Result**:
- Clear error message
- Explains repository not found
- Suggests checking repository name
- Provides usage example

---

#### Test Timeout Scenario

**Steps**:
1. In Slack, type: `/context <very complex query>`
2. Wait for response

**Expected Result**:
- If timeout occurs, clear timeout message
- Suggests retry or simplifying query
- No hanging or unclear state

---

### 3. Frontend Tests

#### Test Landing Page

**Steps**:
1. Open http://localhost:3000 in browser
2. Verify page loads correctly

**Expected Result**:
- Page loads without errors
- All styles applied correctly
- Navigation works
- Links are functional

---

#### Test Signup Flow

**Steps**:
1. Navigate to http://localhost:3000/signup
2. Fill in signup form
3. Submit

**Expected Result**:
- Form validation works
- API call to backend succeeds
- Appropriate success/error message shown

---

#### Test Login Flow

**Steps**:
1. Navigate to http://localhost:3000/login
2. Enter credentials
3. Submit

**Expected Result**:
- Authentication works
- Redirects to dashboard on success
- Shows error on invalid credentials

---

### 4. Integration Flow Tests

#### End-to-End Context Query

**Steps**:
1. Ensure all services running
2. In Slack, type: `/context What is the architecture?`
3. Observe full flow

**Expected Flow**:
1. Slack receives command
2. Slack bot validates and parses
3. Slack bot calls MCP server
4. MCP server calls Go backend
5. Go backend queries database
6. Response flows back through stack
7. Formatted response shown in Slack

**Expected Result**:
- Complete flow succeeds
- Response time < 3 seconds
- Accurate information returned
- Proper formatting maintained

---

#### End-to-End Onboarding

**Steps**:
1. In Slack, type: `/onboard <repository>`
2. Wait for complete response

**Expected Flow**:
1. Slack bot receives command
2. Multiple context queries executed
3. Onboarding guide assembled
4. Formatted response returned

**Expected Result**:
- Comprehensive guide generated
- All sections populated
- Response time < 10 seconds
- Readable and useful information

---

## Integration Tests

### Run Automated Integration Tests

```bash
# Run all integration tests
npm test -- integration.test.ts

# Run specific integration test suite
npm test -- src/mcp/server.integration.test.ts

# Run with verbose output
npm test -- --reporter=verbose integration
```

**Test Coverage**:
- MCP server startup and shutdown
- Tool execution end-to-end
- Resource retrieval end-to-end
- Error handling scenarios
- Slack command processing
- Backend integration

---

## Property-Based Tests

### Run Property-Based Tests

```bash
# Run all property-based tests
npm test -- properties.test.ts

# Run with verbose output
npm test:pbt

# Run specific property test
npm test -- src/mcp/server.properties.test.ts
```

**Property Tests Include**:
- MCP server port binding (Property 1)
- Resource response completeness (Property 2)
- Tool implementation correctness (Property 4)
- Error message formatting (Property 6)
- Configuration validation (Property 10)
- Retry mechanism behavior (Property 12)
- Timeout handling (Property 13)
- MCP protocol compliance (Property 16)

---

## Test Checklist

### Pre-Test Setup
- [ ] All services running (database, backend, MCP, Slack bot, frontend)
- [ ] Environment variables configured
- [ ] Slack app installed in workspace
- [ ] Test repositories ingested in database

### MCP Server Tests
- [ ] Health check passes
- [ ] List resources works
- [ ] List tools works
- [ ] Query context tool works
- [ ] Onboarding summary tool works
- [ ] Get resource works
- [ ] Error handling works
- [ ] Invalid requests handled

### Slack Bot Tests
- [ ] /context command works
- [ ] /onboard command works
- [ ] /recent command works
- [ ] /status command works
- [ ] Help messages display
- [ ] Error messages user-friendly
- [ ] Response formatting correct
- [ ] Signature verification works

### Integration Tests
- [ ] End-to-end context query
- [ ] End-to-end onboarding
- [ ] Backend integration
- [ ] Error propagation
- [ ] Timeout handling
- [ ] Circuit breaker activation

### Performance Tests
- [ ] MCP response time < 100ms
- [ ] Context query < 2s
- [ ] Onboarding < 10s
- [ ] Slack commands < 3s

### Property-Based Tests
- [ ] All properties pass 100+ iterations
- [ ] No failing examples found
- [ ] Edge cases covered
- [ ] Shrinking works correctly

---

## Troubleshooting

### MCP Server Not Responding
```bash
# Check if running
curl http://localhost:3003/health

# Check logs
# (View process output in terminal)

# Restart if needed
# Stop and start npm run dev
```

### Slack Bot Commands Failing
```bash
# Check bot health
curl http://localhost:3002/health

# Verify Slack credentials in .env
cat .env | grep SLACK

# Check signature verification
# Review logs for signature errors
```

### Backend Connection Issues
```bash
# Check backend health
curl http://localhost:8080/health

# Check database connection
docker ps | grep postgres

# Verify DATABASE_URL in .env
```

### Database Issues
```bash
# Check database running
docker ps | grep postgres

# Check database logs
docker logs contextkeeper-postgres-vector

# Restart database if needed
docker restart contextkeeper-postgres-vector
```

---

## Test Results Documentation

### Recording Test Results

Create a test results file:
```bash
# Create test results directory
mkdir -p test-results

# Run tests and save output
npm test > test-results/unit-tests-$(date +%Y%m%d).log 2>&1
npm test:pbt > test-results/pbt-tests-$(date +%Y%m%d).log 2>&1
```

### Test Report Template

```markdown
# Test Report - [Date]

## Environment
- Database: Running ✓
- Go Backend: Running ✓
- MCP Server: Running ✓
- Slack Bot: Running ✓
- Frontend: Running ✓

## Test Results

### Unit Tests
- Total: X
- Passed: X
- Failed: X
- Duration: Xs

### Integration Tests
- Total: X
- Passed: X
- Failed: X
- Duration: Xs

### Property-Based Tests
- Total: X
- Passed: X
- Failed: X
- Iterations: 100+

### Manual Tests
- MCP Protocol: ✓
- Slack Commands: ✓
- Error Handling: ✓
- Performance: ✓

## Issues Found
1. [Issue description]
2. [Issue description]

## Notes
[Any additional observations]
```

---

## Next Steps

After completing all tests:

1. **Document Results**: Record all test outcomes
2. **Fix Issues**: Address any failing tests
3. **Performance Tuning**: Optimize slow operations
4. **User Feedback**: Gather feedback from Slack users
5. **Monitoring**: Set up ongoing health monitoring
6. **Documentation**: Update user guides based on testing

---

## Support

For issues or questions:
- Check logs in terminal/console
- Review error messages carefully
- Verify all services are running
- Check environment configuration
- Consult system documentation
