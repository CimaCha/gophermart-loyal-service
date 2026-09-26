package deps

import (
	"log/slog"

	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/config"
	goodsh "github.com/CimaCha/gophermart-loyal-service/internal/accrual/goods/handler"
	ordersh "github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/handler"
	"github.com/CimaCha/gophermart-loyal-service/internal/shared/ratelimit"
)

type Dependencies struct {
	OrdersHandler *ordersh.Handler
	GoodsHandler  *goodsh.Handler
	Limiter       *ratelimit.RateLimitStorage
	Cfg           *config.Config
	Logger        *slog.Logger
}

func New(
	ordersHandler *ordersh.Handler,
	goodsHandler *goodsh.Handler,
	limiter *ratelimit.RateLimitStorage,
	cfg *config.Config,
	logger *slog.Logger,
) *Dependencies {
	return &Dependencies{
		OrdersHandler: ordersHandler,
		GoodsHandler:  goodsHandler,
		Limiter:       limiter,
		Cfg:           cfg,
		Logger:        logger,
	}
}
