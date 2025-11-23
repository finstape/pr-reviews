package user

import (
	"context"
	"fmt"

	"github.com/finstape/pr-reviews/internal/entity"
	"github.com/finstape/pr-reviews/internal/repo"
)

// UseCase handles user business logic.
type UseCase struct {
	userRepo repo.UserRepo
}

// New creates a new User use case instance.
func New(userRepo repo.UserRepo) *UseCase {
	return &UseCase{
		userRepo: userRepo,
	}
}

// SetIsActive sets user's active status
func (uc *UseCase) SetIsActive(ctx context.Context, userID string, isActive bool) (entity.User, error) {
	// Get user first to return it
	user, err := uc.userRepo.GetUser(ctx, userID)
	if err != nil {
		return entity.User{}, fmt.Errorf("UserUseCase - SetIsActive - GetUser: %w", err)
	}

	// Update status
	err = uc.userRepo.SetIsActive(ctx, userID, isActive)
	if err != nil {
		return entity.User{}, fmt.Errorf("UserUseCase - SetIsActive - SetIsActive: %w", err)
	}

	// Update local copy
	user.IsActive = isActive

	return user, nil
}

// GetUserReviews retrieves all PRs where user is a reviewer
func (uc *UseCase) GetUserReviews(ctx context.Context, userID string) ([]entity.PullRequestShort, error) {
	// Verify user exists
	_, err := uc.userRepo.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - GetUserReviews - GetUser: %w", err)
	}

	prs, err := uc.userRepo.GetUserReviews(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("UserUseCase - GetUserReviews - GetUserReviews: %w", err)
	}

	return prs, nil
}

