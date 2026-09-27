// accrualclient пакет для взаимодействия с accrual через собственный client
package accrualclient

import (
	"context"
	"errors"
	"fmt"
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

// ResultResponse - ответ от accrual сервиса, который содержит результат обработки заказа
type ResultResponse struct {
	// Номер заказа
	Order string `json:"order"`
	// Текущий статус обработки заказа
	Status Status `json:"status"`
	// Если расчёт был окончен - Accrual будет содержать количество баллов
	// Которые необходимы к начислению
	// В ином случае Accrual будет nil
	Accrual *decimal.Decimal `json:"accrual,omitempty"`
}

var (
	// ErrRateLimited возвращается при превышении лимита запросов к accrual (429)
	ErrRateLimited = errors.New("too many requests")
)

func (s Status) String() string {
	return string(s)
}

// Client предоставляет HTTP-клиента для взаимодействия с accrual сервисом
type Client struct {
	httpClient *resty.Client
}

// New создаёт экземпляр HTTP-клиента
// baseURL - необходим для того, чтобы клиент знал по какому адресу находится accrual
// timeout - необходим для возможности прервать запрос, если сервер долго не отвечает клиенту
func New(baseURL string, timeout time.Duration) *Client {
	client := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(timeout).
		SetRetryCount(0)

	return &Client{httpClient: client}
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
		return nil, fmt.Errorf("%w: retry after %s", ErrRateLimited, retryAfter)
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

}
