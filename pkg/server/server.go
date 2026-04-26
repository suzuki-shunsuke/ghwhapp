package server

import (
	"context"
	"log/slog"

	"github.com/suzuki-shunsuke/ghwhapp/pkg/config"
	"github.com/suzuki-shunsuke/ghwhapp/pkg/controller"
)

type Server struct {
	logger     *slog.Logger
	config     *config.Config
	controller Controller
}

type Controller interface {
	Run(ctx context.Context, logger *slog.Logger, req *controller.Request) error
}

func New(logger *slog.Logger, ctrl Controller, cfg *config.Config) (*Server, error) {
	return &Server{
		logger:     logger,
		config:     cfg,
		controller: ctrl,
	}, nil
}
