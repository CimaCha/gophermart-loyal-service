// accrualclient пакет для взаимодействия с accrual через собственный client
package accrualclient

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal"

	// Используется для реализации удобного HTTP-клиента
	"resty.dev/v3"
)

// Status отражает текущее состояние обработки заказа на стороне accrual
type Status string

const (
	// StatusNotRegistered - заказ не зарегистрирован в системе accrual
	// Статус присваивается в том случае, если accrual вернул StatusCode 204
	// Следовательно заказ в системе gophermart не обновляется и остаётся в статусе NEW
	StatusNotRegistered Status = "NOT_REGISTERED"

	// StatusRegistered - заказ зарегистрирован в системе accrual
	// Но начисление баллов ещё не рассчитано
	// Если заказ в системе gophermart находится в статусе NEW
	// То статус заказа обновляется в PROCESSING
	StatusRegistered Status = "REGISTERED"

	// StatusProcessing - заказ зарегистрирован в системе accrual
	// И проходит расчёт баллов
	StatusProcessing Status = "PROCESSING"

	// StatusInvalid - заказ попал в систему accrual, но не был принят к расчёту
	// В таком случае вознаграждение начислено не будет
	StatusInvalid Status = "INVALID"

	// StatusProcessed - заказ успешно прошёл обработку и расчёт начисления окончен
	// Следовательно gophermart обновит статус заказа и пополнит баланс баллов
	StatusProcessed Status = "PROCESSED"
)

const (
	defaultRetryAfter = time.Minute
	minRetryAfter     = time.Second
)

// ResultResponse - ответ от accrual сервиса, который содержит результат обработки заказа
type ResultResponse struct {
	// Номер заказа
	Order string `json:"order"`
	// Текущий статус обработки заказа из accrual
	Status Status `json:"status"`
	// Если расчёт был окончен - Accrual будет содержать количество баллов
	// Которые необходимы к начислению
	// В ином случае Accrual будет nil
	Accrual *decimal.Decimal `json:"accrual,omitempty"`
}

// RateLimitError - кастомная ошибка возварщаемая наверх для обработки RetryAfter
type RateLimitError struct {
	RetryAfter time.Duration
}

// Client предоставляет HTTP-клиента для взаимодействия с accrual сервисом
type Client struct {
	httpClient *resty.Client
	logger     *slog.Logger
}

func (s Status) String() string {
	return string(s)
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("accrual: rate limited, retry after %s", e.RetryAfter)
}

// New создаёт HTTP-клиент с тремя повторами запросов при ответах 5xx.
// cfg задаёт адрес accrual и таймаут одной попытки запроса.
// logger используется для предупреждений о некорректном Retry-After.
func New(cfg Config, logger *slog.Logger) *Client {
	client := resty.New().
		SetBaseURL(cfg.Address).
		SetTimeout(cfg.Timeout).
		SetRetryCount(3).
		SetRetryWaitTime(time.Second).
		SetRetryMaxWaitTime(5 * time.Second).
		SetRetryDefaultConditions(false).
		AddRetryConditions(func(resp *resty.Response, err error) bool {
			return err == nil && resp != nil && resp.StatusCode() >= 500 && resp.StatusCode() < 600
		})

	return &Client{httpClient: client, logger: logger}
}

// SetRetryGate подключает общий gate воркера к повторным HTTP-попыткам до запуска клиента.
func (c *Client) SetRetryGate(wait func(context.Context) error) {
	c.httpClient.AddRequestMiddleware(func(_ *resty.Client, request *resty.Request) error {
		if request.Attempt > 1 {
			return wait(request.Context())
		}
		return nil
	})
}

// GetOrder - часть интерфейса для взаимодействия с accrual сервисом
// Получает результат обработки указанного в параметрах заказа
func (c *Client) GetOrder(ctx context.Context, orderNum string) (*ResultResponse, error) {
	var result ResultResponse

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetPathParam("number", orderNum).
		SetResult(&result).
		Get("/api/orders/{number}")

	if err != nil {
		return nil, fmt.Errorf("accrual request: %w", err)
	}

	switch resp.StatusCode() {
	case 200:
		return &result, nil
	case 204:
		return &ResultResponse{
			Order:  orderNum,
			Status: StatusNotRegistered,
		}, nil
	case 429:
		retryAfter := resp.Header().Get("Retry-After")
		return nil, &RateLimitError{
			RetryAfter: parseRetryAfter(retryAfter, c.logger),
		}
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

}

func parseRetryAfter(h string, logger *slog.Logger) time.Duration {
	d := defaultRetryAfter
	if secs, err := strconv.Atoi(h); err == nil && secs >= 0 {
		d = time.Duration(secs) * time.Second
	} else if t, err := http.ParseTime(h); err == nil {
		d = time.Until(t)
	} else {
		logger.Warn("invalid accrual Retry-After header, using default delay", "retry_after", h)
	}
	return max(d, minRetryAfter)
}
