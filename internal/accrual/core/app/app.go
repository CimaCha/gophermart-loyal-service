package app

import (
	"context"
	"fmt"
	"log/slog"

	goodsh "github.com/CimaCha/gophermart-loyal-service/internal/accrual/goods/handler"
	goodsrepo "github.com/CimaCha/gophermart-loyal-service/internal/accrual/goods/repository"
	goodssvc "github.com/CimaCha/gophermart-loyal-service/internal/accrual/goods/service"
	ordersh "github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/handler"
	ordersrepo "github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/repository"
	orderssvc "github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/service"
	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/ratelimitstore"
	"github.com/CimaCha/gophermart-loyal-service/pkg/http/httpmiddleware"

	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/config"
	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/core/deps"
	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/transport/http/router"
	migrations "github.com/CimaCha/gophermart-loyal-service/migrations/accrual"
	"github.com/CimaCha/gophermart-loyal-service/pkg/httpserver"
	"github.com/CimaCha/gophermart-loyal-service/pkg/postgres"
	"github.com/shopspring/decimal"

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

	pool, err := postgres.New(*cfg.DB, log, migrations.EmbedMigrations)
	if err != nil {
		log.Error("failed to create db connection pool", "err", err)
		return nil, fmt.Errorf("database initialize: %w", err)
	}

	httpServer := httpserver.New(rootRouter, cfg.Server, log)
	rateLimitStore := ratelimitstore.New(ctx, cfg.RLS, log)

	orderRepo := ordersrepo.New(pool)
	goodsRepo := goodsrepo.New(pool)

	orderSvc := orderssvc.New(orderRepo, log)
	goodsSvc := goodssvc.New(goodsRepo)

	orderHandler := ordersh.New(log, orderSvc)
	goodsHandler := goodsh.New(log, goodsSvc)

	dependencies := deps.New(
		orderHandler,
		goodsHandler,
		cfg,
		log,
		rateLimitStore,
	)

	rootRouter.Use(
		chimiddleware.Recoverer,
		httpmiddleware.RequestID(),
		httpmiddleware.Logging(log),
		httpmiddleware.GzipCompress(),
	)

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
