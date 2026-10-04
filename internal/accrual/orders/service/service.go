// Package service предоставляет слой бизнес-логики для управления заказами,
// включая их валидацию, регистрацию в постоянном хранилище и отправку уведомлений воркеру для расчета баллов.
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/model"
	orderrepo "github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/repository"
	"github.com/google/uuid"
)

// OrderNotifier определяет интерфейс для асинхронного оповещения фоновых систем (воркеров)
// о появлении нового зарегистрированного заказа, готового к расчету вознаграждений.
type OrderNotifier interface {
	// Notify передает заказ в систему обработки и вычисления баллов лояльности.
	Notify(order model.Order)
}

// OrderRepository определяет интерфейс взаимодействия со слоем персистентного хранения данных заказов.
type OrderRepository interface {
	// CreateOrder выполняет атомарное сохранение структуры заказа и его товарных позиций.
	CreateOrder(ctx context.Context, order *model.Order) error
	// GetOrder осуществляет поиск и извлечение информации о заказе по его строковому номеру.
	GetOrder(ctx context.Context, orderID string) (model.Order, error)
}

// OrderService инкапсулирует репозиторий, логгер и систему нотификации
// для обеспечения централизованных бизнес-сценариев работы с заказами.
type OrderService struct {
	repo     OrderRepository
	logger   *slog.Logger
	notifier OrderNotifier
}

var (
	// ErrInvalidOrderNum возвращается сервисным слоем, если номер заказа не прошел валидацию (алгоритм Луна).
	ErrInvalidOrderNum = errors.New("invalid order number")
	// ErrOrderAlreadyProcessing возвращается, если заказ с данным номером уже был ранее загружен в систему.
	ErrOrderAlreadyProcessing = errors.New("order has already been uploaded by user")
	// ErrOrdersNotFound возвращается, если искомые заказы отсутствуют в базе данных.
	ErrOrdersNotFound = errors.New("orders not found")
)

// New создает и инициализирует новый экземпляр OrderService с необходимыми внешними зависимостями.
func New(
	repo OrderRepository,
	log *slog.Logger,
	notifier OrderNotifier,
) *OrderService {
	return &OrderService{
		repo:     repo,
		logger:   log,
		notifier: notifier,
	}
}

func (s *OrderService) GetOrder(ctx context.Context, uid uuid.UUID) (model.Order, error) {
	//TODO
	return model.Order{}, nil
}

// UploadOrder выполняет полный бизнес-сценарий регистрации нового заказа:
// 1. Создает доменную модель и валидирует корректность контрольной суммы номера заказа.
// 2. Персистентно сохраняет заказ через репозиторий.
// 3. Отправляет асинхронное уведомление через notifier для немедленного начисления баллов воркером.
func (s *OrderService) UploadOrder(ctx context.Context, order model.Order) error {

	newOrder, err := model.NewOrder(order.OrderNum, order.Goods).Validate()
	if err != nil {
		return ErrInvalidOrderNum
	}

	if err = s.repo.CreateOrder(ctx, newOrder); err != nil {
		switch {
		case errors.Is(err, orderrepo.ErrOrderAlreadyProcessing):
			return ErrOrderAlreadyProcessing
		default:
			return fmt.Errorf("repository create order: %w", err)
		}
	}

	s.notifier.Notify(*newOrder)

	return nil
}
