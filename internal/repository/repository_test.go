package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DevAnuragT/context_keeper/internal/database"
	"github.com/DevAnuragT/context_keeper/internal/models"
	_ "github.com/lib/pq"
)

// TestRepository tests repository operations with an in-memory database
func TestRepository(t *testing.T) {
	// Skip if no database available
	db := setupTestDB(t)
	if db == nil {
		t.Skip("No test database available")
		return
	}
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	// Test repository creation
	testRepo := &models.Repository{
		Name:     "test-repo",
		FullName: "testuser/test-repo",
		Owner:    "testuser",
	}

	err := repo.CreateRepo(ctx, testRepo)
	if err != nil {
		t.Fatalf("CreateRepo failed: %v", err)
	}

	if testRepo.ID == 0 {
		t.Error("Expected repo ID to be set")
	}

	// Test getting repo by ID
	retrieved, err := repo.GetRepoByID(ctx, testRepo.ID)
	if err != nil {
		t.Fatalf("GetRepoByID failed: %v", err)
	}

	if retrieved.FullName != testRepo.FullName {
		t.Errorf("Expected full_name %s, got %s", testRepo.FullName, retrieved.FullName)
	}

	// Test getting repos by user
	repos, err := repo.GetReposByUser(ctx, "testuser")
	if err != nil {
		t.Fatalf("GetReposByUser failed: %v", err)
	}

	if len(repos) != 1 {
		t.Errorf("Expected 1 repo, got %d", len(repos))
	}
}

func TestStringListSerialization(t *testing.T) {
	// Skip if no database available
	db := setupTestDB(t)
	if db == nil {
		t.Skip("No test database available")
		return
	}
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	// Create a test repository first
	testRepo := &models.Repository{
		Name:     "test-repo",
		FullName: "testuser/test-repo",
		Owner:    "testuser",
	}
	err := repo.CreateRepo(ctx, testRepo)
	if err != nil {
		t.Fatalf("CreateRepo failed: %v", err)
	}

	// Test pull request with files and labels
	pr := &models.PullRequest{
		ID:           1,
		RepoID:       testRepo.ID,
		Number:       1,
		Title:        "Test PR",
		Body:         "Test body",
		Author:       "testuser",
		State:        "open",
		CreatedAt:    time.Now(),
		FilesChanged: models.StringList{"file1.go", "file2.go"},
		Labels:       models.StringList{"bug", "enhancement"},
	}

	err = repo.CreatePullRequest(ctx, pr)
	if err != nil {
		t.Fatalf("CreatePullRequest failed: %v", err)
	}

	// Retrieve and verify
	prs, err := repo.GetRecentPRs(ctx, testRepo.ID, 10)
	if err != nil {
		t.Fatalf("GetRecentPRs failed: %v", err)
	}

	if len(prs) != 1 {
		t.Fatalf("Expected 1 PR, got %d", len(prs))
	}

	retrievedPR := prs[0]
	if len(retrievedPR.FilesChanged) != 2 {
		t.Errorf("Expected 2 files, got %d", len(retrievedPR.FilesChanged))
	}

	if len(retrievedPR.Labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(retrievedPR.Labels))
	}

	// Verify specific values
	expectedFiles := []string{"file1.go", "file2.go"}
	for i, file := range retrievedPR.FilesChanged {
		if file != expectedFiles[i] {
			t.Errorf("Expected file %s, got %s", expectedFiles[i], file)
		}
	}
}

func TestJobOperations(t *testing.T) {
	// Skip if no database available
	db := setupTestDB(t)
	if db == nil {
		t.Skip("No test database available")
		return
	}
	defer db.Close()

	repo := New(db)
	ctx := context.Background()

	// Create a test repository first
	testRepo := &models.Repository{
		Name:     "test-repo",
		FullName: "testuser/test-repo",
		Owner:    "testuser",
	}
	err := repo.CreateRepo(ctx, testRepo)
	if err != nil {
		t.Fatalf("CreateRepo failed: %v", err)
	}

	// Create a job
	job := &models.IngestionJob{
		RepoID: testRepo.ID,
		Status: models.JobStatusPending,
	}

	err = repo.CreateJob(ctx, job)
	if err != nil {
		t.Fatalf("CreateJob failed: %v", err)
	}

	if job.ID == 0 {
		t.Error("Expected job ID to be set")
	}

	// Update job status to running
	err = repo.UpdateJobStatus(ctx, job.ID, models.JobStatusRunning, nil)
	if err != nil {
		t.Fatalf("UpdateJobStatus failed: %v", err)
	}

	// Get job and verify status
	retrievedJob, err := repo.GetJobByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("GetJobByID failed: %v", err)
	}

	if retrievedJob.Status != models.JobStatusRunning {
		t.Errorf("Expected status %s, got %s", models.JobStatusRunning, retrievedJob.Status)
	}

	if retrievedJob.StartedAt == nil {
		t.Error("Expected started_at to be set")
	}

	// Update to completed
	err = repo.UpdateJobStatus(ctx, job.ID, models.JobStatusCompleted, nil)
	if err != nil {
		t.Fatalf("UpdateJobStatus to completed failed: %v", err)
	}

	// Verify final status
	finalJob, err := repo.GetJobByID(ctx, job.ID)
	if err != nil {
		t.Fatalf("GetJobByID final failed: %v", err)
	}

	if finalJob.Status != models.JobStatusCompleted {
		t.Errorf("Expected status %s, got %s", models.JobStatusCompleted, finalJob.Status)
	}

	if finalJob.FinishedAt == nil {
		t.Error("Expected finished_at to be set")
	}
}

// setupTestDB creates a test database connection
// Returns nil if no database is available (for CI/local testing)
func setupTestDB(t *testing.T) *sql.DB {
	// Try to connect to test database
	db, err := sql.Open("postgres", "postgres://localhost/contextkeeper_test?sslmode=disable")
	if err != nil {
		t.Logf("Could not connect to test database: %v", err)
		return nil
	}

	// Test connection
	if err := db.Ping(); err != nil {
		t.Logf("Could not ping test database: %v", err)
		db.Close()
		return nil
	}

	// Clean up any existing data
	cleanupTestDB(db)

	// Run migrations
	if err := database.Migrate(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

// cleanupTestDB removes all data from test database
func cleanupTestDB(db *sql.DB) {
	tables := []string{"ingestion_jobs", "commits", "issues", "pull_requests", "repos", "schema_migrations"}
	for _, table := range tables {
		db.Exec("DELETE FROM " + table)
	}
}
