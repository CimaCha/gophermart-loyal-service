package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/accrualclient"
	ordermodel "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/order/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/semaphore"
)

type OrderStore interface {
	// GetPendingOrders - возвращает заказы со статусами NEW или PROCESSING
	GetPendingOrders(ctx context.Context) ([]ordermodel.Order, error)
	// UpdateOrderResult — атомарно обновляет статус+accrual, только если заказ ещё не обработан
	UpdateOrderResultTx(ctx context.Context, tx pgx.Tx, orderNum string, status ordermodel.OrderStatus, accrual *decimal.Decimal) (updated bool, err error)
	// UpdateStatus - обовляет только статус заказа
	UpdateStatus(ctx context.Context, orderNum string, status ordermodel.OrderStatus) error
}

type BalanceAccruer interface {
	// Accrue - обновляет баланс пользователя используяю транзакцию
	AccrueTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, orderNum string, sum decimal.Decimal) error
}

type AccrualClient interface {
	// GetOrder - возвращает результат с содержанием текущего состояния заказа из accrual
	GetOrder(ctx context.Context, orderNum string) (*accrualclient.ResultResponse, error)
}

type Transactor interface {
	BeginFunc(ctx context.Context, fn func(pgx.Tx) error) error
}

type Worker struct {
	// Для взаимодействия с другими сервисами
	orders   OrderStore
	balances BalanceAccruer
	accrual  AccrualClient
	// Интерфейс для выполнения транзакции в одном блоке
	tx     Transactor
	logger *slog.Logger
	// Настройки для polling и семафора
	interval time.Duration
	trigger  chan ordermodel.Order
	sem      *semaphore.Weighted
}

func New(
	orders OrderStore,
	balances BalanceAccruer,
	accrual AccrualClient,
	tx Transactor,
	logger *slog.Logger,
	interval time.Duration,
	maxConcurrent int64,
) *Worker {
	return &Worker{
		orders: orders, balances: balances, accrual: accrual, tx: tx,
		logger: logger, interval: interval,
		trigger: make(chan ordermodel.Order, 256),
		sem:     semaphore.NewWeighted(maxConcurrent),
	}
}

// Notify — неблокирующее уведомление, вызывается сразу после успешного создания
// заказа. Принимает уже готовую структуру
func (w *Worker) Notify(order ordermodel.Order) {
	select {
	case w.trigger <- order:
	default:
		w.logger.Debug(
			"trigger buffer full, falling back to polling",
			"order_num", order.OrderNum,
		)
	}
}

// Run - запускает работу воркера
// Воркер либо принимает order из канала либо по тикеру бежит в БД и ищет NEW или PROCESSING заказы
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		select {
		case <-ctx.Done():
			return
		case order := <-w.trigger:
			w.dispatch(ctx, &wg, order)
		case <-ticker.C:
			orders, err := w.orders.GetPendingOrders(ctx)
			if err != nil {
				w.logger.Error(
					"fetch pending orders failed",
					"err", err,
				)
				continue
			}
			for _, o := range orders {
				w.dispatch(ctx, &wg, o)
			}
		}
	}
}

func (w *Worker) dispatch(ctx context.Context, wg *sync.WaitGroup, order ordermodel.Order) {
	// Acquire - блокирует горутину в случае, если в semaphore закончился ресурс
	// В качестве задающего ресурса выступает maxConcurrent поле в Worker struct
	if err := w.sem.Acquire(ctx, 1); err != nil {
		return
	}

	wg.Go(func() {
		// Release - освобождает одну единицу ресурса в seamaphore после окончания работы processOrder
		defer w.sem.Release(1)
		if err := w.processOrder(ctx, order); err != nil {
			w.logger.Error(
				"process order failed",
				"order_num", order.OrderNum,
				"err", err,
			)
		}
	})

}

// processOrder — единственная точка входа для всех заказов, из Notifier и из Polling в том числе
// если заказы попали в processOrder одновременно, это решается на уровне repository в UpdateOrderResult
func (w *Worker) processOrder(ctx context.Context, order ordermodel.Order) error {
	// Если статус заказа уже PROCESSED или INVALID (попала старая версия заказа), то мы просто промолчим ничего не меняю
	if order.Status == ordermodel.OrderStatusProcessed || order.Status == ordermodel.OrderStatusInvalid {
		return nil
	}

	result, err := w.accrual.GetOrder(ctx, order.OrderNum)
	if err != nil {
		return fmt.Errorf("get order from accrual: %w", err)
	}

	switch result.Status {
	case accrualclient.StatusNotRegistered:
		return nil
	case accrualclient.StatusRegistered, accrualclient.StatusProcessing:
		if order.Status != ordermodel.OrderStatusProcessing {
			return w.orders.UpdateStatus(ctx, order.OrderNum, ordermodel.OrderStatusProcessing)
		}
		return nil
	case accrualclient.StatusProcessed, accrualclient.StatusInvalid:
		return w.finalizeOrder(ctx, order, result)
	default:
		return fmt.Errorf("unexpected accrual status: %s", result.Status)
	}
}

// finalizeOrder - переводит заказ в конечный статус и начисляет боннусы, если статус не INVALID
func (w *Worker) finalizeOrder(ctx context.Context, order ordermodel.Order, result *accrualclient.ResultResponse) error {
	status := toInternalStatus(result.Status)

	return w.tx.BeginFunc(ctx, func(tx pgx.Tx) error {
		updated, err := w.orders.UpdateOrderResultTx(ctx, tx, order.OrderNum, status, result.Accrual)
		if err != nil {
			return err
		}
		if !updated {
			return nil
		}

		if status == ordermodel.OrderStatusProcessed && result.Accrual != nil {
			return w.balances.AccrueTx(ctx, tx, order.UserID, order.OrderNum, *result.Accrual)
		}
		return nil
	})
}

// toInternalStatus - хелпер для конвертации статуса accrual в статус, который понимает gophermart
func toInternalStatus(accrualStatus accrualclient.Status) ordermodel.OrderStatus {
	switch accrualStatus {
	case accrualclient.StatusProcessed:
		return ordermodel.OrderStatusProcessed
	case accrualclient.StatusInvalid:
		return ordermodel.OrderStatusInvalid
	case accrualclient.StatusProcessing, accrualclient.StatusRegistered:
		return ordermodel.OrderStatusProcessing
	default:
		return ordermodel.OrderStatusNew
	}
}
