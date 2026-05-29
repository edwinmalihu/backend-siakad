package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"siakad/backend/internal/config"
	"siakad/backend/internal/database"
	"siakad/backend/internal/httpserver"
	"siakad/backend/internal/modules/academic"
	"siakad/backend/internal/modules/auth"
	"siakad/backend/internal/modules/industryrelations"
	"siakad/backend/internal/modules/master"
	"siakad/backend/internal/modules/shared"
	"siakad/backend/internal/modules/studentaffairs"
	"siakad/backend/internal/modules/usermanagement"
)

type App struct {
	Config config.Config
	Logger *slog.Logger
	DB     *sql.DB
	Server *http.Server
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	var logOutput *os.File
	if f, err := os.OpenFile("siakad.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		logOutput = f
	} else {
		logOutput = os.Stdout
	}

	logger := slog.New(slog.NewTextHandler(logOutput, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	var db *sql.DB
	if cfg.MySQL.Enabled {
		db, err = database.OpenMySQL(context.Background(), cfg.MySQL)
		if err != nil {
			return nil, fmt.Errorf("connect database: %w", err)
		}
	}

	modules := []httpserver.Module{
		auth.NewModule(db, cfg.Auth.TokenSecret, cfg.Auth.TokenTTL),
		master.NewModule(db),
		studentaffairs.NewModule(db),
		academic.NewModule(db),
		industryrelations.NewModule(db),
		shared.NewModule(db),
		usermanagement.NewModule(db),
	}

	authService := auth.NewService(cfg.Auth.TokenSecret, cfg.Auth.TokenTTL)

	server := httpserver.New(httpserver.ServerOptions{
		AppName:      cfg.App.Name,
		Environment:  cfg.App.Env,
		Address:      cfg.App.Address(),
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
		IdleTimeout:  cfg.App.IdleTimeout,
		DB:           db,
		Logger:       logger,
		Modules:      modules,
		AuthService:  authService,
	})

	return &App{
		Config: cfg,
		Logger: logger,
		DB:     db,
		Server: server,
	}, nil
}

func (a *App) Run() error {
	a.Logger.Info("starting siakad backend",
		"address", a.Config.App.Address(),
		"env", a.Config.App.Env,
		"mysql_enabled", a.Config.MySQL.Enabled,
	)

	return a.Server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	if a.DB != nil {
		defer func() {
			_ = a.DB.Close()
		}()
	}

	return a.Server.Shutdown(ctx)
}
