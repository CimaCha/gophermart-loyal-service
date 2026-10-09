package httpio

import "net/http"

type (
	responseData struct {
		statusCode int
		bodySize   int
	}
	// LoggingResponseWriter представляет собой обертку над стандартным http.ResponseWriter,
	// предназначенную для перехвата и логирования HTTP-статуса ответа и размера переданного тела.
	LoggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

var (
	// StatusCodeUninitialized используется в качестве начального значения для statusCode,
	// указывая на то, что ответ от сервера еще не был сформирован.
	StatusCodeUninitialized = -1
)

// NewLoggingResponseWriter создает и инициализирует новый LoggingResponseWriter.
// Инициализирует внутреннюю структуру данных для корректного подсчета байт и статуса.
func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{
		ResponseWriter: w,
		responseData: &responseData{
			statusCode: StatusCodeUninitialized,
		},
	}
}

// Write записывает данные в HTTP-ответ и суммирует количество отправленных байт.
// Соответствует интерфейсу io.Writer.
func (r *LoggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.bodySize += size
	return size, err
}

// WriteHeader отправляет HTTP-статус код клиенту и сохраняет его для последующего логирования.
// Соответствует интерфейсу http.ResponseWriter.
func (r *LoggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.statusCode = statusCode
}

// GetStatusCode возвращает отправленный HTTP-статус код.
// Внимание: вызовет panic, если метод вызван до отправки заголовков (метод WriteHeader или Write не вызывались).
func (r *LoggingResponseWriter) GetStatusCode() int {
	if r.responseData.statusCode == StatusCodeUninitialized {
		panic("no status code set")
	}
	return r.responseData.statusCode
}

// GetBodySize возвращает общий размер (в байтах) тела ответа, отправленного клиенту.
func (r *LoggingResponseWriter) GetBodySize() int {
	return r.responseData.bodySize
}
