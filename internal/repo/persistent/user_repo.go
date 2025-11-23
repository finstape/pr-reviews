package persistent

import (
	"context"
	"fmt"
	"time"

	"github.com/finstape/pr-reviews/internal/entity"
	"github.com/finstape/pr-reviews/pkg/postgres"
)

// UserRepo handles user data persistence.
type UserRepo struct {
	*postgres.Postgres
}

// NewUserRepo creates a new UserRepo instance.
func NewUserRepo(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{pg}
}

// CreateOrUpdateUser creates or updates a user
func (r *UserRepo) CreateOrUpdateUser(ctx context.Context, user entity.User) error {
	sql, args, err := r.Builder.
		Insert("users").
		Columns("user_id", "username", "team_name", "is_active").
		Values(user.UserID, user.Username, user.TeamName, user.IsActive).
		Suffix("ON CONFLICT (user_id) DO UPDATE SET username = EXCLUDED.username, team_name = EXCLUDED.team_name, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP").
		ToSql()
	if err != nil {
		return fmt.Errorf("UserRepo - CreateOrUpdateUser - BuildInsert: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserRepo - CreateOrUpdateUser - Exec: %w", err)
	}

	return nil
}

// GetUser retrieves a user by ID
func (r *UserRepo) GetUser(ctx context.Context, userID string) (entity.User, error) {
	sql, args, err := r.Builder.
		Select("user_id", "username", "team_name", "is_active").
		From("users").
		Where("user_id = ?", userID).
		ToSql()
	if err != nil {
		return entity.User{}, fmt.Errorf("UserRepo - GetUser - BuildSelect: %w", err)
	}

	var user entity.User
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)
	if err != nil {
		return entity.User{}, fmt.Errorf("UserRepo - GetUser - Scan: %w", err)
	}

	return user, nil
}

// SetIsActive updates user's active status
func (r *UserRepo) SetIsActive(ctx context.Context, userID string, isActive bool) error {
	sql, args, err := r.Builder.
		Update("users").
		Set("is_active", isActive).
		Set("updated_at", time.Now()).
		Where("user_id = ?", userID).
		ToSql()
	if err != nil {
		return fmt.Errorf("UserRepo - SetIsActive - BuildUpdate: %w", err)
	}

	result, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserRepo - SetIsActive - Exec: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrNotFound
	}

	return nil
}

// GetActiveTeamMembers retrieves active team members excluding a specific user
func (r *UserRepo) GetActiveTeamMembers(ctx context.Context, teamName string, excludeUserID string) ([]entity.User, error) {
	builder := r.Builder.
		Select("user_id", "username", "team_name", "is_active").
		From("users").
		Where("team_name = ?", teamName).
		Where("is_active = ?", true)

	if excludeUserID != "" {
		builder = builder.Where("user_id != ?", excludeUserID)
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("UserRepo - GetActiveTeamMembers - BuildSelect: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("UserRepo - GetActiveTeamMembers - Query: %w", err)
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, fmt.Errorf("UserRepo - GetActiveTeamMembers - Scan: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("UserRepo - GetActiveTeamMembers - RowsErr: %w", err)
	}

	return users, nil
}

// GetUserReviews retrieves all PRs where user is a reviewer
func (r *UserRepo) GetUserReviews(ctx context.Context, userID string) ([]entity.PullRequestShort, error) {
	sql, args, err := r.Builder.
		Select("pr.pull_request_id", "pr.pull_request_name", "pr.author_id", "pr.status").
		From("pull_requests pr").
		Join("pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id").
		Where("prr.reviewer_id = ?", userID).
		OrderBy("pr.created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("UserRepo - GetUserReviews - BuildSelect: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("UserRepo - GetUserReviews - Query: %w", err)
	}
	defer rows.Close()

	var prs []entity.PullRequestShort
	for rows.Next() {
		var pr entity.PullRequestShort
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, fmt.Errorf("UserRepo - GetUserReviews - Scan: %w", err)
		}
		prs = append(prs, pr)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("UserRepo - GetUserReviews - RowsErr: %w", err)
	}

	return prs, nil
}

