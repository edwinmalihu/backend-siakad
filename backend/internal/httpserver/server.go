package httpserver

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	apidocs "siakad/backend/docs"
	"siakad/backend/internal/database"
	"siakad/backend/internal/response"
)

type Module interface {
	Name() string
	BasePath() string
	RegisterRoutes(mux *http.ServeMux)
}

type ServerOptions struct {
	AppName      string
	Environment  string
	Address      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	DB           *sql.DB
	Logger       *slog.Logger
	Modules      []Module
}

func New(opts ServerOptions) *http.Server {
	mux := http.NewServeMux()
	registerBaseRoutes(mux, opts)

	for _, module := range opts.Modules {
		module.RegisterRoutes(mux)
	}

	return &http.Server{
		Addr:         opts.Address,
		Handler:      loggingMiddleware(opts.Logger, mux),
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		IdleTimeout:  opts.IdleTimeout,
	}
}

func registerBaseRoutes(mux *http.ServeMux, opts ServerOptions) {
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(apidocs.OpenAPIYAML)
	})

	mux.HandleFunc("GET /docs/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(apidocs.OpenAPIYAML)
	})

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "disabled"
		statusCode := http.StatusOK

		if opts.DB != nil {
			if err := database.HealthCheck(r.Context(), opts.DB); err != nil {
				dbStatus = "down"
				statusCode = http.StatusServiceUnavailable
			} else {
				dbStatus = "up"
			}
		}

		response.JSON(w, statusCode, map[string]any{
			"success": true,
			"service": opts.AppName,
			"env":     opts.Environment,
			"db":      dbStatus,
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	mux.HandleFunc("GET /api/v1", func(w http.ResponseWriter, r *http.Request) {
		modules := make([]map[string]string, 0, len(opts.Modules))
		for _, module := range opts.Modules {
			modules = append(modules, map[string]string{
				"name": module.Name(),
				"path": module.BasePath(),
			})
		}

		response.JSON(w, http.StatusOK, map[string]any{
			"success": true,
			"message": "SIAKAD backend API is ready",
			"modules": modules,
		})
	})
}

func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	if logger == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(w, r)
		logger.InfoContext(context.Background(), "http request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"duration", time.Since(startedAt).String(),
		)
	})
}
