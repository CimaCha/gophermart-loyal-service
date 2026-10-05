// Package handler предоставляет HTTP-обработчики для управления заказами
// и обработки запросов на расчет баллов лояльности.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/order/model"
	ordersvc "github.com/CimaCha/gophermart-loyal-service/internal/accrual/order/service"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// OrderService определяет интерфейс бизнес-логики для управления заказами.
// Этот интерфейс должен быть реализован сервисным слоем пакета orders.
type OrderService interface {
	// UploadOrder регистрирует новый заказ в системе для последующего расчета баллов.
	UploadOrder(ctx context.Context, order model.Order) error
	// GetOrder возвращает детальную информацию о заказе по его уникальному UUID-идентификатору.
	GetOrder(ctx context.Context, uid uuid.UUID) (model.Order, error)
}

// Handler инкапсулирует в себе структурированный логгер и сервисную логику
// для обработки входящих HTTP-запросов, связанных с заказами.
type Handler struct {
	logger  *slog.Logger
	service OrderService
}

// New создает и инициализирует новый экземпляр Handler с необходимыми зависимостями.
func New(
	log *slog.Logger,
	service OrderService,
) *Handler {
	return &Handler{
		logger:  log,
		service: service,
	}
}

// GoodRequest представляет входящую JSON-структуру для описания
// одной товарной позиции внутри запроса на создание заказа.
type GoodRequest struct {
	// Description содержит текстовое описание или наименование товара.
	Description string `json:"description"`
	// Price определяет стоимость единицы товара (число произвольной точности).
	Price decimal.Decimal `json:"price"`
}

// CreateOrderRequest описывает структуру тела входящего HTTP-запроса (JSON)
// для регистрации нового заказа, содержащего список товаров.
type CreateOrderRequest struct {
	// OrderNum содержит уникальный строковый номер заказа.
	OrderNum string `json:"order"`
	// Goods содержит перечень товарных позиций, входящих в этот заказ.
	Goods []GoodRequest `json:"goods"`
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	//TODO
}

// CreateOrder обрабатывает HTTP-запрос на регистрацию нового заказа для расчета баллов лояльности.
// Декодирует JSON-тело, выполняет маппинг в доменную модель и транслирует ошибки бизнес-логики в HTTP-статусы:
//   - 202 StatusAccepted — заказ успешно принят в обработку.
//   - 400 Bad Request — невалидный формат JSON или некорректный номер заказа (ошибка Луна).
//   - 409 Conflict — заказ с таким номером уже обрабатывается или был обработан ранее.
//   - 499 Client Closed Request — обработка запроса была прервана клиентом.
//   - 500 Internal Server Error — непредвиденная внутренняя ошибка сервера.
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	order := req.toModel()

	if err := h.service.UploadOrder(r.Context(), order); err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			h.logger.Debug("create order canceled by client")
			w.WriteHeader(499)
			return
		case errors.Is(err, ordersvc.ErrInvalidOrderNum):
			h.logger.Info(
				"invalid order number",
				"order", req.OrderNum,
				"err", err,
			)
			http.Error(
				w,
				http.StatusText(http.StatusBadRequest),
				http.StatusBadRequest,
			)
			return
		case errors.Is(err, ordersvc.ErrOrderAlreadyProcessing):
			h.logger.Info(
				"order already uploaded by user",
				"err", err,
			)
			http.Error(
				w,
				http.StatusText(http.StatusConflict),
				http.StatusConflict,
			)
			return
		default:
			h.logger.Error(
				"unexpected error during order saving",
				"err", err,
			)

			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
			return
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

func (r CreateOrderRequest) toModel() model.Order {
	goods := make([]model.Good, 0, len(r.Goods))

	for _, g := range r.Goods {
		goods = append(goods, model.Good{
			Description: g.Description,
			Price:       g.Price,
		})
	}

	return *model.NewOrder(r.OrderNum, goods)
}
