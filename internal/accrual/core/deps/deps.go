package deps

import (
	"log/slog"

	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/config"
	goodsh "github.com/CimaCha/gophermart-loyal-service/internal/accrual/goods/handler"
	ordersh "github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/handler"
)

type Dependencies struct {
	OrdersHandler *ordersh.Handler
	GoodsHandler  *goodsh.Handler
	Cfg           *config.Config
	Logger        *slog.Logger
}

func New(
	ordersHandler *ordersh.Handler,
	goodsHandler *goodsh.Handler,
	cfg *config.Config,
	logger *slog.Logger,
) *Dependencies {
	return &Dependencies{
		OrdersHandler: ordersHandler,
		GoodsHandler:  goodsHandler,
		Cfg:           cfg,
		Logger:        logger,
	}
}
