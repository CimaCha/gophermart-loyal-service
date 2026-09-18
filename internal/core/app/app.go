package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/CimaCha/gophermart-loyal-service/internal/config"
	"github.com/CimaCha/gophermart-loyal-service/internal/core/db/postgres"
	"github.com/CimaCha/gophermart-loyal-service/internal/core/httpserver"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Server *httpserver.HTTPServer
	DB     *pgxpool.Pool
}

func New(cfg *config.Config, log *slog.Logger) (*App, error) {

	// Root router
	rootRouter := chi.NewRouter()

	db, err := postgres.New(*cfg.DB, log)
	if err != nil {
		log.Error("failed to create db connection pool", "err", err)
		return nil, fmt.Errorf("database initialize: %w", err)
	}

	httpServer := httpserver.New(rootRouter, cfg.Server, log)

	// userRepo
	// orderRepo
	// balanceRepo

	// userService
	// orderService
	// balanceService

	// userHandler
	// orderHandler
	// balanceHandler

	return &App{
		Server: httpServer,
		DB:     db,
	}, nil

}

func (a *App) Run(ctx context.Context) error {

	if err := a.Server.Run(ctx); err != nil {
		return fmt.Errorf("server run: %w", err)
	}

	return nil
}
