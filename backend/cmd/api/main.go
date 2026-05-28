package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"siakad/backend/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logFile, err := os.OpenFile("siakad.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		slog.Error("failed to open log file", "error", err)
		os.Exit(1)
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	application, err := app.New()
	if err != nil {
		slog.Error("failed to initialize application", "error", err)
		os.Exit(1)
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), application.Config.App.ShutdownTimeout)
		defer cancel()

		if err := application.Shutdown(shutdownCtx); err != nil {
			application.Logger.Error("failed to shutdown application gracefully", "error", err)
		}
	}()

	if err := application.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		application.Logger.Error("application stopped with error", "error", err)
		os.Exit(1)
	}
}
