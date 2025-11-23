package pullrequest

import (
	"context"
	"testing"

	"github.com/finstape/pr-reviews/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreatePR_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		prID          string
		prName        string
		authorID      string
		prExists      bool
		authorExists  bool
		candidates    []entity.User
		expectedError error
		expectedCount int
	}{
		{
			name:         "success with 2 candidates",
			prID:         "pr-1",
			prName:       "Test PR",
			authorID:     "u1",
			prExists:     false,
			authorExists: true,
			candidates: []entity.User{
				{UserID: "u2", Username: "Reviewer1", TeamName: "team1", IsActive: true},
				{UserID: "u3", Username: "Reviewer2", TeamName: "team1", IsActive: true},
			},
			expectedError: nil,
			expectedCount: 2,
		},
		{
			name:         "success with 1 candidate",
			prID:         "pr-2",
			prName:       "Test PR 2",
			authorID:     "u1",
			prExists:     false,
			authorExists: true,
			candidates: []entity.User{
				{UserID: "u2", Username: "Reviewer1", TeamName: "team1", IsActive: true},
			},
			expectedError: nil,
			expectedCount: 1,
		},
		{
			name:         "success with no candidates",
			prID:         "pr-3",
			prName:       "Test PR 3",
			authorID:     "u1",
			prExists:     false,
			authorExists: true,
			candidates:   []entity.User{},
			expectedError: nil,
			expectedCount: 0,
		},
		{
			name:          "pr already exists",
			prID:          "pr-4",
			prName:        "Test PR 4",
			authorID:      "u1",
			prExists:      true,
			authorExists:  true,
			candidates:    []entity.User{},
			expectedError: entity.ErrPRExists,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prRepo := new(mockPRRepo)
			userRepo := new(mockUserRepo)
			teamRepo := new(mockTeamRepo)

			uc := New(prRepo, userRepo, teamRepo)
			ctx := context.Background()

			author := entity.User{
				UserID:   tt.authorID,
				Username: "Author",
				TeamName: "team1",
				IsActive: true,
			}

			prRepo.On("PRExists", ctx, tt.prID).Return(tt.prExists, nil)

			if !tt.prExists {
				userRepo.On("GetUser", ctx, tt.authorID).Return(author, nil)
				userRepo.On("GetActiveTeamMembers", ctx, "team1", tt.authorID).Return(tt.candidates, nil)
				prRepo.On("CreatePR", ctx, mock.Anything, mock.Anything).Return(nil)
			}

			pr, err := uc.CreatePR(ctx, tt.prID, tt.prName, tt.authorID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.prID, pr.PullRequestID)
				assert.Len(t, pr.AssignedReviewers, tt.expectedCount)
			}

			prRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
		})
	}
}

