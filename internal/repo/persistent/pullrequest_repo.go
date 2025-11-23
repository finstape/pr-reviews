package persistent

import (
	"context"
	"fmt"
	"time"

	"github.com/finstape/pr-reviews/internal/entity"
	"github.com/finstape/pr-reviews/pkg/postgres"
)

// PullRequestRepo handles pull request data persistence.
type PullRequestRepo struct {
	*postgres.Postgres
}

// NewPullRequestRepo creates a new PullRequestRepo instance.
func NewPullRequestRepo(pg *postgres.Postgres) *PullRequestRepo {
	return &PullRequestRepo{pg}
}

// CreatePR creates a PR and assigns reviewers
func (r *PullRequestRepo) CreatePR(ctx context.Context, pr entity.PullRequest, reviewerIDs []string) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("PullRequestRepo - CreatePR - Begin: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert PR
	var createdAt *time.Time
	if pr.CreatedAt != nil {
		createdAt = pr.CreatedAt
	} else {
		now := time.Now()
		createdAt = &now
	}

	sql, args, err := r.Builder.
		Insert("pull_requests").
		Columns("pull_request_id", "pull_request_name", "author_id", "status", "created_at").
		Values(pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status, createdAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("PullRequestRepo - CreatePR - BuildInsert: %w", err)
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("PullRequestRepo - CreatePR - Exec PR: %w", err)
	}

	// Insert reviewers
	for _, reviewerID := range reviewerIDs {
		sql, args, err := r.Builder.
			Insert("pr_reviewers").
			Columns("pull_request_id", "reviewer_id").
			Values(pr.PullRequestID, reviewerID).
			ToSql()
		if err != nil {
			return fmt.Errorf("PullRequestRepo - CreatePR - BuildInsert reviewer: %w", err)
		}

		_, err = tx.Exec(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("PullRequestRepo - CreatePR - Exec reviewer: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("PullRequestRepo - CreatePR - Commit: %w", err)
	}

	return nil
}

// GetPR retrieves a PR with its reviewers
func (r *PullRequestRepo) GetPR(ctx context.Context, prID string) (entity.PullRequest, error) {
	// Get PR
	sql, args, err := r.Builder.
		Select("pull_request_id", "pull_request_name", "author_id", "status", "created_at", "merged_at").
		From("pull_requests").
		Where("pull_request_id = ?", prID).
		ToSql()
	if err != nil {
		return entity.PullRequest{}, fmt.Errorf("PullRequestRepo - GetPR - BuildSelect: %w", err)
	}

	var pr entity.PullRequest
	var createdAt, mergedAt *time.Time
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&createdAt,
		&mergedAt,
	)
	if err != nil {
		return entity.PullRequest{}, fmt.Errorf("PullRequestRepo - GetPR - Scan: %w", err)
	}

	pr.CreatedAt = createdAt
	pr.MergedAt = mergedAt

	// Get reviewers
	reviewers, err := r.GetPRReviewers(ctx, prID)
	if err != nil {
		return entity.PullRequest{}, fmt.Errorf("PullRequestRepo - GetPR - GetPRReviewers: %w", err)
	}

	pr.AssignedReviewers = reviewers

	return pr, nil
}

// PRExists checks if a PR exists
func (r *PullRequestRepo) PRExists(ctx context.Context, prID string) (bool, error) {
	sql, args, err := r.Builder.
		Select("1").
		From("pull_requests").
		Where("pull_request_id = ?", prID).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("PullRequestRepo - PRExists - BuildSelect: %w", err)
	}

	var exists int
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(&exists)
	if err != nil {
		return false, nil
	}

	return exists == 1, nil
}

// UpdatePRStatus updates PR status
func (r *PullRequestRepo) UpdatePRStatus(ctx context.Context, prID string, status entity.PullRequestStatus, mergedAt *entity.Time) error {
	builder := r.Builder.
		Update("pull_requests").
		Set("status", status).
		Where("pull_request_id = ?", prID)

	if mergedAt != nil {
		builder = builder.Set("merged_at", *mergedAt)
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("PullRequestRepo - UpdatePRStatus - BuildUpdate: %w", err)
	}

	result, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("PullRequestRepo - UpdatePRStatus - Exec: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrNotFound
	}

	return nil
}

// GetPRReviewers retrieves reviewer IDs for a PR
func (r *PullRequestRepo) GetPRReviewers(ctx context.Context, prID string) ([]string, error) {
	sql, args, err := r.Builder.
		Select("reviewer_id").
		From("pr_reviewers").
		Where("pull_request_id = ?", prID).
		OrderBy("reviewer_id").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("PullRequestRepo - GetPRReviewers - BuildSelect: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("PullRequestRepo - GetPRReviewers - Query: %w", err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, fmt.Errorf("PullRequestRepo - GetPRReviewers - Scan: %w", err)
		}
		reviewers = append(reviewers, reviewerID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("PullRequestRepo - GetPRReviewers - RowsErr: %w", err)
	}

	return reviewers, nil
}

// ReassignReviewer replaces one reviewer with another
func (r *PullRequestRepo) ReassignReviewer(ctx context.Context, prID string, oldReviewerID string, newReviewerID string) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("PullRequestRepo - ReassignReviewer - Begin: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete old reviewer
	sql, args, err := r.Builder.
		Delete("pr_reviewers").
		Where("pull_request_id = ?", prID).
		Where("reviewer_id = ?", oldReviewerID).
		ToSql()
	if err != nil {
		return fmt.Errorf("PullRequestRepo - ReassignReviewer - BuildDelete: %w", err)
	}

	result, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("PullRequestRepo - ReassignReviewer - Exec delete: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrNotAssigned
	}

	// Insert new reviewer
	sql, args, err = r.Builder.
		Insert("pr_reviewers").
		Columns("pull_request_id", "reviewer_id").
		Values(prID, newReviewerID).
		ToSql()
	if err != nil {
		return fmt.Errorf("PullRequestRepo - ReassignReviewer - BuildInsert: %w", err)
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("PullRequestRepo - ReassignReviewer - Exec insert: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("PullRequestRepo - ReassignReviewer - Commit: %w", err)
	}

	return nil
}

// GetPRsByReviewer retrieves all PRs where user is a reviewer
func (r *PullRequestRepo) GetPRsByReviewer(ctx context.Context, reviewerID string) ([]entity.PullRequestShort, error) {
	sql, args, err := r.Builder.
		Select("pr.pull_request_id", "pr.pull_request_name", "pr.author_id", "pr.status").
		From("pull_requests pr").
		Join("pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id").
		Where("prr.reviewer_id = ?", reviewerID).
		OrderBy("pr.created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("PullRequestRepo - GetPRsByReviewer - BuildSelect: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("PullRequestRepo - GetPRsByReviewer - Query: %w", err)
	}
	defer rows.Close()

	var prs []entity.PullRequestShort
	for rows.Next() {
		var pr entity.PullRequestShort
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, fmt.Errorf("PullRequestRepo - GetPRsByReviewer - Scan: %w", err)
		}
		prs = append(prs, pr)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("PullRequestRepo - GetPRsByReviewer - RowsErr: %w", err)
	}

	return prs, nil
}

// SelectRandomReviewers selects up to maxCount reviewers from the candidates list.
// Currently uses simple selection (first N candidates). For production, consider using crypto/rand for true randomness.
func SelectRandomReviewers(candidates []entity.User, maxCount int) []string {
	if len(candidates) == 0 {
		return []string{}
	}

	if maxCount <= 0 {
		maxCount = 2
	}

	count := len(candidates)
	if count > maxCount {
		count = maxCount
	}

	// Simple selection: take first count candidates
	reviewerIDs := make([]string, 0, count)
	for i := 0; i < count; i++ {
		reviewerIDs = append(reviewerIDs, candidates[i].UserID)
	}

	return reviewerIDs
}

