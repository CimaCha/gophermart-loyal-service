// Package deps предоставляет контейнер зависимостей,
// используемый для передачи инициализированных слоев и обработчиков в подсистему маршрутизации gophermart.
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

// Dependencies агрегирует в себе все HTTP-обработчики (handlers), сервисы аутентификации,
// настройки конфигурации, логгер и хранилище лимитов, необходимые для сборки API-маршрутов.
type Dependencies struct {
	// UserHandler отвечает за обработку HTTP-запросов аутентификации и регистрации пользователей.
	UserHandler *userh.Handler
	// OrderHandler отвечает за обработку HTTP-запросов загрузки и получения статуса заказов.
	OrderHandler *orderh.Handler
	// BalanceHandler отвечает за обработку HTTP-запросов проверки баланса и списания баллов.
	BalanceHandler *balansh.Handler
	// TokenSvc предоставляет логику генерации и валидации сессионных JWT-токенов.
	TokenSvc *authentication.TokenService
	// RLS предоставляет доступ к хранилищу лимитов частоты запросов.
	RLS *ratelimitstore.RateLimitStorage
	// Cfg хранит глобальные конфигурационные параметры сервиса gophermart.
	Cfg *config.Config
	// Logger предоставляет экземпляр структурированного логгера для трассировки запросов.
	Logger *slog.Logger
}

// New создает и возвращает новый заполненный экземпляр Dependencies,
// инкапсулируя переданные компоненты приложения в единый неизменяемый объект.
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
