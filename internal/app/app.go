// Package app configures and runs application.
package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/finstape/pr-reviews/config"
	"github.com/finstape/pr-reviews/internal/controller/http"
	"github.com/finstape/pr-reviews/internal/repo/persistent"
	"github.com/finstape/pr-reviews/internal/usecase/pullrequest"
	"github.com/finstape/pr-reviews/internal/usecase/team"
	"github.com/finstape/pr-reviews/internal/usecase/user"
	"github.com/finstape/pr-reviews/pkg/httpserver"
	"github.com/finstape/pr-reviews/pkg/logger"
	"github.com/finstape/pr-reviews/pkg/postgres"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	l := logger.New(cfg.Log.Level)

	// Repository
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	// Repositories
	teamRepo := persistent.NewTeamRepo(pg)
	userRepo := persistent.NewUserRepo(pg)
	prRepo := persistent.NewPullRequestRepo(pg)

	// Use cases
	teamUseCase := team.New(teamRepo)
	userUseCase := user.New(userRepo)
	pullRequestUseCase := pullrequest.New(prRepo, userRepo, teamRepo)

	// HTTP Server
	httpServer := httpserver.New(l, httpserver.Port(cfg.HTTP.Port), httpserver.Prefork(cfg.HTTP.UsePreforkMode))
	http.NewRouter(httpServer.App, cfg, teamUseCase, userUseCase, pullRequestUseCase, l)

	// Start servers
	httpServer.Start()

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("app - Run - signal: %s", s.String())
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}

