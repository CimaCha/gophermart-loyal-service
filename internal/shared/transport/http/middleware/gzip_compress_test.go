package middleware

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Тест проверяет автоматическую распаковку тела входящего запроса от клиента
// если он прислал его в формате gzip
func TestGzipCompress_RequestDecompression(t *testing.T) {
	handler := GzipCompress()(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			body, err := io.ReadAll(r.Body)

			require.NoError(t, err)

			assert.JSONEq(
				t,
				`{"message":"Hello, World!"}`,
				string(body),
			)
		},
	))

	var buf bytes.Buffer

	zw := gzip.NewWriter(&buf)

	_, err := zw.Write([]byte(`{"message":"Hello, World!"}`))
	require.NoError(t, err)

	zw.Close()

	req := httptest.NewRequest(http.MethodGet, "/", &buf)

	req.Header.Set("Content-Encoding", "gzip")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
}

// Тест проверяет, что gzip middleware не будет сжимать ответ сервера, если клиент его не ожидает
func TestGzipCompress_NoAcceptEncoding(t *testing.T) {

	handler := GzipCompress()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(
			"Content-Type",
			"application/json",
		)

		json.NewEncoder(w).Encode(map[string]string{"message": "Hello, World!"})
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)

	req.Header.Set("Accept-Encoding", "")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	resp := rec.Result()

	assert.Empty(t, resp.Header.Get("Content-Encoding"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.JSONEq(t, `{"message":"Hello, World!"}`, string(body))
}

// Тест проверяет, что gzip middleware не сжимает ответы с типом контента, для которых сжатие отключено
func TestGzipCompress_UnsupportedContentType(t *testing.T) {
	handler := GzipCompress()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		w.Write([]byte("Hello, World!"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	resp := rec.Result()

	assert.Empty(t, resp.Header.Get("Content-Encoding"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, "Hello, World!", string(body))
}

// Тест проверяет, что gzip middleware сжимает ответ от сервера с HTML контентом
func TestGzipCompress_ResponseCompressedHTML(t *testing.T) {
	handler := GzipCompress()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(
			"Content-Type",
			"text/html",
		)

		w.Write([]byte(
			`<!DOCTYPE html>
				<title>Hello, World!</title>
				<p>Hello!`))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	resp := rec.Result()

	assert.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))

	zr, err := gzip.NewReader(resp.Body)
	require.NoError(t, err)

	body, err := io.ReadAll(zr)
	require.NoError(t, err)

	assert.Equal(
		t,
		`<!DOCTYPE html>
				<title>Hello, World!</title>
				<p>Hello!`,
		string(body))
}

// Тест проверяет, что gzip middleware корректно обрабатывает ошибку, если клиент прислал некорректные сжатые данные
func TestGzipCompress_InvalidGzip(t *testing.T) {

	handler := GzipCompress()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Is nothing here"))
	}))

	var buf bytes.Buffer

	_, err := buf.Write([]byte("Hello, World!"))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/", &buf)

	req.Header.Set("Content-Encoding", "gzip")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.NotEqual(t, http.StatusCreated, rec.Code)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	responseBody := rec.Body.String()
	assert.Contains(t, responseBody, "invalid gzip body")
}

// Тест проверяет, что gzip middleware сжимает ответ от сервера с JSON контентом
func TestGzipComrpess_ResponseCompressedJSON(t *testing.T) {

	handler := GzipCompress()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(
			"Content-Type",
			"application/json",
		)

		json.NewEncoder(w).Encode(map[string]string{"message": "Hello, World!"})
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	resp := rec.Result()

	assert.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))

	// Распаковывем и проверяем
	zr, err := gzip.NewReader(resp.Body)
	require.NoError(t, err)

	body, err := io.ReadAll(zr)
	require.NoError(t, err)

	assert.JSONEq(t, `{"message":"Hello, World!"}`, string(body))
}

// Тест на сжатие JSON в ответе от сервера
func TestGzipMiddleware_Response(t *testing.T) {
	router := chi.NewRouter()

	router.Use(GzipCompress())

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set(
			"Content-Type",
			"application/json",
		)

		json.NewEncoder(w).Encode(
			map[string]string{
				"message": "hello",
			},
		)
	})

	server := httptest.NewServer(router)
	defer server.Close()

	req, err := http.NewRequest(
		http.MethodGet,
		server.URL+"/",
		nil,
	)
	require.NoError(t, err)

	req.Header.Set(
		"Accept-Encoding",
		"gzip",
	)

	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(
		t,
		"gzip",
		resp.Header.Get("Content-Encoding"),
	)

	zr, err := gzip.NewReader(resp.Body)
	require.NoError(t, err)

	defer zr.Close()

	body, err := io.ReadAll(zr)
	require.NoError(t, err)

	assert.JSONEq(
		t,
		`{"message":"hello"}`,
		string(body),
	)
}

// Тест проверяет распоковку входящего запроса с JSON телом
func TestGzipMiddleware_Request(t *testing.T) {
	router := chi.NewRouter()

	router.Use(GzipCompress())

	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		assert.Equal(
			t,
			`{"msg":"JSON"}`,
			string(body),
		)

		w.WriteHeader(http.StatusCreated)
	})

	server := httptest.NewServer(router)
	defer server.Close()

	req, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/",
		gzipData(
			t,
			[]byte(`{"msg":"JSON"}`),
		),
	)
	require.NoError(t, err)

	req.Header.Set(
		"Content-Encoding",
		"gzip",
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(
		t,
		http.StatusCreated,
		resp.StatusCode,
	)
}

func gzipData(t *testing.T, data []byte) io.Reader {
	t.Helper()

	var buf bytes.Buffer

	zw := gzip.NewWriter(&buf)

	_, err := zw.Write(data)
	require.NoError(t, err)

	err = zw.Close()
	require.NoError(t, err)

	return &buf
}
