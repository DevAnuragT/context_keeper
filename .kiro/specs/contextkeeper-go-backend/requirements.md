# Requirements Document

## Introduction

ContextKeeper is a GitHub repository memory system that preserves engineering context from PRs, issues, and commits. The Go backend serves as the core orchestration layer, handling GitHub OAuth authentication, data ingestion from GitHub repositories, normalized storage in PostgreSQL, and API coordination with a Python AI service for context restoration and requirement clarification.

## Glossary

- **ContextKeeper_Backend**: The Go-based backend service that orchestrates all system operations
- **GitHub_API**: GitHub's REST API used for repository data extraction
- **AI_Service**: External Python service that processes repository metadata for context queries
- **Ingestion_Job**: Background process that extracts and stores repository data
- **Repository_Metadata**: Structured data extracted from GitHub repositories (PRs, issues, commits)
- **Context_Query**: User request for repository context restoration or requirement clarification

## Requirements

### Requirement 1: GitHub OAuth Authentication

**User Story:** As a developer, I want to authenticate with GitHub OAuth, so that I can securely access my repository data.

#### Acceptance Criteria

1. WHEN a user initiates GitHub OAuth, THE ContextKeeper_Backend SHALL request scopes: public_repo, read:user, user:email
2. WHEN GitHub OAuth callback is received, THE ContextKeeper_Backend SHALL exchange the authorization code for an access token
3. WHEN OAuth is successful, THE ContextKeeper_Backend SHALL generate and return a JWT token for subsequent API calls
4. WHEN OAuth fails, THE ContextKeeper_Backend SHALL return a descriptive error message
5. WHEN a JWT token is provided, THE ContextKeeper_Backend SHALL validate it before processing API requests

### Requirement 2: Repository Data Ingestion

**User Story:** As a developer, I want to manually trigger repository ingestion, so that I can preserve context from specific repositories I choose.

#### Acceptance Criteria

1. WHEN a user triggers repository ingestion, THE ContextKeeper_Backend SHALL create an Ingestion_Job with status "pending"
2. WHEN processing an ingestion job, THE ContextKeeper_Backend SHALL extract the last 50 pull requests from the repository
3. WHEN processing an ingestion job, THE ContextKeeper_Backend SHALL extract the last 50 issues from the repository
4. WHEN processing an ingestion job, THE ContextKeeper_Backend SHALL extract the last 100 commits from the repository
5. WHEN making GitHub_API calls, THE ContextKeeper_Backend SHALL implement sequential processing with basic error handling
6. WHEN GitHub_API calls fail, THE ContextKeeper_Backend SHALL retry at most once before skipping the item
7. WHEN GitHub_API rate limits are encountered, THE ContextKeeper_Backend SHALL handle the error gracefully and mark the job as "partial"

### Requirement 3: Metadata Extraction and Storage

**User Story:** As a system administrator, I want repository data stored in a normalized relational schema, so that queries are efficient and data integrity is maintained.

#### Acceptance Criteria

1. WHEN storing pull request data, THE ContextKeeper_Backend SHALL extract only: id, number, title, body, author, state, created_at, merged_at, files_changed, labels
2. WHEN storing issue data, THE ContextKeeper_Backend SHALL extract only: id, title, body, author, state, created_at, closed_at, labels
3. WHEN storing commit data, THE ContextKeeper_Backend SHALL extract only: sha, message, author, created_at, files_changed
4. WHEN storing files_changed data, THE ContextKeeper_Backend SHALL serialize it as JSONB arrays
5. WHEN storing labels data, THE ContextKeeper_Backend SHALL serialize it as JSONB arrays
6. THE ContextKeeper_Backend SHALL maintain database indexes on: (repo_id), (repo_id, created_at), (repo_id, author)

### Requirement 4: Background Job Processing

**User Story:** As a developer, I want ingestion jobs to run in the background, so that I can continue working while repository data is being processed.

#### Acceptance Criteria

1. WHEN an ingestion job is created, THE ContextKeeper_Backend SHALL process it using goroutines
2. WHEN a job starts processing, THE ContextKeeper_Backend SHALL update its status to "running"
3. WHEN a job completes successfully, THE ContextKeeper_Backend SHALL update its status to "completed"
4. WHEN a job encounters errors but processes some data, THE ContextKeeper_Backend SHALL update its status to "partial"
5. WHEN a job fails completely, THE ContextKeeper_Backend SHALL update its status to "failed" and store the error message
6. WHEN queried for job status, THE ContextKeeper_Backend SHALL return current status, timestamps, and any error messages

### Requirement 5: API Endpoint Design

**User Story:** As a client application, I want action-based REST endpoints, so that I can perform specific operations efficiently.

#### Acceptance Criteria

1. WHEN receiving POST /api/auth/github requests, THE ContextKeeper_Backend SHALL handle GitHub OAuth callbacks
2. WHEN receiving GET /api/repos requests, THE ContextKeeper_Backend SHALL return the authenticated user's ingested repositories
3. WHEN receiving POST /api/repos/ingest requests, THE ContextKeeper_Backend SHALL trigger repository ingestion jobs
4. WHEN receiving GET /api/repos/{id}/status requests, THE ContextKeeper_Backend SHALL return ingestion job status for the specified repository
5. WHEN receiving POST /api/context/query requests, THE ContextKeeper_Backend SHALL process unified context queries
6. WHEN processing any API request, THE ContextKeeper_Backend SHALL validate JWT authentication before proceeding

### Requirement 6: AI Service Integration

**User Story:** As a developer, I want to query repository context, so that I can restore engineering context and clarify requirements.

#### Acceptance Criteria

1. WHEN processing context queries, THE ContextKeeper_Backend SHALL filter and send the last 10 pull requests to the AI_Service
2. WHEN processing context queries, THE ContextKeeper_Backend SHALL filter and send the last 10 issues to the AI_Service
3. WHEN processing context queries, THE ContextKeeper_Backend SHALL filter and send the last 20 commits to the AI_Service
4. WHEN calling the AI_Service, THE ContextKeeper_Backend SHALL implement a 30-second timeout
5. WHEN the AI_Service call times out, THE ContextKeeper_Backend SHALL return an error response immediately
6. WHEN the AI_Service returns a response, THE ContextKeeper_Backend SHALL forward it to the client without caching
7. THE ContextKeeper_Backend SHALL NOT perform semantic relevance ranking; all semantic processing is delegated to the AI_Service

### Requirement 7: Database Schema Design

**User Story:** As a database administrator, I want a normalized relational schema, so that data is efficiently stored and queried.

#### Acceptance Criteria

1. THE ContextKeeper_Backend SHALL create a repos table with: id, name, full_name, owner, created_at, updated_at
2. THE ContextKeeper_Backend SHALL create a pull_requests table with: id, repo_id, number, title, body, author, state, created_at, merged_at, files_changed (JSONB), labels (JSONB)
3. THE ContextKeeper_Backend SHALL create an issues table with: id, repo_id, title, body, author, state, created_at, closed_at, labels (JSONB)
4. THE ContextKeeper_Backend SHALL create a commits table with: sha, repo_id, message, author, created_at, files_changed (JSONB)
5. THE ContextKeeper_Backend SHALL create an ingestion_jobs table with: id, repo_id, status, started_at, finished_at, error_message
6. THE ContextKeeper_Backend SHALL enforce foreign key relationships between all tables and repos

### Requirement 8: Error Handling and Resilience

**User Story:** As a system operator, I want fail-fast error handling, so that issues are quickly identified and communicated.

#### Acceptance Criteria

1. WHEN GitHub_API calls fail, THE ContextKeeper_Backend SHALL log the error and continue processing other items
2. WHEN database operations fail, THE ContextKeeper_Backend SHALL return appropriate HTTP error codes
3. WHEN AI_Service calls fail, THE ContextKeeper_Backend SHALL return error responses without retrying
4. WHEN invalid JWT tokens are provided, THE ContextKeeper_Backend SHALL return 401 Unauthorized responses
5. WHEN malformed requests are received, THE ContextKeeper_Backend SHALL return 400 Bad Request responses with descriptive messages

### Requirement 9: Technical Architecture Constraints

**User Story:** As a system architect, I want the backend built with Go standard library only, so that the system remains lightweight and maintainable.

#### Acceptance Criteria

1. THE ContextKeeper_Backend SHALL use only Go standard library packages (including but not limited to net/http, database/sql, context)
2. THE ContextKeeper_Backend SHALL NOT use external frameworks like Gin, Fiber, or GORM
3. THE ContextKeeper_Backend SHALL be deployable as a single Docker container
4. THE ContextKeeper_Backend SHALL support local-first architecture without external dependencies
5. THE ContextKeeper_Backend SHALL implement all HTTP routing using net/http package

### Requirement 10: Data Filtering and Rate Limiting

**User Story:** As a cost-conscious developer, I want data ingestion to be limited and filtered, so that API usage and storage costs remain manageable.

#### Acceptance Criteria

1. WHEN ingesting pull requests, THE ContextKeeper_Backend SHALL limit extraction to the most recent 50 items
2. WHEN ingesting issues, THE ContextKeeper_Backend SHALL limit extraction to the most recent 50 items
3. WHEN ingesting commits, THE ContextKeeper_Backend SHALL limit extraction to the most recent 100 items
4. WHEN sending data to AI_Service, THE ContextKeeper_Backend SHALL filter to the most recent 10 pull requests
5. WHEN sending data to AI_Service, THE ContextKeeper_Backend SHALL filter to the most recent 10 issues
6. WHEN sending data to AI_Service, THE ContextKeeper_Backend SHALL filter to the most recent 20 commits