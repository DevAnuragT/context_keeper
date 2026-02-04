# Implementation Plan: ContextKeeper Go Backend

## Overview

This implementation plan breaks down the ContextKeeper Go Backend into discrete, incremental coding tasks. Each task builds on previous work, focusing on core functionality first, then adding background processing, external integrations, and comprehensive testing. The approach emphasizes clean architecture with clear separation between HTTP handlers, business logic, and data access layers.

## Tasks

- [x] 1. Set up project structure and core interfaces
  - Create Go module with standard project layout (cmd/, internal/, pkg/)
  - Define core domain models and interfaces for all services
  - Set up PostgreSQL database connection and migration system
  - Create basic HTTP server with graceful shutdown
  - _Requirements: 9.1, 9.2, 9.5_

- [x] 2. Implement database schema and repository layer
  - [x] 2.1 Create database migration files for all tables
    - Write SQL migrations for repos, pull_requests, issues, commits, ingestion_jobs tables
    - Include all indexes and foreign key constraints
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 3.6_
  
  - [x] 2.2 Implement repository interfaces and SQL operations
    - Write repository implementations for all CRUD operations
    - Implement data filtering queries for AI service (recent 10/10/20 limits)
    - Handle JSONB serialization for files_changed and labels
    - _Requirements: 3.4, 3.5, 6.1, 6.2, 6.3_
  
  - [x]* 2.3 Write property test for structured array serialization
    - **Property 5: Structured Array Serialization**
    - **Validates: Requirements 3.4, 3.5**

- [x] 3. Implement GitHub OAuth authentication service
  - [x] 3.1 Create GitHub OAuth flow handlers
    - Implement OAuth initiation with correct scopes (public_repo, read:user, user:email)
    - Handle OAuth callback and token exchange
    - _Requirements: 1.1, 1.2_
  
  - [x] 3.2 Implement JWT token generation and validation
    - Create JWT tokens with user ID and GitHub token claims
    - Implement JWT validation middleware for API endpoints
    - _Requirements: 1.3, 1.5_
  
  - [x]* 3.3 Write property test for OAuth scope consistency
    - **Property 1: OAuth Scope Consistency**
    - **Validates: Requirements 1.1**
  
  - [x]* 3.4 Write property test for JWT authentication round trip
    - **Property 2: JWT Authentication Round Trip**
    - **Validates: Requirements 1.3, 1.5**

- [x] 4. Implement GitHub API client service
  - [x] 4.1 Create GitHub API client with rate limiting
    - Implement HTTP client for GitHub REST API
    - Add sequential processing with single retry logic
    - Handle rate limiting and API errors gracefully
    - _Requirements: 2.5, 2.6, 2.7_
  
  - [x] 4.2 Implement repository data extraction methods
    - Create methods to fetch PRs, issues, and commits with limits (50/50/100)
    - Extract only specified metadata fields for each data type
    - Handle pagination and data filtering
    - _Requirements: 2.2, 2.3, 2.4, 3.1, 3.2, 3.3_
  
  - [x]* 4.3 Write property test for repository data extraction limits
    - **Property 3: Repository Data Extraction Limits**
    - **Validates: Requirements 2.2, 2.3, 2.4, 10.1, 10.2, 10.3**
  
  - [x]* 4.4 Write property test for metadata field extraction
    - **Property 4: Repository Metadata Field Extraction**
    - **Validates: Requirements 3.1, 3.2, 3.3**

- [x] 5. Checkpoint - Core services integration
  - Ensure all tests pass, ask the user if questions arise.

- [x] 6. Implement background job processing service
  - [x] 6.1 Create ingestion job management
    - Implement job creation, status tracking, and persistence
    - Create goroutine-based job processor with proper error handling
    - _Requirements: 4.1, 4.2_
  
  - [x] 6.2 Implement job lifecycle and status transitions
    - Handle job state transitions (pending → running → completed/partial/failed)
    - Store timestamps and error messages for each status change
    - _Requirements: 4.3, 4.4, 4.5, 4.6_
  
  - [x]* 6.3 Write property test for ingestion job lifecycle
    - **Property 6: Ingestion Job Lifecycle**
    - **Validates: Requirements 4.1, 4.2, 4.3, 4.4, 4.5, 4.6**

- [x] 7. Implement AI service integration
  - [x] 7.1 Create AI service client with timeout handling
    - Implement HTTP client for Python AI service with 30-second timeout
    - Handle timeout errors and service failures without retries
    - _Requirements: 6.4, 6.5, 8.3_
  
  - [x] 7.2 Implement context query processing
    - Filter repository data to recent limits (10/10/20) for AI service
    - Forward AI responses without caching or semantic processing
    - _Requirements: 6.1, 6.2, 6.3, 6.6, 6.7_
  
  - [x]* 7.3 Write property test for AI context payload filtering
    - **Property 7: AI Context Payload Filtering**
    - **Validates: Requirements 6.1, 6.2, 6.3, 6.7, 10.4, 10.5, 10.6**
  
  - [x]* 7.4 Write property test for AI service timeout enforcement
    - **Property 8: AI Service Timeout Enforcement**
    - **Validates: Requirements 6.4, 6.5**

- [x] 8. Implement HTTP API endpoints and middleware
  - [x] 8.1 Create authentication middleware and error handling
    - Implement JWT validation middleware for all protected endpoints
    - Create consistent error response formatting with appropriate HTTP status codes
    - _Requirements: 5.6, 8.2, 8.4, 8.5_
  
  - [x] 8.2 Implement REST API endpoints
    - POST /api/auth/github (OAuth callback)
    - GET /api/repos (list ingested repositories)
    - POST /api/repos/ingest (trigger ingestion)
    - GET /api/repos/{id}/status (ingestion status)
    - POST /api/context/query (unified context queries)
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_
  
  - [x]* 8.3 Write property test for API authentication enforcement
    - **Property 9: Backend API Authentication Enforcement**
    - **Validates: Requirements 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 8.4**
  
  - [x]* 8.4 Write property test for fail-fast error handling
    - **Property 10: Fail-Fast Error Handling**
    - **Validates: Requirements 2.6, 8.1, 8.2, 8.3, 8.4, 8.5**

- [x] 9. Integration and service wiring
  - [x] 9.1 Wire all services together in main application
    - Connect HTTP handlers to service layer
    - Initialize database connections and migrations
    - Set up dependency injection and service configuration
    - _Requirements: 9.1, 9.3, 9.4_
  
  - [x] 9.2 Implement application configuration and environment handling
    - Add configuration for database, GitHub OAuth, AI service endpoints
    - Support environment variables and local development setup
    - _Requirements: 9.3, 9.4_
  
  - [x]* 9.3 Write integration tests for complete API flows
    - Test end-to-end OAuth → ingestion → context query flows
    - Verify database persistence and job processing
    - _Requirements: All requirements integration_

- [x] 10. Final checkpoint and deployment preparation
  - [x] 10.1 Create Docker configuration
    - Write Dockerfile for single binary deployment
    - Create docker-compose for local development with PostgreSQL
    - _Requirements: 9.3_
  
  - [x] 10.2 Add application logging and monitoring
    - Implement structured logging for all operations
    - Add health check endpoint for deployment monitoring
    - _Requirements: 8.1, 8.2_
  
  - [x]* 10.3 Write comprehensive system tests
    - Test complete system behavior with real database
    - Verify all error handling paths and edge cases
    - _Requirements: All requirements validation_

- [x] 11. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation and allow for user feedback
- Property tests validate universal correctness properties from the design document
- Integration tests verify end-to-end system behavior
- The implementation follows clean architecture principles with clear service boundaries