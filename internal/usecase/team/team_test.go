package team

import (
	"context"
	"testing"

	"github.com/finstape/pr-reviews/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func TestCreateTeam_Success(t *testing.T) {
	repo := new(mockTeamRepo)
	uc := New(repo)

	ctx := context.Background()
	team := entity.Team{
		TeamName: "backend",
		Members: []entity.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: true},
		},
	}

	repo.On("TeamExists", ctx, "backend").Return(false, nil)
	repo.On("CreateTeam", ctx, team).Return(nil)

	err := uc.CreateTeam(ctx, team)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestCreateTeam_AlreadyExists(t *testing.T) {
	repo := new(mockTeamRepo)
	uc := New(repo)

	ctx := context.Background()
	team := entity.Team{
		TeamName: "backend",
		Members:  []entity.TeamMember{},
	}

	repo.On("TeamExists", ctx, "backend").Return(true, nil)

	err := uc.CreateTeam(ctx, team)

	assert.Error(t, err)
	assert.Equal(t, entity.ErrTeamExists, err)
}

func TestGetTeam_Success(t *testing.T) {
	repo := new(mockTeamRepo)
	uc := New(repo)

	ctx := context.Background()
	expectedTeam := entity.Team{
		TeamName: "backend",
		Members: []entity.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: true},
		},
	}

	repo.On("GetTeam", ctx, "backend").Return(expectedTeam, nil)

	team, err := uc.GetTeam(ctx, "backend")

	assert.NoError(t, err)
	assert.Equal(t, expectedTeam, team)
	repo.AssertExpectations(t)
}

