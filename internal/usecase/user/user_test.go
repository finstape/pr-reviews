package user

import (
	"context"
	"testing"

	"github.com/finstape/pr-reviews/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) CreateOrUpdateUser(ctx context.Context, user entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) GetUser(ctx context.Context, userID string) (entity.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return entity.User{}, args.Error(1)
	}
	return args.Get(0).(entity.User), args.Error(1)
}

func (m *mockUserRepo) SetIsActive(ctx context.Context, userID string, isActive bool) error {
	args := m.Called(ctx, userID, isActive)
	return args.Error(0)
}

func (m *mockUserRepo) GetActiveTeamMembers(ctx context.Context, teamName string, excludeUserID string) ([]entity.User, error) {
	args := m.Called(ctx, teamName, excludeUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.User), args.Error(1)
}

func (m *mockUserRepo) GetUserReviews(ctx context.Context, userID string) ([]entity.PullRequestShort, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.PullRequestShort), args.Error(1)
}

func TestSetIsActive_Success(t *testing.T) {
	repo := new(mockUserRepo)
	uc := New(repo)

	ctx := context.Background()
	userID := "u1"
	user := entity.User{
		UserID:   userID,
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	repo.On("GetUser", ctx, userID).Return(user, nil)
	repo.On("SetIsActive", ctx, userID, false).Return(nil)

	result, err := uc.SetIsActive(ctx, userID, false)

	assert.NoError(t, err)
	assert.Equal(t, userID, result.UserID)
	assert.False(t, result.IsActive)
	repo.AssertExpectations(t)
}

func TestGetUserReviews_Success(t *testing.T) {
	repo := new(mockUserRepo)
	uc := New(repo)

	ctx := context.Background()
	userID := "u1"
	user := entity.User{
		UserID:   userID,
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	expectedPRs := []entity.PullRequestShort{
		{
			PullRequestID:   "pr-1",
			PullRequestName: "Test PR",
			AuthorID:        "u2",
			Status:          entity.PullRequestStatusOpen,
		},
	}

	repo.On("GetUser", ctx, userID).Return(user, nil)
	repo.On("GetUserReviews", ctx, userID).Return(expectedPRs, nil)

	prs, err := uc.GetUserReviews(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedPRs, prs)
	repo.AssertExpectations(t)
}

func TestGetUserReviews_UserNotFound(t *testing.T) {
	repo := new(mockUserRepo)
	uc := New(repo)

	ctx := context.Background()
	userID := "u99"

	repo.On("GetUser", ctx, userID).Return(entity.User{}, entity.ErrNotFound)

	_, err := uc.GetUserReviews(ctx, userID)

	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestSetIsActive_UserNotFound(t *testing.T) {
	repo := new(mockUserRepo)
	uc := New(repo)

	ctx := context.Background()
	userID := "u99"

	repo.On("GetUser", ctx, userID).Return(entity.User{}, entity.ErrNotFound)

	_, err := uc.SetIsActive(ctx, userID, false)

	assert.Error(t, err)
	repo.AssertExpectations(t)
}

