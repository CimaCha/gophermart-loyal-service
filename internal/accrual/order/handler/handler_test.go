package handler

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/order/model"
	ordersvc "github.com/CimaCha/gophermart-loyal-service/internal/accrual/order/service"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_CreateOrder_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := NewMockOrderService(t)
	h := New(logger, svc)

	body := `{
		"order":"12345678903",
		"goods":[
			{
				"description":"iPhone",
				"price":1000
			}
		]
	}`

	svc.EXPECT().
		UploadOrder(
			mock.Anything,
			mock.MatchedBy(func(order model.Order) bool {
				return order.OrderNum == "12345678903" &&
					order.Status == model.OrderStatusRegistered &&
					len(order.Goods) == 1 &&
					order.Goods[0].Description == "iPhone" &&
					order.Goods[0].Price.Equal(decimal.RequireFromString("1000"))
			}),
		).
		Return(nil).
		Once()

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/orders",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	h.CreateOrder(rec, req)

	require.Equal(t, http.StatusAccepted, rec.Code)
}

func TestHandler_CreateOrder_InvalidJSON(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := NewMockOrderService(t)
	h := New(logger, svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/orders",
		strings.NewReader("{"),
	)

	rec := httptest.NewRecorder()

	h.CreateOrder(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_CreateOrder_InvalidOrderNumber(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := NewMockOrderService(t)
	h := New(logger, svc)

	svc.EXPECT().
		UploadOrder(
			mock.Anything,
			mock.AnythingOfType("model.Order"),
		).
		Return(ordersvc.ErrInvalidOrderNum).
		Once()

	body := `{
		"order":"12345678903",
		"goods":[]
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/orders",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	h.CreateOrder(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateOrderRequest_toModel(t *testing.T) {
	price := decimal.RequireFromString("100")

	req := CreateOrderRequest{
		OrderNum: "123",
		Goods: []GoodRequest{
			{
				Description: "Tea",
				Price:       price,
			},
		},
	}

	got := req.toModel()

	require.Equal(t, "123", got.OrderNum)
	require.Equal(t, model.OrderStatusRegistered, got.Status)
	require.Nil(t, got.Accrual)

	require.Len(t, got.Goods, 1)

	assert.Equal(t, "Tea", got.Goods[0].Description)
	assert.True(t, price.Equal(got.Goods[0].Price))
}
