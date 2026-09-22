package app

import (
	"context"
	"fmt"
	"log/slog"

	authentication "github.com/CimaCha/gophermart-loyal-service/internal/auth"
	"github.com/CimaCha/gophermart-loyal-service/internal/config"
	"github.com/CimaCha/gophermart-loyal-service/internal/core/deps"
	"github.com/CimaCha/gophermart-loyal-service/internal/db/postgres"
	"github.com/CimaCha/gophermart-loyal-service/internal/httpserver"
	orderh "github.com/CimaCha/gophermart-loyal-service/internal/order/handler"
	orderrepo "github.com/CimaCha/gophermart-loyal-service/internal/order/repository"
	ordersvc "github.com/CimaCha/gophermart-loyal-service/internal/order/service"
	"github.com/CimaCha/gophermart-loyal-service/internal/ratelimit"
	"github.com/CimaCha/gophermart-loyal-service/internal/transport/http/middleware"
	"github.com/CimaCha/gophermart-loyal-service/internal/transport/http/router"
	userh "github.com/CimaCha/gophermart-loyal-service/internal/user/handler"
	userrepo "github.com/CimaCha/gophermart-loyal-service/internal/user/repository"
	usersvc "github.com/CimaCha/gophermart-loyal-service/internal/user/service"
	"github.com/CimaCha/gophermart-loyal-service/pkg/passhasher"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Server  *httpserver.HTTPServer
	Pgxpool *pgxpool.Pool
}

func New(ctx context.Context, cfg *config.Config, log *slog.Logger) (*App, error) {

	// Root router
	rootRouter := chi.NewRouter()

	pool, err := postgres.New(*cfg.DB, log)
	if err != nil {
		log.Error("failed to create db connection pool", "err", err)
		return nil, fmt.Errorf("database initialize: %w", err)
	}

	httpServer := httpserver.New(rootRouter, cfg.Server, log)

	limiter := ratelimit.NewRLS(ctx, cfg.RLS, log)
	tokenBuilder := authentication.NewJWTBuilder([]byte("temp-secret-key"))
	tokenValidator := authentication.NewUserIDParser([]byte("temp-secret-key"))
	passHasher := new(passhasher.Argon2Hasher)

	userRepo := userrepo.New(pool)
	orderRepo := orderrepo.New(pool)
	// balanceRepo

	userSvc := usersvc.New(userRepo, tokenBuilder, passHasher)
	orderSvc := ordersvc.New(orderRepo, log)
	// balanceService

	userHandler := userh.New(log, userSvc)
	orderHandler := orderh.New(log, orderSvc)
	// balanceHandler

	dependencies := deps.New(
		userHandler,
		orderHandler,
		tokenValidator,
		limiter,
		cfg,
		log,
	)

	rootRouter.Use(chimiddleware.Recoverer, middleware.RequestID(), middleware.Logging(log))

	// Регистрируем все маршруты здесь
	router.SetupRoutes(rootRouter, dependencies)

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
