package team

import (
	"context"
	"fmt"

	"github.com/finstape/pr-reviews/internal/entity"
	"github.com/finstape/pr-reviews/internal/repo"
)

// UseCase handles team business logic.
type UseCase struct {
	teamRepo repo.TeamRepo
}

// New creates a new Team use case instance.
func New(teamRepo repo.TeamRepo) *UseCase {
	return &UseCase{
		teamRepo: teamRepo,
	}
}

// CreateTeam creates a team with its members
func (uc *UseCase) CreateTeam(ctx context.Context, team entity.Team) error {
	// Check if team already exists
	exists, err := uc.teamRepo.TeamExists(ctx, team.TeamName)
	if err != nil {
		return fmt.Errorf("TeamUseCase - CreateTeam - TeamExists: %w", err)
	}

	if exists {
		return entity.ErrTeamExists
	}

	// Create team
	err = uc.teamRepo.CreateTeam(ctx, team)
	if err != nil {
		return fmt.Errorf("TeamUseCase - CreateTeam - CreateTeam: %w", err)
	}

	return nil
}

// GetTeam retrieves a team with its members
func (uc *UseCase) GetTeam(ctx context.Context, teamName string) (entity.Team, error) {
	team, err := uc.teamRepo.GetTeam(ctx, teamName)
	if err != nil {
		return entity.Team{}, fmt.Errorf("TeamUseCase - GetTeam - GetTeam: %w", err)
	}

	return team, nil
}

