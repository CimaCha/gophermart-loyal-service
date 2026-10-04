package handler

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/model"
	ordersvc "github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/service"
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

	expected := model.NewOrder("12345678903", []model.Good{
		{
			Description: "iPhone",
			Price:       decimal.RequireFromString("1000"),
		},
	})

	svc.
		On("UploadOrder", mock.Anything, *expected).
		Return(nil).
		Once()

	req := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.CreateOrder(rec, req)

	assert.Equal(t, http.StatusAccepted, rec.Code)
	svc.AssertExpectations(t)
}

func TestHandler_CreateOrder_InvalidJSON(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := new(MockOrderService)
	h := New(logger, svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/orders",
		strings.NewReader("{"),
	)

	rec := httptest.NewRecorder()

	h.CreateOrder(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	svc.AssertNotCalled(t, "UploadOrder")
}

func TestHandler_CreateOrder_InvalidOrderNumber(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := new(MockOrderService)
	h := New(logger, svc)

	body := `{
		"order":"12345678903",
		"goods":[]
	}`

	svc.
		On("UploadOrder", mock.Anything, mock.AnythingOfType("model.Order")).
		Return(ordersvc.ErrInvalidOrderNum).
		Once()

	req := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.CreateOrder(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
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

	assert.Equal(t, "123", got.OrderNum)
	assert.Equal(t, model.OrderStatusRegistered, got.Status)
	assert.Nil(t, got.Accrual)

	require.Len(t, got.Goods, 1)

	assert.Equal(t, "Tea", got.Goods[0].Description)
	assert.True(t, price.Equal(got.Goods[0].Price))
}
