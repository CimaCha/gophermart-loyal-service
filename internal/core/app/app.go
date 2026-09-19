package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/CimaCha/gophermart-loyal-service/internal/config"
	"github.com/CimaCha/gophermart-loyal-service/internal/core/auth/jwtoken"
	"github.com/CimaCha/gophermart-loyal-service/internal/core/db/postgres"
	"github.com/CimaCha/gophermart-loyal-service/internal/core/httpserver"
	"github.com/CimaCha/gophermart-loyal-service/internal/core/transport/http/middleware"
	orderh "github.com/CimaCha/gophermart-loyal-service/internal/order/handler"
	orderrepo "github.com/CimaCha/gophermart-loyal-service/internal/order/repository"
	ordersvc "github.com/CimaCha/gophermart-loyal-service/internal/order/service"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Server  *httpserver.HTTPServer
	Pgxpool *pgxpool.Pool
}

func New(cfg *config.Config, log *slog.Logger) (*App, error) {

	// Root router
	rootRouter := chi.NewRouter()

	pool, err := postgres.New(*cfg.DB, log)
	if err != nil {
		log.Error("failed to create db connection pool", "err", err)
		return nil, fmt.Errorf("database initialize: %w", err)
	}

	httpServer := httpserver.New(rootRouter, cfg.Server, log)

	// Потом убрать в .env
	tokenService := jwtoken.NewTokenService([]byte("our-secret-key"))

	// userRepo
	orderRepo := orderrepo.New(pool)
	// balanceRepo

	// userService
	orderSvc := ordersvc.New(orderRepo, log)
	// balanceService

	// userHandler
	orderHandler := orderh.New(orderSvc, log)
	// balanceHandler

	// Регаем приватные маршруты для order handler
	rootRouter.Group(func(r chi.Router) {
		r.Use(middleware.Auth(tokenService, log))

		orderHandler.RegisterPrivateAPI(r)
	})

	return &App{
		Server:  httpServer,
		Pgxpool: pool,
	}, nil

}

func (a *App) Run(ctx context.Context) error {

	if err := a.Server.Run(ctx); err != nil {
		return fmt.Errorf("server run: %w", err)
	}

	return nil
}
