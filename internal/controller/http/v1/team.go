package v1

import (
	"github.com/finstape/pr-reviews/internal/controller/http/v1/request"
	"github.com/finstape/pr-reviews/internal/entity"
	"github.com/gofiber/fiber/v2"
)

// createTeam - POST /team/add
func (v *V1) createTeam(c *fiber.Ctx) error {
	var req request.CreateTeamRequest
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

	// Convert request to entity
	members := make([]entity.TeamMember, 0, len(req.Members))
	for _, m := range req.Members {
		members = append(members, entity.TeamMember{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	team := entity.Team{
		TeamName: req.TeamName,
		Members:  members,
	}

	err := v.teamUseCase.CreateTeam(c.Context(), team)
	if err != nil {
		return v.handleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"team": team,
	})
}

// getTeam - GET /team/get
func (v *V1) getTeam(c *fiber.Ctx) error {
	teamName := c.Query("team_name")
	if teamName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "BAD_REQUEST",
				"message": "team_name is required",
			},
		})
	}

	team, err := v.teamUseCase.GetTeam(c.Context(), teamName)
	if err != nil {
		return v.handleError(c, err)
	}

	return c.JSON(team)
}

