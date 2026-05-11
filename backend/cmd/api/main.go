package main

import (
	"context"
	"errors"
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
