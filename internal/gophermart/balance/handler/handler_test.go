package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/balance/model"
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/transport/ctxkeys"
)

// TestHandler_GetBalance_Success проверяет успешный сценарий:
// userID есть в контексте, сервис вернул баланс, хендлер отдал 200 и JSON.
func TestHandler_GetBalance_Success(t *testing.T) {
	userID := uuid.New()
	expected := model.Balance{Current: 500.5, Withdrawn: 42}

	svc := NewMockBalanceService(t)
	svc.On("GetBalance", mock.Anything, userID).Return(expected, nil)

	h := New(slog.Default(), svc)

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req = req.WithContext(ctxkeys.WithUserID(req.Context(), userID))
	rec := httptest.NewRecorder()

	h.GetBalance(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var got model.Balance
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.Equal(t, expected, got)
	svc.AssertExpectations(t)
}

// TestHandler_GetBalance_NoUserID проверяет, что без userID в контексте
// хендлер возвращает 401.
func TestHandler_GetBalance_NoUserID(t *testing.T) {

	svc := NewMockBalanceService(t)
	h := New(slog.Default(), svc)

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	rec := httptest.NewRecorder()

	h.GetBalance(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestHandler_GetBalance_ServiceError проверяет, что ошибка сервиса
// маппится в 500.
func TestHandler_GetBalance_ServiceError(t *testing.T) {
	userID := uuid.New()
	svc := NewMockBalanceService(t)
	svc.On("GetBalance", mock.Anything, userID).
		Return(model.Balance{}, errors.New("boom"))

	h := New(slog.Default(), svc)

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req = req.WithContext(ctxkeys.WithUserID(req.Context(), userID))
	rec := httptest.NewRecorder()

	h.GetBalance(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	svc.AssertExpectations(t)
}
