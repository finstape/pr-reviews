package v1

import (
	"github.com/finstape/pr-reviews/internal/controller/http/v1/request"
	"github.com/gofiber/fiber/v2"
)

// createPR - POST /pullRequest/create
func (v *V1) createPR(c *fiber.Ctx) error {
	var req request.CreatePRRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "BAD_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	if err := v.v.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "BAD_REQUEST",
				"message": err.Error(),
			},
		})
	}

	pr, err := v.pullRequestUseCase.CreatePR(c.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		return v.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"pr": pr,
	})
}

// mergePR - POST /pullRequest/merge
func (v *V1) mergePR(c *fiber.Ctx) error {
	var req request.MergePRRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "BAD_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	if err := v.v.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "BAD_REQUEST",
				"message": err.Error(),
			},
		})
	}

	pr, err := v.pullRequestUseCase.MergePR(c.Context(), req.PullRequestID)
	if err != nil {
		return v.handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"pr": pr,
	})
}

// reassignReviewer - POST /pullRequest/reassign
func (v *V1) reassignReviewer(c *fiber.Ctx) error {
	var req request.ReassignReviewerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "BAD_REQUEST",
				"message": "invalid request body",
			},
		})
	}

	if err := v.v.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "BAD_REQUEST",
				"message": err.Error(),
			},
		})
	}

	pr, newReviewerID, err := v.pullRequestUseCase.ReassignReviewer(c.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		return v.handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"pr":         pr,
		"replaced_by": newReviewerID,
	})
}

