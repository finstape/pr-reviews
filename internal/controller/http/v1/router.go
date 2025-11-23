package v1

import (
	"github.com/finstape/pr-reviews/internal/usecase"
	"github.com/finstape/pr-reviews/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

// NewRouter -.
func NewRouter(apiGroup fiber.Router, teamUseCase usecase.Team, userUseCase usecase.User, pullRequestUseCase usecase.PullRequest, l logger.Interface) {
	v1 := New(teamUseCase, userUseCase, pullRequestUseCase, l)

	// Teams
	apiGroup.Post("/team/add", v1.createTeam)
	apiGroup.Get("/team/get", v1.getTeam)

	// Users
	apiGroup.Post("/users/setIsActive", v1.setIsActive)
	apiGroup.Get("/users/getReview", v1.getUserReviews)

	// Pull Requests
	apiGroup.Post("/pullRequest/create", v1.createPR)
	apiGroup.Post("/pullRequest/merge", v1.mergePR)
	apiGroup.Post("/pullRequest/reassign", v1.reassignReviewer)
}

