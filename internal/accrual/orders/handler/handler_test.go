package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mocks "github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/handler/mock"
	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/model"
	ordersvc "github.com/CimaCha/gophermart-loyal-service/internal/accrual/orders/service"
	"github.com/stretchr/testify/assert"
)

func TestCreateOrder(t *testing.T) {
	const validBody = `{"order":"12345678903","goods":[{"description":"Чайник Bork","price":7000}]}`
	tests := []struct {
		name       string
		body       string
		decodeOK   bool
		serviceErr error
		wantCode   int
		wantOrder  *model.Order
	}{
		{name: "success with goods and specification field", body: validBody, decodeOK: true, wantCode: http.StatusAccepted},
		{name: "empty body", wantCode: http.StatusBadRequest},
		{name: "malformed JSON", body: `{"order":`, wantCode: http.StatusBadRequest},
		{name: "array instead of object", body: `[]`, wantCode: http.StatusBadRequest},
		{name: "numeric order", body: `{"order":12345678903}`, wantCode: http.StatusBadRequest},
		{name: "string price", body: `{"goods":[{"price":"7000"}]}`, wantCode: http.StatusBadRequest},
		{name: "missing order", body: `{}`, decodeOK: true, serviceErr: ordersvc.ErrInvalidOrderNum, wantCode: http.StatusBadRequest, wantOrder: &model.Order{}},
		{name: "null body", body: `null`, decodeOK: true, serviceErr: ordersvc.ErrInvalidOrderNum, wantCode: http.StatusBadRequest, wantOrder: &model.Order{}},
		{name: "empty goods", body: `{"order":"12345678903","goods":[]}`, decodeOK: true, wantCode: http.StatusAccepted, wantOrder: &model.Order{OrderNum: "12345678903", Goods: []model.Good{}}},
		// SPECIFICATION.md defines 400 for malformed orders; 422 belongs to the gophermart API.
		{name: "invalid order returns specification status", body: validBody, decodeOK: true, serviceErr: ordersvc.ErrInvalidOrderNum, wantCode: http.StatusBadRequest},
		{name: "wrapped invalid order", body: validBody, decodeOK: true, serviceErr: fmt.Errorf("validate: %w", ordersvc.ErrInvalidOrderNum), wantCode: http.StatusBadRequest},
		{name: "duplicate", body: validBody, decodeOK: true, serviceErr: ordersvc.ErrOrderAlreadyProcessing, wantCode: http.StatusConflict},
		{name: "wrapped duplicate", body: validBody, decodeOK: true, serviceErr: fmt.Errorf("upload: %w", ordersvc.ErrOrderAlreadyProcessing), wantCode: http.StatusConflict},
		{name: "internal error", body: validBody, decodeOK: true, serviceErr: errors.New("database unavailable"), wantCode: http.StatusInternalServerError},
		{name: "canceled", body: validBody, decodeOK: true, serviceErr: fmt.Errorf("upload: %w", context.Canceled), wantCode: 499},
		{name: "deadline", body: validBody, decodeOK: true, serviceErr: context.DeadlineExceeded, wantCode: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewMockOrderService(t)
			h := New(slog.New(slog.NewTextHandler(io.Discard, nil)), svc)
			req := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(tt.body)).WithContext(t.Context())
			req.Header.Set("Content-Type", "application/json")
			if tt.decodeOK {
				wantOrder := model.Order{
					OrderNum: "12345678903",
					Goods:    []model.Good{{Description: "Чайник Bork", Price: 7000}},
				}
				if tt.wantOrder != nil {
					wantOrder = *tt.wantOrder
				}
				svc.EXPECT().UploadOrder(req.Context(), wantOrder).Return(tt.serviceErr).Once()
			}
			rec := httptest.NewRecorder()
			h.CreateOrder(rec, req)
			assert.Equal(t, tt.wantCode, rec.Code)
			if tt.wantCode == http.StatusAccepted || tt.wantCode == 499 {
				assert.Empty(t, rec.Body.String())
			}
		})
	}
}
