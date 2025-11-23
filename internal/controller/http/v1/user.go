package v1

import (
	"github.com/finstape/pr-reviews/internal/controller/http/v1/request"
	"github.com/gofiber/fiber/v2"
)

// setIsActive - POST /users/setIsActive
func (v *V1) setIsActive(c *fiber.Ctx) error {
	var req request.SetIsActiveRequest
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

	user, err := v.userUseCase.SetIsActive(c.Context(), req.UserID, req.IsActive)
	if err != nil {
		return v.handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"user": user,
	})
}

// getUserReviews - GET /users/getReview
func (v *V1) getUserReviews(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "BAD_REQUEST",
				"message": "user_id is required",
			},
		})
	}

	prs, err := v.userUseCase.GetUserReviews(c.Context(), userID)
	if err != nil {
		return v.handleError(c, err)
	}

	return c.JSON(fiber.Map{
		"user_id":       userID,
		"pull_requests": prs,
	})
}

