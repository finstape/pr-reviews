package team

import (
	"context"
	"testing"

	"github.com/finstape/pr-reviews/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestCreateTeam_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		team          entity.Team
		teamExists    bool
		createError   error
		expectedError error
	}{
		{
			name: "success create new team",
			team: entity.Team{
				TeamName: "new-team",
				Members: []entity.TeamMember{
					{UserID: "u1", Username: "User1", IsActive: true},
				},
			},
			teamExists:    false,
			createError:   nil,
			expectedError: nil,
		},
		{
			name: "team already exists",
			team: entity.Team{
				TeamName: "existing-team",
				Members:  []entity.TeamMember{},
			},
			teamExists:    true,
			createError:   nil,
			expectedError: entity.ErrTeamExists,
		},
		{
			name: "create error",
			team: entity.Team{
				TeamName: "error-team",
				Members:  []entity.TeamMember{},
			},
			teamExists:    false,
			createError:   entity.ErrNotFound,
			expectedError: entity.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockTeamRepo)
			uc := New(repo)
			ctx := context.Background()

			repo.On("TeamExists", ctx, tt.team.TeamName).Return(tt.teamExists, nil)

			if !tt.teamExists {
				repo.On("CreateTeam", ctx, tt.team).Return(tt.createError)
			}

			err := uc.CreateTeam(ctx, tt.team)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}

