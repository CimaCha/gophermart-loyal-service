package accrualclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestClient_GetOrder_OK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/orders/12345", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"order":"12345",
			"status":"PROCESSED",
			"accrual":500.25
		}`))
	}))
	defer ts.Close()

	client := New(Config{Address: ts.URL, Timeout: time.Second})

	resp, err := client.GetOrder(context.Background(), "12345")

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "12345", resp.Order)
	require.Equal(t, StatusProcessed, resp.Status)

	expected := decimal.NewFromFloat(500.25)
	require.NotNil(t, resp.Accrual)
	require.True(t, resp.Accrual.Equal(expected))
}

func TestClient_GetOrder_NoContent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := New(Config{Address: ts.URL, Timeout: time.Second})

	resp, err := client.GetOrder(context.Background(), "12345")

	require.NoError(t, err)
	require.Equal(t, "12345", resp.Order)
	require.Equal(t, StatusNotRegistered, resp.Status)
	require.Nil(t, resp.Accrual)
}

func TestClient_GetOrder_RateLimited(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	client := New(Config{Address: ts.URL, Timeout: time.Second})

	resp, err := client.GetOrder(context.Background(), "12345")

	var rateLimErr *RateLimitError

	require.Nil(t, resp)
	require.Error(t, err)

	require.ErrorAs(t, err, &rateLimErr)
	require.Equal(t, time.Second*60, rateLimErr.RetryAfter)
}

func TestClient_GetOrder_UnexpectedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := New(Config{Address: ts.URL, Timeout: time.Second})

	resp, err := client.GetOrder(context.Background(), "12345")

	require.Nil(t, resp)
	require.Error(t, err)
}

func TestClient_GetOrder_RequestError(t *testing.T) {
	client := New(Config{Address: "http://127.0.0.1:1", Timeout: time.Millisecond})

	resp, err := client.GetOrder(context.Background(), "12345")

	var rateLimErr *RateLimitError

	require.Nil(t, resp)
	require.Error(t, err)
	require.False(t, errors.As(err, &rateLimErr))
}
