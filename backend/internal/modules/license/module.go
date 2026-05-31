package license

import (
	"database/sql"
	"net/http"

	"siakad/backend/internal/config"
)

type Module struct {
	handler *Handler
}

func NewModule(db *sql.DB, cfg config.LicenseConfig) *Module {
	repo := NewRepository(db)
	client := NewGeneratorClient(cfg.GeneratorURL, cfg.APIKey)
	handler := NewHandler(repo, client)
	return &Module{handler: handler}
}

func (m *Module) Name() string {
	return "license"
}

func (m *Module) BasePath() string {
	return "/api/v1/license"
}

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	m.handler.RegisterRoutes(mux)
}
