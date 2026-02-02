package database

import (
	"database/sql"
	"fmt"
)

// Migration represents a database migration
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// migrations contains all database migrations in order
var migrations = []Migration{
	{
		Version: 1,
		Name:    "create_repos_table",
		SQL: `
			CREATE TABLE IF NOT EXISTS repos (
				id BIGSERIAL PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				full_name VARCHAR(255) NOT NULL UNIQUE,
				owner VARCHAR(255) NOT NULL,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			);
			
			CREATE INDEX IF NOT EXISTS idx_repos_owner ON repos(owner);
		`,
	},
	{
		Version: 2,
		Name:    "create_pull_requests_table",
		SQL: `
			CREATE TABLE IF NOT EXISTS pull_requests (
				id BIGINT PRIMARY KEY,
				repo_id BIGINT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
				number INTEGER NOT NULL,
				title TEXT NOT NULL,
				body TEXT,
				author VARCHAR(255) NOT NULL,
				state VARCHAR(50) NOT NULL,
				created_at TIMESTAMP WITH TIME ZONE NOT NULL,
				merged_at TIMESTAMP WITH TIME ZONE,
				files_changed JSONB,
				labels JSONB,
				UNIQUE(repo_id, number)
			);
			
			CREATE INDEX IF NOT EXISTS idx_pull_requests_repo_created ON pull_requests(repo_id, created_at DESC);
			CREATE INDEX IF NOT EXISTS idx_pull_requests_repo_author ON pull_requests(repo_id, author);
		`,
	},
	{
		Version: 3,
		Name:    "create_issues_table",
		SQL: `
			CREATE TABLE IF NOT EXISTS issues (
				id BIGINT PRIMARY KEY,
				repo_id BIGINT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
				title TEXT NOT NULL,
				body TEXT,
				author VARCHAR(255) NOT NULL,
				state VARCHAR(50) NOT NULL,
				created_at TIMESTAMP WITH TIME ZONE NOT NULL,
				closed_at TIMESTAMP WITH TIME ZONE,
				labels JSONB
			);
			
			CREATE INDEX IF NOT EXISTS idx_issues_repo_created ON issues(repo_id, created_at DESC);
			CREATE INDEX IF NOT EXISTS idx_issues_repo_author ON issues(repo_id, author);
		`,
	},
	{
		Version: 4,
		Name:    "create_commits_table",
		SQL: `
			CREATE TABLE IF NOT EXISTS commits (
				sha VARCHAR(40) PRIMARY KEY,
				repo_id BIGINT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
				message TEXT NOT NULL,
				author VARCHAR(255) NOT NULL,
				created_at TIMESTAMP WITH TIME ZONE NOT NULL,
				files_changed JSONB
			);
			
			CREATE INDEX IF NOT EXISTS idx_commits_repo_created ON commits(repo_id, created_at DESC);
			CREATE INDEX IF NOT EXISTS idx_commits_repo_author ON commits(repo_id, author);
		`,
	},
	{
		Version: 5,
		Name:    "create_ingestion_jobs_table",
		SQL: `
			CREATE TABLE IF NOT EXISTS ingestion_jobs (
				id BIGSERIAL PRIMARY KEY,
				repo_id BIGINT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
				status VARCHAR(50) NOT NULL DEFAULT 'pending',
				started_at TIMESTAMP WITH TIME ZONE,
				finished_at TIMESTAMP WITH TIME ZONE,
				error_message TEXT
			);
			
			CREATE INDEX IF NOT EXISTS idx_ingestion_jobs_repo ON ingestion_jobs(repo_id);
		`,
	},
	{
		Version: 6,
		Name:    "create_schema_migrations_table",
		SQL: `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version INTEGER PRIMARY KEY,
				applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
			);
		`,
	},
}

// Migrate runs all pending migrations
func Migrate(db *sql.DB) error {
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current version
	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Run pending migrations
	for _, migration := range migrations {
		if migration.Version <= currentVersion {
			continue
		}

		if err := runMigration(db, migration); err != nil {
			return fmt.Errorf("failed to run migration %d (%s): %w", migration.Version, migration.Name, err)
		}
	}

	return nil
}

func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`
	_, err := db.Exec(query)
	return err
}

func getCurrentVersion(db *sql.DB) (int, error) {
	var version int
	query := "SELECT COALESCE(MAX(version), 0) FROM schema_migrations"
	err := db.QueryRow(query).Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

func runMigration(db *sql.DB, migration Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(migration.SQL); err != nil {
		return err
	}

	// Record migration
	if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", migration.Version); err != nil {
		return err
	}

	return tx.Commit()
}