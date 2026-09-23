package deps

import (
	"log/slog"

	authentication "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/auth"
	balansh "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/balance/handler"
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/config"
	orderh "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/order/handler"
	userh "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/user/handler"
	"github.com/CimaCha/gophermart-loyal-service/internal/shared/ratelimit"
)

type Dependencies struct {
	UserHandler    *userh.Handler
	OrderHandler   *orderh.Handler
	BalanceHandler *balansh.Handler
	Tv             *authentication.Parser
	Limiter        *ratelimit.RateLimitStorage
	Cfg            *config.Config
	Logger         *slog.Logger
}

func New(
	userHandler *userh.Handler,
	orderHandler *orderh.Handler,
	balanceHandler *balansh.Handler,
	tokenValidator *authentication.Parser,
	limiter *ratelimit.RateLimitStorage,
	cfg *config.Config,
	logger *slog.Logger,
) *Dependencies {
	return &Dependencies{
		UserHandler:    userHandler,
		OrderHandler:   orderHandler,
		BalanceHandler: balanceHandler,
		Tv:             tokenValidator,
		Limiter:        limiter,
		Cfg:            cfg,
		Logger:         logger,
	}
}
