package app

import (
	"context"
	"fmt"
	"log/slog"

	authentication "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/auth"
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/config"
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/core/deps"
	orderh "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/order/handler"
	orderrepo "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/order/repository"
	ordersvc "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/order/service"
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/transport/http/middleware"
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/transport/http/router"
	userh "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/user/handler"
	userrepo "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/user/repository"
	usersvc "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/user/service"
	"github.com/CimaCha/gophermart-loyal-service/internal/shared/db/postgres"
	"github.com/CimaCha/gophermart-loyal-service/internal/shared/httpserver"
	"github.com/CimaCha/gophermart-loyal-service/internal/shared/ratelimit"
	"github.com/CimaCha/gophermart-loyal-service/pkg/passhasher"
	"github.com/shopspring/decimal"

	balanceh "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/balance/handler"
	balancerepo "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/balance/repository"
	balancesvc "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/balance/service"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Server  *httpserver.HTTPServer
	Pgxpool *pgxpool.Pool
}

func New(ctx context.Context, cfg *config.Config, log *slog.Logger) (*App, error) {

	// Отключаем кавычки при JSON-сериализации decimal.Decimal,
	// чтобы баланс отдавался числом (500.5), а не строкой ("500.5").
	decimal.MarshalJSONWithoutQuotes = true
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
	balanceRepo := balancerepo.New(pool)

	userSvc := usersvc.New(userRepo, tokenBuilder, passHasher)
	orderSvc := ordersvc.New(orderRepo, log)
	balanceSvc := balancesvc.New(balanceRepo)

	userHandler := userh.New(log, userSvc)
	orderHandler := orderh.New(log, orderSvc)
	balanceHandler := balanceh.New(log, balanceSvc)

	dependencies := deps.New(
		userHandler,
		orderHandler,
		balanceHandler,
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
