package main

import (
	"log"

	"github.com/finstape/pr-reviews/config"
	"github.com/finstape/pr-reviews/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}

