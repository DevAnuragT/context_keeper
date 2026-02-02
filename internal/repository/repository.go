package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/DevAnuragT/context_keeper/internal/models"
	"github.com/DevAnuragT/context_keeper/internal/services"
)

// Repository implements the RepositoryStore interface
type Repository struct {
	db *sql.DB
}

// New creates a new repository instance
func New(db *sql.DB) services.RepositoryStore {
	return &Repository{db: db}
}

// Repository operations
func (r *Repository) CreateRepo(ctx context.Context, repo *models.Repository) error {
	query := `
		INSERT INTO repos (name, full_name, owner, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (full_name) DO UPDATE SET
			name = EXCLUDED.name,
			owner = EXCLUDED.owner,
			updated_at = EXCLUDED.updated_at
		RETURNING id`

	now := time.Now()
	repo.CreatedAt = now
	repo.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query, repo.Name, repo.FullName, repo.Owner, repo.CreatedAt, repo.UpdatedAt).Scan(&repo.ID)
	return err
}

func (r *Repository) GetReposByUser(ctx context.Context, userID string) ([]models.Repository, error) {
	query := `
		SELECT id, name, full_name, owner, created_at, updated_at
		FROM repos
		WHERE owner = $1
		ORDER BY updated_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []models.Repository
	for rows.Next() {
		var repo models.Repository
		err := rows.Scan(&repo.ID, &repo.Name, &repo.FullName, &repo.Owner, &repo.CreatedAt, &repo.UpdatedAt)
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}

	return repos, rows.Err()
}

func (r *Repository) GetRepoByID(ctx context.Context, repoID int64) (*models.Repository, error) {
	query := `
		SELECT id, name, full_name, owner, created_at, updated_at
		FROM repos
		WHERE id = $1`

	var repo models.Repository
	err := r.db.QueryRowContext(ctx, query, repoID).Scan(
		&repo.ID, &repo.Name, &repo.FullName, &repo.Owner, &repo.CreatedAt, &repo.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &repo, nil
}

// Pull request operations
func (r *Repository) CreatePullRequest(ctx context.Context, pr *models.PullRequest) error {
	query := `
		INSERT INTO pull_requests (id, repo_id, number, title, body, author, state, created_at, merged_at, files_changed, labels)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (repo_id, number) DO UPDATE SET
			title = EXCLUDED.title,
			body = EXCLUDED.body,
			state = EXCLUDED.state,
			merged_at = EXCLUDED.merged_at,
			files_changed = EXCLUDED.files_changed,
			labels = EXCLUDED.labels`

	_, err := r.db.ExecContext(ctx, query,
		pr.ID, pr.RepoID, pr.Number, pr.Title, pr.Body, pr.Author, pr.State,
		pr.CreatedAt, pr.MergedAt, pr.FilesChanged, pr.Labels)
	return err
}

func (r *Repository) GetRecentPRs(ctx context.Context, repoID int64, limit int) ([]models.PullRequest, error) {
	query := `
		SELECT id, repo_id, number, title, body, author, state, created_at, merged_at, files_changed, labels
		FROM pull_requests
		WHERE repo_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, repoID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []models.PullRequest
	for rows.Next() {
		var pr models.PullRequest
		err := rows.Scan(&pr.ID, &pr.RepoID, &pr.Number, &pr.Title, &pr.Body, &pr.Author,
			&pr.State, &pr.CreatedAt, &pr.MergedAt, &pr.FilesChanged, &pr.Labels)
		if err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}

	return prs, rows.Err()
}

// Issue operations
func (r *Repository) CreateIssue(ctx context.Context, issue *models.Issue) error {
	query := `
		INSERT INTO issues (id, repo_id, title, body, author, state, created_at, closed_at, labels)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			body = EXCLUDED.body,
			state = EXCLUDED.state,
			closed_at = EXCLUDED.closed_at,
			labels = EXCLUDED.labels`

	_, err := r.db.ExecContext(ctx, query,
		issue.ID, issue.RepoID, issue.Title, issue.Body, issue.Author,
		issue.State, issue.CreatedAt, issue.ClosedAt, issue.Labels)
	return err
}

func (r *Repository) GetRecentIssues(ctx context.Context, repoID int64, limit int) ([]models.Issue, error) {
	query := `
		SELECT id, repo_id, title, body, author, state, created_at, closed_at, labels
		FROM issues
		WHERE repo_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, repoID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []models.Issue
	for rows.Next() {
		var issue models.Issue
		err := rows.Scan(&issue.ID, &issue.RepoID, &issue.Title, &issue.Body, &issue.Author,
			&issue.State, &issue.CreatedAt, &issue.ClosedAt, &issue.Labels)
		if err != nil {
			return nil, err
		}
		issues = append(issues, issue)
	}

	return issues, rows.Err()
}

// Commit operations
func (r *Repository) CreateCommit(ctx context.Context, commit *models.Commit) error {
	query := `
		INSERT INTO commits (sha, repo_id, message, author, created_at, files_changed)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (sha) DO UPDATE SET
			message = EXCLUDED.message,
			author = EXCLUDED.author,
			created_at = EXCLUDED.created_at,
			files_changed = EXCLUDED.files_changed`

	_, err := r.db.ExecContext(ctx, query,
		commit.SHA, commit.RepoID, commit.Message, commit.Author, commit.CreatedAt, commit.FilesChanged)
	return err
}

func (r *Repository) GetRecentCommits(ctx context.Context, repoID int64, limit int) ([]models.Commit, error) {
	query := `
		SELECT sha, repo_id, message, author, created_at, files_changed
		FROM commits
		WHERE repo_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, repoID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commits []models.Commit
	for rows.Next() {
		var commit models.Commit
		err := rows.Scan(&commit.SHA, &commit.RepoID, &commit.Message, &commit.Author,
			&commit.CreatedAt, &commit.FilesChanged)
		if err != nil {
			return nil, err
		}
		commits = append(commits, commit)
	}

	return commits, rows.Err()
}

// Job operations
func (r *Repository) CreateJob(ctx context.Context, job *models.IngestionJob) error {
	query := `
		INSERT INTO ingestion_jobs (repo_id, status, started_at, finished_at, error_message)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query, job.RepoID, job.Status, job.StartedAt, job.FinishedAt, job.ErrorMsg).Scan(&job.ID)
	return err
}

func (r *Repository) UpdateJobStatus(ctx context.Context, jobID int64, status models.JobStatus, errorMsg *string) error {
	var query string
	var args []interface{}

	if status == models.JobStatusRunning {
		query = `UPDATE ingestion_jobs SET status = $1, started_at = $2 WHERE id = $3`
		args = []interface{}{status, time.Now(), jobID}
	} else {
		query = `UPDATE ingestion_jobs SET status = $1, finished_at = $2, error_message = $3 WHERE id = $4`
		args = []interface{}{status, time.Now(), errorMsg, jobID}
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *Repository) GetJobByID(ctx context.Context, jobID int64) (*models.IngestionJob, error) {
	query := `
		SELECT id, repo_id, status, started_at, finished_at, error_message
		FROM ingestion_jobs
		WHERE id = $1`

	var job models.IngestionJob
	err := r.db.QueryRowContext(ctx, query, jobID).Scan(
		&job.ID, &job.RepoID, &job.Status, &job.StartedAt, &job.FinishedAt, &job.ErrorMsg)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (r *Repository) GetJobsByRepo(ctx context.Context, repoID int64) ([]models.IngestionJob, error) {
	query := `
		SELECT id, repo_id, status, started_at, finished_at, error_message
		FROM ingestion_jobs
		WHERE repo_id = $1
		ORDER BY id DESC`

	rows, err := r.db.QueryContext(ctx, query, repoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []models.IngestionJob
	for rows.Next() {
		var job models.IngestionJob
		err := rows.Scan(&job.ID, &job.RepoID, &job.Status, &job.StartedAt, &job.FinishedAt, &job.ErrorMsg)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}
