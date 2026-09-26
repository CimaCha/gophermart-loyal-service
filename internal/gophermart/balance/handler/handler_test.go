package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/balance/model"
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/transport/ctxkeys"
)

func init() {
	// decimal в JSON — числом, а не строкой.
	decimal.MarshalJSONWithoutQuotes = true
}

// mustDecimal парсит строку в decimal.Decimal, паникует при ошибке.
func mustDecimal(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

// newTestLogger возвращает logger, который пишет в никуда.
func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestHandler_GetBalance_Success(t *testing.T) {
	userID := uuid.New()
	expected := model.Balance{
		UserID:    userID,
		Current:   mustDecimal("500.5"),
		Withdrawn: mustDecimal("42"),
	}

	svc := NewMockBalanceService(t)
	svc.On("GetBalance", mock.Anything, userID).Return(expected, nil)

	h := New(newTestLogger(), svc)

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req = req.WithContext(ctxkeys.WithUserID(req.Context(), userID))
	rec := httptest.NewRecorder()

	h.GetBalance(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var got model.Balance
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&got))
	assert.True(t, expected.Current.Equal(got.Current), "current mismatch")
	assert.True(t, expected.Withdrawn.Equal(got.Withdrawn), "withdrawn mismatch")
	svc.AssertExpectations(t)
}

func TestHandler_GetBalance_NoUserID(t *testing.T) {
	svc := NewMockBalanceService(t)
	h := New(newTestLogger(), svc)

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	rec := httptest.NewRecorder()

	h.GetBalance(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_GetBalance_ServiceError(t *testing.T) {
	userID := uuid.New()
	svc := NewMockBalanceService(t)
	svc.On("GetBalance", mock.Anything, userID).
		Return(model.Balance{}, errors.New("boom"))

	h := New(newTestLogger(), svc)

	req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req = req.WithContext(ctxkeys.WithUserID(req.Context(), userID))
	rec := httptest.NewRecorder()

	h.GetBalance(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	svc.AssertExpectations(t)
}
