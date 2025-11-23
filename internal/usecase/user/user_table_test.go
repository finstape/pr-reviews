package user

import (
	"context"
	"testing"

	"github.com/finstape/pr-reviews/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestSetIsActive_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		isActive      bool
		userExists    bool
		updateError   error
		expectedError error
		expectedActive bool
	}{
		{
			name:          "success set to false",
			userID:        "u1",
			isActive:      false,
			userExists:    true,
			updateError:   nil,
			expectedError: nil,
			expectedActive: false,
		},
		{
			name:          "success set to true",
			userID:        "u1",
			isActive:      true,
			userExists:    true,
			updateError:   nil,
			expectedError: nil,
			expectedActive: true,
		},
		{
			name:          "user not found",
			userID:        "u99",
			isActive:      false,
			userExists:    false,
			updateError:   nil,
			expectedError: entity.ErrNotFound,
			expectedActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockUserRepo)
			uc := New(repo)
			ctx := context.Background()

			if tt.userExists {
				user := entity.User{
					UserID:   tt.userID,
					Username: "Test User",
					TeamName: "team1",
					IsActive: !tt.isActive, // Opposite of what we're setting
				}
				repo.On("GetUser", ctx, tt.userID).Return(user, nil)
				repo.On("SetIsActive", ctx, tt.userID, tt.isActive).Return(tt.updateError)
			} else {
				repo.On("GetUser", ctx, tt.userID).Return(entity.User{}, entity.ErrNotFound)
			}

			result, err := uc.SetIsActive(ctx, tt.userID, tt.isActive)

			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedActive, result.IsActive)
			}

			repo.AssertExpectations(t)
		})
	}
}

