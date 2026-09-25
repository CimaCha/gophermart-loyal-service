package middleware

import (
	"net/http"
	"strings"

	httprequest "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/transport/http/request"
	httpresponse "github.com/CimaCha/gophermart-loyal-service/internal/gophermart/transport/http/response"
)

func GzipCompress() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			originalWriter := w

			acceptsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")

			if acceptsGzip {
				gzipWriter := httpresponse.NewGzipWriter(w)

				originalWriter = gzipWriter

				defer gzipWriter.Close()
			}

			sendsGzip := strings.Contains(r.Header.Get("Content-Encoding"), "gzip")

			if sendsGzip {
				gzipReader, err := httprequest.NewGzipReader(r.Body)
				if err != nil {
					http.Error(
						w,
						"invalid gzip body",
						http.StatusBadRequest,
					)
					return
				}

				r.Body = gzipReader
				defer gzipReader.Close()
			}

			next.ServeHTTP(originalWriter, r)
		})
	}
}
