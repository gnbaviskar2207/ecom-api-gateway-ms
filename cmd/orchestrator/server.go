package main

import (
	"log/slog"
	"net/http"

	"github.com/gnbaviskar2207/ecom-api-gateway-ms/internal/config"
)

type Server struct {
	logger *slog.Logger
	cfg    *config.Config
	apiMux *http.ServeMux
}

func New(cfg *config.Config, logger *slog.Logger) *Server {
	return &Server{
		cfg:    cfg,
		logger: logger,
		apiMux: http.NewServeMux(),
	}
}
