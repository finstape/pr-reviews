// Package repo implements application outer layer logic. Each logic group in own file.
package repo

import (
	"context"

	"github.com/finstape/pr-reviews/internal/entity"
)

//go:generate mockgen -source=contracts.go -destination=../usecase/mocks_repo_test.go -package=usecase_test

type (
	// TeamRepo defines team repository interface.
	TeamRepo interface {
		CreateTeam(ctx context.Context, team entity.Team) error
		GetTeam(ctx context.Context, teamName string) (entity.Team, error)
		TeamExists(ctx context.Context, teamName string) (bool, error)
	}

	// UserRepo defines user repository interface.
	UserRepo interface {
		CreateOrUpdateUser(ctx context.Context, user entity.User) error
		GetUser(ctx context.Context, userID string) (entity.User, error)
		SetIsActive(ctx context.Context, userID string, isActive bool) error
		GetActiveTeamMembers(ctx context.Context, teamName string, excludeUserID string) ([]entity.User, error)
		GetUserReviews(ctx context.Context, userID string) ([]entity.PullRequestShort, error)
	}

	// PullRequestRepo defines pull request repository interface.
	PullRequestRepo interface {
		CreatePR(ctx context.Context, pr entity.PullRequest, reviewerIDs []string) error
		GetPR(ctx context.Context, prID string) (entity.PullRequest, error)
		PRExists(ctx context.Context, prID string) (bool, error)
		UpdatePRStatus(ctx context.Context, prID string, status entity.PullRequestStatus, mergedAt *entity.Time) error
		GetPRReviewers(ctx context.Context, prID string) ([]string, error)
		ReassignReviewer(ctx context.Context, prID string, oldReviewerID string, newReviewerID string) error
		GetPRsByReviewer(ctx context.Context, reviewerID string) ([]entity.PullRequestShort, error)
	}
)

