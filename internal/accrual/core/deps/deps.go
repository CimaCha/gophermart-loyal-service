// Package deps предоставляет контейнер зависимостей,
// используемый для удобной передачи инициализированных слоев и компонентов в подсистему маршрутизации.
package deps

import (
	"log/slog"

	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/config"
	goodsh "github.com/CimaCha/gophermart-loyal-service/internal/accrual/goods/handler"
	ordersh "github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/handler"
	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/ratelimitstore"
)

// Dependencies объединяет в себе все HTTP-обработчики (handlers), настройки конфигурации,
// логгер и хранилище лимитов, необходимые для конфигурирования эндпоинтов API.
type Dependencies struct {
	// OrdersHandler отвечает за обработку входящих HTTP-запросов, связанных с заказами.
	OrdersHandler *ordersh.Handler
	// GoodsHandler отвечает за обработку входящих HTTP-запросов, связанных с товарами и правилами начисления.
	GoodsHandler *goodsh.Handler
	// Cfg хранит глобальную конфигурацию приложения.
	Cfg *config.Config
	// Logger предоставляет доступ к структурированному логированию.
	Logger *slog.Logger
	// RLS предоставляет доступ к хранилищу состояний лимитов запросов.
	RLS *ratelimitstore.RateLimitStorage
}

// New создает и возвращает новый заполненный экземпляр Dependencies,
// инкапсулируя все переданные слои приложения в единый объект.
func New(
	ordersHandler *ordersh.Handler,
	goodsHandler *goodsh.Handler,
	cfg *config.Config,
	logger *slog.Logger,
	rls *ratelimitstore.RateLimitStorage,
) *Dependencies {
	return &Dependencies{
		OrdersHandler: ordersHandler,
		GoodsHandler:  goodsHandler,
		Cfg:           cfg,
		Logger:        logger,
		RLS:           rls,
	}
}
