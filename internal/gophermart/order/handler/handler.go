package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/order/model"
	ordersvc "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/order/service"
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/transport/ctxkeys"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderService interface {
	UploadOrder(ctx context.Context, orderNumStr string, uid uuid.UUID) error
	GetOrders(ctx context.Context, uid uuid.UUID) ([]model.Order, error)
}

type Handler struct {
	l   *slog.Logger
	svc OrderService
}

func New(
	log *slog.Logger,
	service OrderService,
) *Handler {
	return &Handler{
		l:   log,
		svc: service,
	}
}

type OrderResponse struct {
	OrderNum   string           `json:"number"`
	Status     string           `json:"status"`
	Accrual    *decimal.Decimal `json:"accrual,omitempty"`
	UploadedAt time.Time        `json:"uploaded_at"`
}

const maxBodySize = 32

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	uid, err := ctxkeys.GetUserID(r.Context())
	if err != nil {
		h.l.Error(
			"failed to get user ID from context",
			"err", err,
		)
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}

	orders, err := h.svc.GetOrders(r.Context(), uid)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			h.l.Debug("create short url canceled by client")
			w.WriteHeader(499)
			return
		default:
			h.l.Error(
				"unexpected error during orders loading",
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

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := make([]OrderResponse, len(orders))

	for i, o := range orders {
		resp[i] = toOrderResponse(o)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.l.Error(
			"failed to marshal response json",
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

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {

	// Защита от DDos атак если отправляют запрос с телом размера в гигабайты
	body := http.MaxBytesReader(
		w,
		r.Body,
		maxBodySize,
	)

	data, err := io.ReadAll(body)
	if err != nil || len(data) == 0 {
		h.l.Info(
			"failed to read body",
			"err", err,
		)
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	uidStr, err := ctxkeys.GetUserID(r.Context())
	if err != nil {
		h.l.Error(
			"failed to get user ID from context",
			"err", err,
		)
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}

	orderNumStr := strings.TrimSpace(string(data))

	if err := h.svc.UploadOrder(r.Context(), orderNumStr, uidStr); err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			h.l.Debug("create short url canceled by client")
			w.WriteHeader(499)
			return
		case errors.Is(err, ordersvc.ErrInvalidOrderNum):
			h.l.Info(
				"invalid order number",
				"order", orderNumStr,
				"err", err,
			)
			http.Error(
				w,
				http.StatusText(http.StatusUnprocessableEntity),
				http.StatusUnprocessableEntity,
			)
			return
		case errors.Is(err, ordersvc.ErrOrderAlreadyCreatedByAnotherUser):
			h.l.Info(
				"order already uploaded by another user",
				"err", err,
			)
			http.Error(
				w,
				http.StatusText(http.StatusConflict),
				http.StatusConflict,
			)
			return
		case errors.Is(err, ordersvc.ErrOrderAlreadyProcessing):
			h.l.Info(
				"order has already been uploaded by this user",
				"err", err,
			)
			w.WriteHeader(http.StatusOK)
			return
		default:
			h.l.Error(
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

func toOrderResponse(o model.Order) OrderResponse {
	return OrderResponse{
		OrderNum:   strconv.FormatInt(o.OrderNum, 10),
		Status:     o.Status,
		Accrual:    o.Accrual,
		UploadedAt: o.UploadedAt,
	}
}
