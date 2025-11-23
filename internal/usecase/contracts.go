// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"github.com/finstape/pr-reviews/internal/entity"
)

//go:generate mockgen -source=contracts.go -destination=./mocks_usecase_test.go -package=usecase_test

type (
	// Team defines team use case interface.
	Team interface {
		CreateTeam(ctx context.Context, team entity.Team) error
		GetTeam(ctx context.Context, teamName string) (entity.Team, error)
	}

	// User defines user use case interface.
	User interface {
		SetIsActive(ctx context.Context, userID string, isActive bool) (entity.User, error)
		GetUserReviews(ctx context.Context, userID string) ([]entity.PullRequestShort, error)
	}

	// PullRequest defines pull request use case interface.
	PullRequest interface {
		CreatePR(ctx context.Context, prID string, prName string, authorID string) (entity.PullRequest, error)
		MergePR(ctx context.Context, prID string) (entity.PullRequest, error)
		ReassignReviewer(ctx context.Context, prID string, oldReviewerID string) (entity.PullRequest, string, error)
	}
)

