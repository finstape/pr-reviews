// Package http implements routing paths.
package http

import (
	"net/http"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/finstape/pr-reviews/config"
	"github.com/finstape/pr-reviews/internal/controller/http/middleware"
	v1 "github.com/finstape/pr-reviews/internal/controller/http/v1"
	"github.com/finstape/pr-reviews/internal/usecase"
	"github.com/finstape/pr-reviews/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

// NewRouter -.
func NewRouter(app *fiber.App, cfg *config.Config, teamUseCase usecase.Team, userUseCase usecase.User, pullRequestUseCase usecase.PullRequest, l logger.Interface) {
	// Options
	app.Use(middleware.LoggerMiddleware(l))
	app.Use(middleware.Recovery(l))

	// Prometheus metrics
	if cfg.Metrics.Enabled {
		prometheus := fiberprometheus.New("pr-review-service")
		prometheus.RegisterAt(app, "/metrics")
		app.Use(prometheus.Middleware)
	}

	// K8s probe
	app.Get("/healthz", func(ctx *fiber.Ctx) error { return ctx.SendStatus(http.StatusOK) })

	// API routes
	v1.NewRouter(app, teamUseCase, userUseCase, pullRequestUseCase, l)
}

