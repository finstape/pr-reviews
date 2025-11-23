package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/finstape/pr-reviews/internal/controller/http/v1/request"
	"github.com/finstape/pr-reviews/internal/entity"
	"github.com/finstape/pr-reviews/internal/usecase"
	"github.com/finstape/pr-reviews/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/gofiber/fiber/v2"
)

type mockTeamUseCase struct {
	mock.Mock
}

func (m *mockTeamUseCase) CreateTeam(ctx context.Context, team entity.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *mockTeamUseCase) GetTeam(ctx context.Context, teamName string) (entity.Team, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return entity.Team{}, args.Error(1)
	}
	return args.Get(0).(entity.Team), args.Error(1)
}

var _ usecase.Team = (*mockTeamUseCase)(nil)

type mockUserUseCase struct {
	mock.Mock
}

func (m *mockUserUseCase) SetIsActive(ctx context.Context, userID string, isActive bool) (entity.User, error) {
	args := m.Called(ctx, userID, isActive)
	if args.Get(0) == nil {
		return entity.User{}, args.Error(1)
	}
	return args.Get(0).(entity.User), args.Error(1)
}

func (m *mockUserUseCase) GetUserReviews(ctx context.Context, userID string) ([]entity.PullRequestShort, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.PullRequestShort), args.Error(1)
}

var _ usecase.User = (*mockUserUseCase)(nil)

type mockPullRequestUseCase struct {
	mock.Mock
}

func (m *mockPullRequestUseCase) CreatePR(ctx context.Context, prID string, prName string, authorID string) (entity.PullRequest, error) {
	args := m.Called(ctx, prID, prName, authorID)
	if args.Get(0) == nil {
		return entity.PullRequest{}, args.Error(1)
	}
	return args.Get(0).(entity.PullRequest), args.Error(1)
}

func (m *mockPullRequestUseCase) MergePR(ctx context.Context, prID string) (entity.PullRequest, error) {
	args := m.Called(ctx, prID)
	if args.Get(0) == nil {
		return entity.PullRequest{}, args.Error(1)
	}
	return args.Get(0).(entity.PullRequest), args.Error(1)
}

func (m *mockPullRequestUseCase) ReassignReviewer(ctx context.Context, prID string, oldReviewerID string) (entity.PullRequest, string, error) {
	args := m.Called(ctx, prID, oldReviewerID)
	if args.Get(0) == nil {
		return entity.PullRequest{}, "", args.Error(2)
	}
	return args.Get(0).(entity.PullRequest), args.String(1), args.Error(2)
}

var _ usecase.PullRequest = (*mockPullRequestUseCase)(nil)

func TestCreateTeamHandler_Success(t *testing.T) {
	app := fiber.New()
	teamUC := new(mockTeamUseCase)
	userUC := new(mockUserUseCase)
	prUC := new(mockPullRequestUseCase)
	
	v1 := New(teamUC, userUC, prUC, logger.New("error"))

	reqBody := request.CreateTeamRequest{
		TeamName: "test-team",
		Members: []request.CreateTeamMemberRequest{
			{UserID: "u1", Username: "User1", IsActive: true},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	teamUC.On("CreateTeam", mock.Anything, mock.AnythingOfType("entity.Team")).Return(nil)

	app.Post("/team/add", v1.createTeam)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	teamUC.AssertExpectations(t)
}

func TestCreateTeamHandler_InvalidBody(t *testing.T) {
	app := fiber.New()
	teamUC := new(mockTeamUseCase)
	userUC := new(mockUserUseCase)
	prUC := new(mockPullRequestUseCase)
	
	v1 := New(teamUC, userUC, prUC, logger.New("error"))

	req := httptest.NewRequest("POST", "/team/add", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	app.Post("/team/add", v1.createTeam)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetTeamHandler_Success(t *testing.T) {
	app := fiber.New()
	teamUC := new(mockTeamUseCase)
	userUC := new(mockUserUseCase)
	prUC := new(mockPullRequestUseCase)
	
	v1 := New(teamUC, userUC, prUC, logger.New("error"))

	expectedTeam := entity.Team{
		TeamName: "test-team",
		Members: []entity.TeamMember{
			{UserID: "u1", Username: "User1", IsActive: true},
		},
	}

	req := httptest.NewRequest("GET", "/team/get?team_name=test-team", nil)
	teamUC.On("GetTeam", mock.Anything, "test-team").Return(expectedTeam, nil)

	app.Get("/team/get", v1.getTeam)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	teamUC.AssertExpectations(t)
}

func TestGetTeamHandler_MissingParam(t *testing.T) {
	app := fiber.New()
	teamUC := new(mockTeamUseCase)
	userUC := new(mockUserUseCase)
	prUC := new(mockPullRequestUseCase)
	
	v1 := New(teamUC, userUC, prUC, logger.New("error"))

	req := httptest.NewRequest("GET", "/team/get", nil)

	app.Get("/team/get", v1.getTeam)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

