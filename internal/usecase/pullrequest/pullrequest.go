package pullrequest

import (
	"context"
	"fmt"
	"time"

	"github.com/finstape/pr-reviews/internal/entity"
	"github.com/finstape/pr-reviews/internal/repo"
	"github.com/finstape/pr-reviews/internal/repo/persistent"
)

// UseCase handles pull request business logic.
type UseCase struct {
	prRepo   repo.PullRequestRepo
	userRepo repo.UserRepo
	teamRepo repo.TeamRepo
}

// New creates a new PullRequest use case instance.
func New(prRepo repo.PullRequestRepo, userRepo repo.UserRepo, teamRepo repo.TeamRepo) *UseCase {
	return &UseCase{
		prRepo:   prRepo,
		userRepo: userRepo,
		teamRepo: teamRepo,
	}
}

// CreatePR creates a PR and automatically assigns up to 2 reviewers from author's team
func (uc *UseCase) CreatePR(ctx context.Context, prID string, prName string, authorID string) (entity.PullRequest, error) {
	// Check if PR already exists
	exists, err := uc.prRepo.PRExists(ctx, prID)
	if err != nil {
		return entity.PullRequest{}, fmt.Errorf("PullRequestUseCase - CreatePR - PRExists: %w", err)
	}

	if exists {
		return entity.PullRequest{}, entity.ErrPRExists
	}

	// Get author to find their team
	author, err := uc.userRepo.GetUser(ctx, authorID)
	if err != nil {
		return entity.PullRequest{}, fmt.Errorf("PullRequestUseCase - CreatePR - GetUser: %w", err)
	}

	// Get active team members (excluding author)
	candidates, err := uc.userRepo.GetActiveTeamMembers(ctx, author.TeamName, authorID)
	if err != nil {
		return entity.PullRequest{}, fmt.Errorf("PullRequestUseCase - CreatePR - GetActiveTeamMembers: %w", err)
	}

	// Select up to 2 random reviewers
	reviewerIDs := persistent.SelectRandomReviewers(candidates, 2)

	// Create PR
	now := time.Now()
	pr := entity.PullRequest{
		PullRequestID:    prID,
		PullRequestName:  prName,
		AuthorID:         authorID,
		Status:           entity.PullRequestStatusOpen,
		AssignedReviewers: reviewerIDs,
		CreatedAt:        &now,
	}

	err = uc.prRepo.CreatePR(ctx, pr, reviewerIDs)
	if err != nil {
		return entity.PullRequest{}, fmt.Errorf("PullRequestUseCase - CreatePR - CreatePR: %w", err)
	}

	return pr, nil
}

// MergePR marks a PR as merged (idempotent)
func (uc *UseCase) MergePR(ctx context.Context, prID string) (entity.PullRequest, error) {
	// Get PR
	pr, err := uc.prRepo.GetPR(ctx, prID)
	if err != nil {
		return entity.PullRequest{}, fmt.Errorf("PullRequestUseCase - MergePR - GetPR: %w", err)
	}

	// If already merged, return current state (idempotent)
	if pr.Status == entity.PullRequestStatusMerged {
		return pr, nil
	}

	// Update status to MERGED
	now := time.Now()
	mergedAt := entity.Time(now)
	err = uc.prRepo.UpdatePRStatus(ctx, prID, entity.PullRequestStatusMerged, &mergedAt)
	if err != nil {
		return entity.PullRequest{}, fmt.Errorf("PullRequestUseCase - MergePR - UpdatePRStatus: %w", err)
	}

	// Get updated PR
	pr, err = uc.prRepo.GetPR(ctx, prID)
	if err != nil {
		return entity.PullRequest{}, fmt.Errorf("PullRequestUseCase - MergePR - GetPR after update: %w", err)
	}

	return pr, nil
}

// ReassignReviewer replaces one reviewer with another from the same team
func (uc *UseCase) ReassignReviewer(ctx context.Context, prID string, oldReviewerID string) (entity.PullRequest, string, error) {
	// Get PR
	pr, err := uc.prRepo.GetPR(ctx, prID)
	if err != nil {
		return entity.PullRequest{}, "", fmt.Errorf("PullRequestUseCase - ReassignReviewer - GetPR: %w", err)
	}

	// Check if PR is merged
	if pr.Status == entity.PullRequestStatusMerged {
		return entity.PullRequest{}, "", entity.ErrPRMerged
	}

	// Verify old reviewer is assigned
	reviewers := pr.AssignedReviewers
	found := false
	for _, reviewerID := range reviewers {
		if reviewerID == oldReviewerID {
			found = true
			break
		}
	}

	if !found {
		return entity.PullRequest{}, "", entity.ErrNotAssigned
	}

	// Get old reviewer to find their team
	oldReviewer, err := uc.userRepo.GetUser(ctx, oldReviewerID)
	if err != nil {
		return entity.PullRequest{}, "", fmt.Errorf("PullRequestUseCase - ReassignReviewer - GetUser old: %w", err)
	}

	// Get active team members from old reviewer's team (excluding old reviewer and already assigned reviewers)
	candidates, err := uc.userRepo.GetActiveTeamMembers(ctx, oldReviewer.TeamName, oldReviewerID)
	if err != nil {
		return entity.PullRequest{}, "", fmt.Errorf("PullRequestUseCase - ReassignReviewer - GetActiveTeamMembers: %w", err)
	}

	// Filter out already assigned reviewers
	availableCandidates := make([]entity.User, 0)
	for _, candidate := range candidates {
		isAssigned := false
		for _, reviewerID := range reviewers {
			if candidate.UserID == reviewerID {
				isAssigned = true
				break
			}
		}
		if !isAssigned {
			availableCandidates = append(availableCandidates, candidate)
		}
	}

	if len(availableCandidates) == 0 {
		return entity.PullRequest{}, "", entity.ErrNoCandidate
	}

	// Select random replacement
	selected := persistent.SelectRandomReviewers(availableCandidates, 1)
	if len(selected) == 0 {
		return entity.PullRequest{}, "", entity.ErrNoCandidate
	}

	newReviewerID := selected[0]

	// Reassign
	err = uc.prRepo.ReassignReviewer(ctx, prID, oldReviewerID, newReviewerID)
	if err != nil {
		return entity.PullRequest{}, "", fmt.Errorf("PullRequestUseCase - ReassignReviewer - ReassignReviewer: %w", err)
	}

	// Get updated PR
	pr, err = uc.prRepo.GetPR(ctx, prID)
	if err != nil {
		return entity.PullRequest{}, "", fmt.Errorf("PullRequestUseCase - ReassignReviewer - GetPR after reassign: %w", err)
	}

	return pr, newReviewerID, nil
}

