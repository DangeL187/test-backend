package main

import (
	"go.uber.org/zap"

	"back/internal/app"
	"back/internal/infra/http/server"
)

func main() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	defer func() {
		_ = logger.Sync()
	}()

	application, err := app.NewApp()
	if err != nil {
		zap.S().Fatalf("Failed to create App:\n%f", err)
	}

	httpServer := server.NewServer(application)
	err = httpServer.Run("0.0.0.0:8000")
	if err != nil {
		zap.S().Fatalf("Failed to run HTTP server:\n%f", err)
	}
}
