package main

import (
	"os"

	"github.com/be-users-metadata-service/internal/bootstrap"
	"github.com/be-users-metadata-service/internal/config"
)

func main() {
	cfg := config.Load()
	app, err := bootstrap.New(cfg)
	if err != nil {
		os.Exit(1)
	}
	if err := app.Run(); err != nil {
		os.Exit(1)
	}
}
