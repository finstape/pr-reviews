package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/finstape/pr-reviews/internal/controller/http/v1/request"
	"github.com/finstape/pr-reviews/internal/entity"
	"github.com/finstape/pr-reviews/internal/usecase"
	"github.com/finstape/pr-reviews/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/gofiber/fiber/v2"
)

type mockTeamUseCaseForPR struct {
	mock.Mock
}

func (m *mockTeamUseCaseForPR) CreateTeam(ctx context.Context, team entity.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *mockTeamUseCaseForPR) GetTeam(ctx context.Context, teamName string) (entity.Team, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return entity.Team{}, args.Error(1)
	}
	return args.Get(0).(entity.Team), args.Error(1)
}

var _ usecase.Team = (*mockTeamUseCaseForPR)(nil)

type mockUserUseCaseForPR struct {
	mock.Mock
}

func (m *mockUserUseCaseForPR) SetIsActive(ctx context.Context, userID string, isActive bool) (entity.User, error) {
	args := m.Called(ctx, userID, isActive)
	if args.Get(0) == nil {
		return entity.User{}, args.Error(1)
	}
	return args.Get(0).(entity.User), args.Error(1)
}

func (m *mockUserUseCaseForPR) GetUserReviews(ctx context.Context, userID string) ([]entity.PullRequestShort, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.PullRequestShort), args.Error(1)
}

var _ usecase.User = (*mockUserUseCaseForPR)(nil)

type mockPullRequestUseCaseForPR struct {
	mock.Mock
}

func (m *mockPullRequestUseCaseForPR) CreatePR(ctx context.Context, prID string, prName string, authorID string) (entity.PullRequest, error) {
	args := m.Called(ctx, prID, prName, authorID)
	if args.Get(0) == nil {
		return entity.PullRequest{}, args.Error(1)
	}
	return args.Get(0).(entity.PullRequest), args.Error(1)
}

func (m *mockPullRequestUseCaseForPR) MergePR(ctx context.Context, prID string) (entity.PullRequest, error) {
	args := m.Called(ctx, prID)
	if args.Get(0) == nil {
		return entity.PullRequest{}, args.Error(1)
	}
	return args.Get(0).(entity.PullRequest), args.Error(1)
}

func (m *mockPullRequestUseCaseForPR) ReassignReviewer(ctx context.Context, prID string, oldReviewerID string) (entity.PullRequest, string, error) {
	args := m.Called(ctx, prID, oldReviewerID)
	if args.Get(0) == nil {
		return entity.PullRequest{}, "", args.Error(2)
	}
	return args.Get(0).(entity.PullRequest), args.String(1), args.Error(2)
}

var _ usecase.PullRequest = (*mockPullRequestUseCaseForPR)(nil)

func TestCreatePRHandler_Success(t *testing.T) {
	app := fiber.New()
	teamUC := new(mockTeamUseCaseForPR)
	userUC := new(mockUserUseCaseForPR)
	prUC := new(mockPullRequestUseCaseForPR)
	
	v1 := New(teamUC, userUC, prUC, logger.New("error"))

	reqBody := request.CreatePRRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Test PR",
		AuthorID:        "u1",
	}

	now := time.Now()
	expectedPR := entity.PullRequest{
		PullRequestID:    "pr-1",
		PullRequestName:  "Test PR",
		AuthorID:         "u1",
		Status:           entity.PullRequestStatusOpen,
		AssignedReviewers: []string{"u2"},
		CreatedAt:        &now,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	prUC.On("CreatePR", mock.Anything, "pr-1", "Test PR", "u1").Return(expectedPR, nil)

	app.Post("/pullRequest/create", v1.createPR)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	prUC.AssertExpectations(t)
}

func TestMergePRHandler_Success(t *testing.T) {
	app := fiber.New()
	teamUC := new(mockTeamUseCaseForPR)
	userUC := new(mockUserUseCaseForPR)
	prUC := new(mockPullRequestUseCaseForPR)
	
	v1 := New(teamUC, userUC, prUC, logger.New("error"))

	reqBody := request.MergePRRequest{
		PullRequestID: "pr-1",
	}

	now := time.Now()
	mergedAt := now.Add(time.Hour)
	expectedPR := entity.PullRequest{
		PullRequestID:    "pr-1",
		PullRequestName:  "Test PR",
		AuthorID:         "u1",
		Status:           entity.PullRequestStatusMerged,
		AssignedReviewers: []string{"u2"},
		CreatedAt:        &now,
		MergedAt:         &mergedAt,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/pullRequest/merge", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	prUC.On("MergePR", mock.Anything, "pr-1").Return(expectedPR, nil)

	app.Post("/pullRequest/merge", v1.mergePR)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	prUC.AssertExpectations(t)
}

func TestReassignReviewerHandler_Success(t *testing.T) {
	app := fiber.New()
	teamUC := new(mockTeamUseCaseForPR)
	userUC := new(mockUserUseCaseForPR)
	prUC := new(mockPullRequestUseCaseForPR)
	
	v1 := New(teamUC, userUC, prUC, logger.New("error"))

	reqBody := request.ReassignReviewerRequest{
		PullRequestID: "pr-1",
		OldUserID:     "u2",
	}

	now := time.Now()
	expectedPR := entity.PullRequest{
		PullRequestID:    "pr-1",
		PullRequestName:  "Test PR",
		AuthorID:         "u1",
		Status:           entity.PullRequestStatusOpen,
		AssignedReviewers: []string{"u3"},
		CreatedAt:        &now,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/pullRequest/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	prUC.On("ReassignReviewer", mock.Anything, "pr-1", "u2").Return(expectedPR, "u3", nil)

	app.Post("/pullRequest/reassign", v1.reassignReviewer)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	prUC.AssertExpectations(t)
}

