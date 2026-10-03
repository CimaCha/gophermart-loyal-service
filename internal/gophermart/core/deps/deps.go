package deps

import (
	"log/slog"

	authentication "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/auth"
	balansh "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/balance/handler"
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/config"
	orderh "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/order/handler"
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/ratelimitstore"
	userh "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/user/handler"
)

type Dependencies struct {
	UserHandler    *userh.Handler
	OrderHandler   *orderh.Handler
	BalanceHandler *balansh.Handler
	TokenSvc       *authentication.TokenService
	RLS            *ratelimitstore.RateLimitStorage
	Cfg            *config.Config
	Logger         *slog.Logger
}

func New(
	userHandler *userh.Handler,
	orderHandler *orderh.Handler,
	balanceHandler *balansh.Handler,
	tokenSvc *authentication.TokenService,
	rls *ratelimitstore.RateLimitStorage,
	cfg *config.Config,
	logger *slog.Logger,
) *Dependencies {
	return &Dependencies{
		UserHandler:    userHandler,
		OrderHandler:   orderHandler,
		BalanceHandler: balanceHandler,
		TokenSvc:       tokenSvc,
		RLS:            rls,
		Cfg:            cfg,
		Logger:         logger,
	}
}
