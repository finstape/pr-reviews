package pullrequest

import (
	"context"
	"testing"
	"time"

	"github.com/finstape/pr-reviews/internal/entity"
	"github.com/finstape/pr-reviews/internal/repo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type mockPRRepo struct {
	mock.Mock
}

func (m *mockPRRepo) CreatePR(ctx context.Context, pr entity.PullRequest, reviewerIDs []string) error {
	args := m.Called(ctx, pr, reviewerIDs)
	return args.Error(0)
}

func (m *mockPRRepo) GetPR(ctx context.Context, prID string) (entity.PullRequest, error) {
	args := m.Called(ctx, prID)
	if args.Get(0) == nil {
		return entity.PullRequest{}, args.Error(1)
	}
	return args.Get(0).(entity.PullRequest), args.Error(1)
}

func (m *mockPRRepo) PRExists(ctx context.Context, prID string) (bool, error) {
	args := m.Called(ctx, prID)
	return args.Bool(0), args.Error(1)
}

func (m *mockPRRepo) UpdatePRStatus(ctx context.Context, prID string, status entity.PullRequestStatus, mergedAt *entity.Time) error {
	args := m.Called(ctx, prID, status, mergedAt)
	return args.Error(0)
}

func (m *mockPRRepo) GetPRReviewers(ctx context.Context, prID string) ([]string, error) {
	args := m.Called(ctx, prID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockPRRepo) ReassignReviewer(ctx context.Context, prID string, oldReviewerID string, newReviewerID string) error {
	args := m.Called(ctx, prID, oldReviewerID, newReviewerID)
	return args.Error(0)
}

func (m *mockPRRepo) GetPRsByReviewer(ctx context.Context, reviewerID string) ([]entity.PullRequestShort, error) {
	args := m.Called(ctx, reviewerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.PullRequestShort), args.Error(1)
}

var _ repo.PullRequestRepo = (*mockPRRepo)(nil)

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

var _ repo.UserRepo = (*mockUserRepo)(nil)

type mockTeamRepo struct {
	mock.Mock
}

func (m *mockTeamRepo) CreateTeam(ctx context.Context, team entity.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *mockTeamRepo) GetTeam(ctx context.Context, teamName string) (entity.Team, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return entity.Team{}, args.Error(1)
	}
	return args.Get(0).(entity.Team), args.Error(1)
}

func (m *mockTeamRepo) TeamExists(ctx context.Context, teamName string) (bool, error) {
	args := m.Called(ctx, teamName)
	return args.Bool(0), args.Error(1)
}

var _ repo.TeamRepo = (*mockTeamRepo)(nil)

func TestCreatePR_Success(t *testing.T) {
	prRepo := new(mockPRRepo)
	userRepo := new(mockUserRepo)
	teamRepo := new(mockTeamRepo)

	uc := New(prRepo, userRepo, teamRepo)

	ctx := context.Background()
	prID := "pr-1"
	prName := "Test PR"
	authorID := "u1"

	author := entity.User{
		UserID:   authorID,
		Username: "Author",
		TeamName: "team1",
		IsActive: true,
	}

	candidates := []entity.User{
		{UserID: "u2", Username: "Reviewer1", TeamName: "team1", IsActive: true},
		{UserID: "u3", Username: "Reviewer2", TeamName: "team1", IsActive: true},
	}

	prRepo.On("PRExists", ctx, prID).Return(false, nil)
	userRepo.On("GetUser", ctx, authorID).Return(author, nil)
	userRepo.On("GetActiveTeamMembers", ctx, "team1", authorID).Return(candidates, nil)

	prRepo.On("CreatePR", ctx, mock.Anything, []string{"u2", "u3"}).Return(nil)

	pr, err := uc.CreatePR(ctx, prID, prName, authorID)

	assert.NoError(t, err)
	assert.Equal(t, prID, pr.PullRequestID)
	assert.Equal(t, prName, pr.PullRequestName)
	assert.Equal(t, authorID, pr.AuthorID)
	assert.Equal(t, entity.PullRequestStatusOpen, pr.Status)
	assert.Len(t, pr.AssignedReviewers, 2)

	prRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestCreatePR_AlreadyExists(t *testing.T) {
	prRepo := new(mockPRRepo)
	userRepo := new(mockUserRepo)
	teamRepo := new(mockTeamRepo)

	uc := New(prRepo, userRepo, teamRepo)

	ctx := context.Background()
	prID := "pr-1"

	prRepo.On("PRExists", ctx, prID).Return(true, nil)

	_, err := uc.CreatePR(ctx, prID, "Test", "u1")

	assert.Error(t, err)
	assert.Equal(t, entity.ErrPRExists, err)
}

func TestCreatePR_AuthorNotFound(t *testing.T) {
	prRepo := new(mockPRRepo)
	userRepo := new(mockUserRepo)
	teamRepo := new(mockTeamRepo)

	uc := New(prRepo, userRepo, teamRepo)

	ctx := context.Background()
	prID := "pr-1"

	prRepo.On("PRExists", ctx, prID).Return(false, nil)
	userRepo.On("GetUser", ctx, "u99").Return(entity.User{}, entity.ErrNotFound)

	_, err := uc.CreatePR(ctx, prID, "Test", "u99")

	assert.Error(t, err)
	prRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestMergePR_NotFound(t *testing.T) {
	prRepo := new(mockPRRepo)
	userRepo := new(mockUserRepo)
	teamRepo := new(mockTeamRepo)

	uc := New(prRepo, userRepo, teamRepo)

	ctx := context.Background()
	prID := "pr-99"

	prRepo.On("GetPR", ctx, prID).Return(entity.PullRequest{}, entity.ErrNotFound)

	_, err := uc.MergePR(ctx, prID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrNotFound)
}

func TestMergePR_Success(t *testing.T) {
	prRepo := new(mockPRRepo)
	userRepo := new(mockUserRepo)
	teamRepo := new(mockTeamRepo)

	uc := New(prRepo, userRepo, teamRepo)

	ctx := context.Background()
	prID := "pr-1"

	now := time.Now()
	pr := entity.PullRequest{
		PullRequestID:    prID,
		PullRequestName:  "Test PR",
		AuthorID:         "u1",
		Status:           entity.PullRequestStatusOpen,
		AssignedReviewers: []string{"u2"},
		CreatedAt:        &now,
	}

	mergedAt := entity.Time(time.Now())
	mergedPR := pr
	mergedPR.Status = entity.PullRequestStatusMerged
	mergedPR.MergedAt = &time.Time{}
	*mergedPR.MergedAt = time.Time(mergedAt)

	prRepo.On("GetPR", ctx, prID).Return(pr, nil).Once()
	prRepo.On("UpdatePRStatus", ctx, prID, entity.PullRequestStatusMerged, mock.Anything).Return(nil)
	prRepo.On("GetPR", ctx, prID).Return(mergedPR, nil).Once()

	result, err := uc.MergePR(ctx, prID)

	assert.NoError(t, err)
	assert.Equal(t, entity.PullRequestStatusMerged, result.Status)
	assert.NotNil(t, result.MergedAt)

	prRepo.AssertExpectations(t)
}

func TestMergePR_Idempotent(t *testing.T) {
	prRepo := new(mockPRRepo)
	userRepo := new(mockUserRepo)
	teamRepo := new(mockTeamRepo)

	uc := New(prRepo, userRepo, teamRepo)

	ctx := context.Background()
	prID := "pr-1"

	now := time.Now()
	mergedAt := now.Add(time.Hour)
	pr := entity.PullRequest{
		PullRequestID:    prID,
		PullRequestName:  "Test PR",
		AuthorID:         "u1",
		Status:           entity.PullRequestStatusMerged,
		AssignedReviewers: []string{"u2"},
		CreatedAt:        &now,
		MergedAt:         &mergedAt,
	}

	prRepo.On("GetPR", ctx, prID).Return(pr, nil)

	result, err := uc.MergePR(ctx, prID)

	assert.NoError(t, err)
	assert.Equal(t, entity.PullRequestStatusMerged, result.Status)
	// Should not call UpdatePRStatus
	prRepo.AssertNotCalled(t, "UpdatePRStatus")
}

func TestReassignReviewer_Success(t *testing.T) {
	prRepo := new(mockPRRepo)
	userRepo := new(mockUserRepo)
	teamRepo := new(mockTeamRepo)

	uc := New(prRepo, userRepo, teamRepo)

	ctx := context.Background()
	prID := "pr-1"
	oldReviewerID := "u2"
	newReviewerID := "u4"

	now := time.Now()
	pr := entity.PullRequest{
		PullRequestID:    prID,
		PullRequestName:  "Test PR",
		AuthorID:         "u1",
		Status:           entity.PullRequestStatusOpen,
		AssignedReviewers: []string{oldReviewerID, "u3"},
		CreatedAt:        &now,
	}

	oldReviewer := entity.User{
		UserID:   oldReviewerID,
		Username: "Old Reviewer",
		TeamName: "team1",
		IsActive: true,
	}

	candidates := []entity.User{
		{UserID: newReviewerID, Username: "New Reviewer", TeamName: "team1", IsActive: true},
	}

	updatedPR := pr
	updatedPR.AssignedReviewers = []string{newReviewerID, "u3"}

	prRepo.On("GetPR", ctx, prID).Return(pr, nil).Once()
	userRepo.On("GetUser", ctx, oldReviewerID).Return(oldReviewer, nil)
	userRepo.On("GetActiveTeamMembers", ctx, "team1", oldReviewerID).Return(candidates, nil)
	prRepo.On("ReassignReviewer", ctx, prID, oldReviewerID, newReviewerID).Return(nil)
	prRepo.On("GetPR", ctx, prID).Return(updatedPR, nil).Once()

	result, newID, err := uc.ReassignReviewer(ctx, prID, oldReviewerID)

	assert.NoError(t, err)
	assert.Equal(t, newReviewerID, newID)
	assert.Contains(t, result.AssignedReviewers, newReviewerID)
	assert.NotContains(t, result.AssignedReviewers, oldReviewerID)

	prRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestReassignReviewer_MergedPR(t *testing.T) {
	prRepo := new(mockPRRepo)
	userRepo := new(mockUserRepo)
	teamRepo := new(mockTeamRepo)

	uc := New(prRepo, userRepo, teamRepo)

	ctx := context.Background()
	prID := "pr-1"

	now := time.Now()
	mergedAt := now.Add(time.Hour)
	pr := entity.PullRequest{
		PullRequestID:    prID,
		PullRequestName:  "Test PR",
		AuthorID:         "u1",
		Status:           entity.PullRequestStatusMerged,
		AssignedReviewers: []string{"u2"},
		CreatedAt:        &now,
		MergedAt:         &mergedAt,
	}

	prRepo.On("GetPR", ctx, prID).Return(pr, nil)

	_, _, err := uc.ReassignReviewer(ctx, prID, "u2")

	assert.Error(t, err)
	assert.Equal(t, entity.ErrPRMerged, err)
}

func TestReassignReviewer_NotAssigned(t *testing.T) {
	prRepo := new(mockPRRepo)
	userRepo := new(mockUserRepo)
	teamRepo := new(mockTeamRepo)

	uc := New(prRepo, userRepo, teamRepo)

	ctx := context.Background()
	prID := "pr-1"

	now := time.Now()
	pr := entity.PullRequest{
		PullRequestID:    prID,
		PullRequestName:  "Test PR",
		AuthorID:         "u1",
		Status:           entity.PullRequestStatusOpen,
		AssignedReviewers: []string{"u2"},
		CreatedAt:        &now,
	}

	prRepo.On("GetPR", ctx, prID).Return(pr, nil)

	_, _, err := uc.ReassignReviewer(ctx, prID, "u99")

	assert.Error(t, err)
	assert.Equal(t, entity.ErrNotAssigned, err)
}

func TestReassignReviewer_NoCandidate(t *testing.T) {
	prRepo := new(mockPRRepo)
	userRepo := new(mockUserRepo)
	teamRepo := new(mockTeamRepo)

	uc := New(prRepo, userRepo, teamRepo)

	ctx := context.Background()
	prID := "pr-1"
	oldReviewerID := "u2"

	now := time.Now()
	pr := entity.PullRequest{
		PullRequestID:    prID,
		PullRequestName:  "Test PR",
		AuthorID:         "u1",
		Status:           entity.PullRequestStatusOpen,
		AssignedReviewers: []string{oldReviewerID},
		CreatedAt:        &now,
	}

	oldReviewer := entity.User{
		UserID:   oldReviewerID,
		Username: "Old Reviewer",
		TeamName: "team1",
		IsActive: true,
	}

	prRepo.On("GetPR", ctx, prID).Return(pr, nil)
	userRepo.On("GetUser", ctx, oldReviewerID).Return(oldReviewer, nil)
	userRepo.On("GetActiveTeamMembers", ctx, "team1", oldReviewerID).Return([]entity.User{}, nil)

	_, _, err := uc.ReassignReviewer(ctx, prID, oldReviewerID)

	assert.Error(t, err)
	assert.Equal(t, entity.ErrNoCandidate, err)
}

