package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

func TestLogging_CallsNextHandler(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	logger := newTestLogger(buf)

	called := false

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	Logging(logger)(handler).ServeHTTP(rec, req)

	require.True(t, called)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestLogging_ResponseIsNotModified(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	logger := newTestLogger(buf)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, err := w.Write([]byte("hello"))
		require.NoError(t, err)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/user/login", nil)
	rec := httptest.NewRecorder()

	Logging(logger)(handler).ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Equal(t, "hello", rec.Body.String())
}

func TestLogging_WritesLogs(t *testing.T) {
	t.Parallel()

	buf := new(bytes.Buffer)
	logger := newTestLogger(buf)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("response"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(requestHeader, "req-123")
	req.Header.Set("User-Agent", "Go-Test")

	rec := httptest.NewRecorder()

	Logging(logger)(handler).ServeHTTP(rec, req)

	logs := buf.String()

	require.Contains(t, logs, "HTTP Request started")
	require.Contains(t, logs, "HTTP Request done")

	require.Contains(t, logs, "request_id=req-123")
	require.Contains(t, logs, "user_agent=Go-Test")
	require.Contains(t, logs, "URI=/test")
	require.Contains(t, logs, "method=GET")

	require.Contains(t, logs, "status=200")
	require.Contains(t, logs, "size=8")
	require.Contains(t, logs, "latency=")
}
