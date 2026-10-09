package accrualclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"resty.dev/v3"
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

	client := New(Config{Address: ts.URL, Timeout: time.Second}, slog.New(slog.DiscardHandler))

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

	client := New(Config{Address: ts.URL, Timeout: time.Second}, slog.New(slog.DiscardHandler))

	resp, err := client.GetOrder(context.Background(), "12345")

	require.NoError(t, err)
	require.Equal(t, "12345", resp.Order)
	require.Equal(t, StatusNotRegistered, resp.Status)
	require.Nil(t, resp.Accrual)
}

func TestClient_GetOrder_RateLimited(t *testing.T) {
	for _, header := range []string{"60", "tomorrow"} {
		t.Run(header, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Retry-After", header)
				w.WriteHeader(http.StatusTooManyRequests)
			}))
			defer ts.Close()
			var logs bytes.Buffer
			client := New(Config{Address: ts.URL, Timeout: time.Second}, slog.New(slog.NewJSONHandler(&logs, nil)))
			resp, err := client.GetOrder(context.Background(), "12345")
			var rateLimErr *RateLimitError
			require.Nil(t, resp)
			require.ErrorAs(t, err, &rateLimErr)
			require.Equal(t, time.Minute, rateLimErr.RetryAfter)
			if header == "tomorrow" {
				var entry map[string]any
				require.NoError(t, json.Unmarshal(logs.Bytes(), &entry))
				require.Equal(t, "WARN", entry["level"])
				require.Equal(t, header, entry["retry_after"])
			} else {
				require.Empty(t, logs.String())
			}
		})
	}
}

func TestClient_GetOrder_UnexpectedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := New(Config{Address: ts.URL, Timeout: time.Second}, slog.New(slog.DiscardHandler))

	client.httpClient.SetRetryWaitTime(time.Millisecond).SetRetryMaxWaitTime(time.Millisecond)

	resp, err := client.GetOrder(context.Background(), "12345")

	require.Nil(t, resp)
	require.Error(t, err)
}

func TestClient_GetOrder_RequestError(t *testing.T) {
	client := New(Config{Address: "http://127.0.0.1:1", Timeout: time.Millisecond}, slog.New(slog.DiscardHandler))

	resp, err := client.GetOrder(context.Background(), "12345")

	var rateLimErr *RateLimitError

	require.Nil(t, resp)
	require.Error(t, err)
	require.False(t, errors.As(err, &rateLimErr))
}

func TestClient_GetOrder_Retries(t *testing.T) {
	tests := []struct {
		name        string
		statuses    []int
		wantCalls   int32
		wantErr     bool
		rateLimited bool
	}{
		{name: "recovers from 500 and 503", statuses: []int{500, 503, 200}, wantCalls: 3},
		{name: "stops after three retries", statuses: []int{502}, wantCalls: 4, wantErr: true},
		{name: "retries 501", statuses: []int{501, 200}, wantCalls: 2},
		{name: "does not retry 400", statuses: []int{400}, wantCalls: 1, wantErr: true},
		{name: "does not retry 429", statuses: []int{429}, wantCalls: 1, wantErr: true, rateLimited: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int32
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				attempt := int(calls.Add(1)) - 1
				status := tt.statuses[min(attempt, len(tt.statuses)-1)]
				w.Header().Set("Content-Type", "application/json")
				if status == http.StatusTooManyRequests {
					w.Header().Set("Retry-After", "60")
				}
				w.WriteHeader(status)
				if status == http.StatusOK {
					_, _ = w.Write([]byte(`{"order":"12345","status":"PROCESSED","accrual":500.25}`))
				}
			}))
			defer ts.Close()
			client := New(Config{Address: ts.URL, Timeout: time.Second}, slog.New(slog.DiscardHandler))
			client.httpClient.SetRetryWaitTime(time.Millisecond).SetRetryMaxWaitTime(time.Millisecond)
			resp, err := client.GetOrder(context.Background(), "12345")
			require.Equal(t, tt.wantCalls, calls.Load())
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.Equal(t, StatusProcessed, resp.Status)
				require.True(t, decimal.RequireFromString("500.25").Equal(*resp.Accrual))
			}
			if tt.rateLimited {
				var rateLimitErr *RateLimitError
				require.ErrorAs(t, err, &rateLimitErr)
				require.Equal(t, time.Minute, rateLimitErr.RetryAfter)
			}
		})
	}
}

func TestParseRetryAfter(t *testing.T) {
	future := time.Now().Add(time.Minute).UTC().Truncate(time.Second)
	tests := []struct {
		name        string
		header      string
		want        time.Duration
		wantWarning bool
	}{
		{name: "seconds", header: "60", want: time.Minute},
		{name: "zero", header: "0", want: minRetryAfter},
		{name: "http date", header: future.Format(http.TimeFormat), want: time.Until(future)},
		{name: "past date", header: time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat), want: minRetryAfter},
		{name: "invalid", header: "tomorrow", want: defaultRetryAfter, wantWarning: true},
		{name: "negative", header: "-1", want: defaultRetryAfter, wantWarning: true},
		{name: "empty", want: defaultRetryAfter, wantWarning: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logs, nil))
			got := parseRetryAfter(tt.header, logger)
			require.InDelta(t, tt.want.Seconds(), got.Seconds(), 1)
			if tt.wantWarning {
				var entry map[string]any
				require.NoError(t, json.Unmarshal(logs.Bytes(), &entry))
				require.Equal(t, "WARN", entry["level"])
				require.Equal(t, tt.header, entry["retry_after"])
			} else {
				require.Empty(t, logs.String())
			}
		})
	}
}

func TestClient_GetOrder_CancelDuringRetry(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := New(Config{Address: ts.URL, Timeout: time.Second}, slog.New(slog.DiscardHandler))
	client.httpClient.AddRetryHooks(func(*resty.Response, error) { cancel() })
	resp, err := client.GetOrder(ctx, "12345")
	require.Nil(t, resp)
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, int32(1), calls.Load())
}
