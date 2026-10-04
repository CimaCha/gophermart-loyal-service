package httpio

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type gzipWriter struct {
	// Имплементим методы от оригинала
	w  http.ResponseWriter
	zw *gzip.Writer

	// Если заголовок не был проставлен - сами поставим
	hasHeader bool
	// Если Content-Type НЕ application/json или text/html то не сжимаем
	// и пишем через оригинальный responswWriter
	compress bool
}

// NewGzipWriter создает и инициализирует новый gzipWriter, обертывающий стандартный http.ResponseWriter.
// Компрессия будет применена только для типов контента application/json и text/html.
func NewGzipWriter(w http.ResponseWriter) *gzipWriter {
	zw := gzip.NewWriter(w)

	return &gzipWriter{
		w:         w,
		zw:        zw,
		hasHeader: false,
	}
}

// Header возвращает карту заголовков (http.Header) оригинального ResponseWriter.
// Соответствует интерфейсу http.ResponseWriter.
func (gw *gzipWriter) Header() http.Header {
	return gw.w.Header()
}

// Write записывает данные в HTTP-ответ. Если включена компрессия, данные автоматически сжимаются.
// Если заголовки ответа еще не были отправлены, автоматически вызывает WriteHeader со статусом 200 OK.
// Соответствует интерфейсу io.Writer.
func (gw *gzipWriter) Write(p []byte) (int, error) {

	if !gw.hasHeader {
		gw.WriteHeader(http.StatusOK)
	}

	if !gw.compress {
		return gw.w.Write(p)
	}

	return gw.zw.Write(p)
}

// WriteHeader отправляет HTTP-статус код клиенту.
// Проверяет Content-Type ответа: если это JSON или HTML, активирует Gzip-сжатие
// и добавляет заголовок Content-Encoding: gzip.
// Соответствует интерфейсу http.ResponseWriter.
func (gw *gzipWriter) WriteHeader(statusCode int) {
	// Статус коды без тела ответа игноирруем
	if statusCode != http.StatusNoContent &&
		statusCode != http.StatusNotModified {

		// Смотрим есть ли заголовок с содержимом ответа от сервера
		contentType := gw.w.Header().Get("Content-Type")
		// Компрессим только json и html, остальное игнорируем
		if strings.Contains(contentType, "application/json") ||
			strings.Contains(contentType, "text/html") {

			gw.compress = true
			// Выставляем для принимающего клиента ответ
			gw.w.Header().Set(
				"Content-Encoding",
				"gzip",
			)

		}
	}

	gw.w.WriteHeader(statusCode)
	gw.hasHeader = true
}

// Close завершает процесс сжатия и сбрасывает оставшиеся буферы в поток ответа.
// Должен быть вызван обязательно, если была включена компрессия.
func (gw *gzipWriter) Close() error {

	if !gw.compress {
		return nil
	}

	return gw.zw.Close()
}
