package v1

import (
	"github.com/finstape/pr-reviews/internal/controller/http/v1/response"
	"github.com/finstape/pr-reviews/internal/entity"
	"github.com/finstape/pr-reviews/internal/usecase"
	"github.com/finstape/pr-reviews/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// V1 represents API v1 controller.
type V1 struct {
	teamUseCase        usecase.Team
	userUseCase        usecase.User
	pullRequestUseCase usecase.PullRequest
	l                  logger.Interface
	v                  *validator.Validate
}

// New creates a new V1 controller instance.
func New(teamUseCase usecase.Team, userUseCase usecase.User, pullRequestUseCase usecase.PullRequest, l logger.Interface) *V1 {
	return &V1{
		teamUseCase:        teamUseCase,
		userUseCase:        userUseCase,
		pullRequestUseCase: pullRequestUseCase,
		l:                  l,
		v:                  validator.New(validator.WithRequiredStructEnabled()),
	}
}

// handleError handles domain errors and returns appropriate HTTP response
func (v *V1) handleError(c *fiber.Ctx, err error) error {
	code := entity.GetErrorCode(err)
	message := err.Error()

	var statusCode int
	switch code {
	case entity.ErrorCodeTeamExists, entity.ErrorCodePRExists:
		statusCode = fiber.StatusConflict
	case entity.ErrorCodePRMerged, entity.ErrorCodeNotAssigned, entity.ErrorCodeNoCandidate:
		statusCode = fiber.StatusConflict
	case entity.ErrorCodeNotFound:
		statusCode = fiber.StatusNotFound
	default:
		statusCode = fiber.StatusInternalServerError
		// Don't expose internal error details
		message = "internal server error"
		v.l.Error(err, "internal error")
	}

	return c.Status(statusCode).JSON(response.NewErrorResponse(code, message))
}

