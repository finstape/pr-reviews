//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/finstape/pr-reviews/internal/entity"
	"github.com/finstape/pr-reviews/internal/repo/persistent"
	"github.com/finstape/pr-reviews/pkg/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDB *postgres.Postgres

func TestMain(m *testing.M) {
	// Setup test database
	dbURL := os.Getenv("PG_URL")
	if dbURL == "" {
		dbURL = "postgres://test_user:test_password@db-test:5432/db_test?sslmode=disable"
	}

	var err error
	testDB, err = postgres.New(dbURL, postgres.MaxPoolSize(2))
	if err != nil {
		fmt.Printf("Failed to connect to test database: %v\n", err)
		os.Exit(1)
	}
	defer testDB.Close()

	code := m.Run()
	os.Exit(code)
}

func TestIntegration_Repository_CreateTeamAndGetTeam(t *testing.T) {
	ctx := context.Background()
	teamRepo := persistent.NewTeamRepo(testDB)

	// Clean up before test
	_, _ = testDB.Pool.Exec(ctx, "DELETE FROM teams WHERE team_name = 'integration-test-team'")

	team := entity.Team{
		TeamName: "integration-test-team",
		Members: []entity.TeamMember{
			{UserID: "int-u1", Username: "Integration User 1", IsActive: true},
			{UserID: "int-u2", Username: "Integration User 2", IsActive: true},
		},
	}

	// Test CreateTeam
	err := teamRepo.CreateTeam(ctx, team)
	require.NoError(t, err)

	// Test GetTeam
	retrievedTeam, err := teamRepo.GetTeam(ctx, "integration-test-team")
	require.NoError(t, err)
	assert.Equal(t, team.TeamName, retrievedTeam.TeamName)
	assert.Len(t, retrievedTeam.Members, 2)

	// Clean up
	_, _ = testDB.Pool.Exec(ctx, "DELETE FROM teams WHERE team_name = 'integration-test-team'")
}

func TestIntegration_Repository_CreateAndMergePR(t *testing.T) {
	ctx := context.Background()
	teamRepo := persistent.NewTeamRepo(testDB)
	userRepo := persistent.NewUserRepo(testDB)
	prRepo := persistent.NewPullRequestRepo(testDB)

	// Setup: create team
	team := entity.Team{
		TeamName: "merge-repo-test-team",
		Members: []entity.TeamMember{
			{UserID: "merge-u1", Username: "Merge User 1", IsActive: true},
			{UserID: "merge-u2", Username: "Merge User 2", IsActive: true},
		},
	}

	// Clean up
	_, _ = testDB.Pool.Exec(ctx, "DELETE FROM teams WHERE team_name = 'merge-repo-test-team'")

	err := teamRepo.CreateTeam(ctx, team)
	require.NoError(t, err)

	prID := "pr-merge-repo-test"
	now := time.Now()
	pr := entity.PullRequest{
		PullRequestID:    prID,
		PullRequestName:  "Merge Repository Test PR",
		AuthorID:         "merge-u1",
		Status:           entity.PullRequestStatusOpen,
		AssignedReviewers: []string{"merge-u2"},
		CreatedAt:        &now,
	}

	// Clean up PR
	_, _ = testDB.Pool.Exec(ctx, "DELETE FROM pull_requests WHERE pull_request_id = $1", prID)

	// Test CreatePR
	err = prRepo.CreatePR(ctx, pr, []string{"merge-u2"})
	require.NoError(t, err)

	// Test GetPR
	retrievedPR, err := prRepo.GetPR(ctx, prID)
	require.NoError(t, err)
	assert.Equal(t, prID, retrievedPR.PullRequestID)
	assert.Equal(t, entity.PullRequestStatusOpen, retrievedPR.Status)

	// Test UpdatePRStatus (merge)
	mergedAt := entity.Time(time.Now())
	err = prRepo.UpdatePRStatus(ctx, prID, entity.PullRequestStatusMerged, &mergedAt)
	require.NoError(t, err)

	// Verify merge
	mergedPR, err := prRepo.GetPR(ctx, prID)
	require.NoError(t, err)
	assert.Equal(t, entity.PullRequestStatusMerged, mergedPR.Status)
	assert.NotNil(t, mergedPR.MergedAt)

	// Clean up
	_, _ = testDB.Pool.Exec(ctx, "DELETE FROM pull_requests WHERE pull_request_id = $1", prID)
	_, _ = testDB.Pool.Exec(ctx, "DELETE FROM teams WHERE team_name = 'merge-repo-test-team'")
}

func TestIntegration_Repository_ReassignReviewer(t *testing.T) {
	ctx := context.Background()
	teamRepo := persistent.NewTeamRepo(testDB)
	prRepo := persistent.NewPullRequestRepo(testDB)

	// Setup: create team
	team := entity.Team{
		TeamName: "reassign-repo-test-team",
		Members: []entity.TeamMember{
			{UserID: "reassign-u1", Username: "Reassign User 1", IsActive: true},
			{UserID: "reassign-u2", Username: "Reassign User 2", IsActive: true},
			{UserID: "reassign-u3", Username: "Reassign User 3", IsActive: true},
		},
	}

	// Clean up
	_, _ = testDB.Pool.Exec(ctx, "DELETE FROM teams WHERE team_name = 'reassign-repo-test-team'")

	err := teamRepo.CreateTeam(ctx, team)
	require.NoError(t, err)

	prID := "pr-reassign-repo-test"
	now := time.Now()
	pr := entity.PullRequest{
		PullRequestID:    prID,
		PullRequestName:  "Reassign Repository Test PR",
		AuthorID:         "reassign-u1",
		Status:           entity.PullRequestStatusOpen,
		AssignedReviewers: []string{"reassign-u2"},
		CreatedAt:        &now,
	}

	// Clean up PR
	_, _ = testDB.Pool.Exec(ctx, "DELETE FROM pull_requests WHERE pull_request_id = $1", prID)

	err = prRepo.CreatePR(ctx, pr, []string{"reassign-u2"})
	require.NoError(t, err)

	// Test ReassignReviewer
	err = prRepo.ReassignReviewer(ctx, prID, "reassign-u2", "reassign-u3")
	require.NoError(t, err)

	// Verify reassignment
	updatedPR, err := prRepo.GetPR(ctx, prID)
	require.NoError(t, err)
	assert.Contains(t, updatedPR.AssignedReviewers, "reassign-u3")
	assert.NotContains(t, updatedPR.AssignedReviewers, "reassign-u2")

	// Clean up
	_, _ = testDB.Pool.Exec(ctx, "DELETE FROM pull_requests WHERE pull_request_id = $1", prID)
	_, _ = testDB.Pool.Exec(ctx, "DELETE FROM teams WHERE team_name = 'reassign-repo-test-team'")
}

func TestIntegration_Repository_GetUserReviews(t *testing.T) {
	ctx := context.Background()
	teamRepo := persistent.NewTeamRepo(testDB)
	userRepo := persistent.NewUserRepo(testDB)
	prRepo := persistent.NewPullRequestRepo(testDB)

	// Setup: create team
	team := entity.Team{
		TeamName: "reviews-repo-test-team",
		Members: []entity.TeamMember{
			{UserID: "reviews-u1", Username: "Reviews User 1", IsActive: true},
			{UserID: "reviews-u2", Username: "Reviews User 2", IsActive: true},
		},
	}

	// Clean up
	_, _ = testDB.Pool.Exec(ctx, "DELETE FROM teams WHERE team_name = 'reviews-repo-test-team'")

	err := teamRepo.CreateTeam(ctx, team)
	require.NoError(t, err)

	prID := "pr-reviews-repo-test"
	now := time.Now()
	pr := entity.PullRequest{
		PullRequestID:    prID,
		PullRequestName:  "Reviews Repository Test PR",
		AuthorID:         "reviews-u1",
		Status:           entity.PullRequestStatusOpen,
		AssignedReviewers: []string{"reviews-u2"},
		CreatedAt:        &now,
	}

	// Clean up PR
	_, _ = testDB.Pool.Exec(ctx, "DELETE FROM pull_requests WHERE pull_request_id = $1", prID)

	err = prRepo.CreatePR(ctx, pr, []string{"reviews-u2"})
	require.NoError(t, err)

	// Test GetUserReviews
	reviews, err := userRepo.GetUserReviews(ctx, "reviews-u2")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(reviews), 1)

	found := false
	for _, review := range reviews {
		if review.PullRequestID == prID {
			found = true
			break
		}
	}
	assert.True(t, found, "PR should be in user reviews")

	// Clean up
	_, _ = testDB.Pool.Exec(ctx, "DELETE FROM pull_requests WHERE pull_request_id = $1", prID)
	_, _ = testDB.Pool.Exec(ctx, "DELETE FROM teams WHERE team_name = 'reviews-repo-test-team'")
}

func TestIntegration_Repository_SetIsActive(t *testing.T) {
	ctx := context.Background()
	teamRepo := persistent.NewTeamRepo(testDB)
	userRepo := persistent.NewUserRepo(testDB)

	// Setup: create team
	team := entity.Team{
		TeamName: "active-test-team",
		Members: []entity.TeamMember{
			{UserID: "active-u1", Username: "Active User 1", IsActive: true},
		},
	}

	// Clean up
	_, _ = testDB.Pool.Exec(ctx, "DELETE FROM teams WHERE team_name = 'active-test-team'")

	err := teamRepo.CreateTeam(ctx, team)
	require.NoError(t, err)

	// Test SetIsActive
	err = userRepo.SetIsActive(ctx, "active-u1", false)
	require.NoError(t, err)

	// Verify
	user, err := userRepo.GetUser(ctx, "active-u1")
	require.NoError(t, err)
	assert.False(t, user.IsActive)

	// Set back to active
	err = userRepo.SetIsActive(ctx, "active-u1", true)
	require.NoError(t, err)

	user, err = userRepo.GetUser(ctx, "active-u1")
	require.NoError(t, err)
	assert.True(t, user.IsActive)

	// Clean up
	_, _ = testDB.Pool.Exec(ctx, "DELETE FROM teams WHERE team_name = 'active-test-team'")
}

