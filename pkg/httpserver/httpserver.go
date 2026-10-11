// Package httpserver предоставляет абстракцию над стандартным http.Server
// для удобного запуска и плавного (graceful) завершения работы сервера.
package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// HTTPServer оборачивает стандартный HTTP-сервер и логгер,
// обеспечивая управление жизненным циклом приложения.
type HTTPServer struct {
	sv  *http.Server
	log *slog.Logger
}

// New создает и инициализирует новый экземпляр HTTPServer.
// Принимает обработчик запросов (handler), конфигурацию сервера (cfg) и настроенный slog.Logger.
func New(handler http.Handler, cfg *Config, log *slog.Logger) *HTTPServer {
	return &HTTPServer{
		sv: &http.Server{
			Addr:    cfg.Addr,
			Handler: handler,
		},
		log: log,
	}
}

// Run запускает HTTP-сервер в отдельной горутине и блокирует текущий поток выполнения
// до возникновения критической ошибки или до сигнала отмены контекста ctx (Graceful Shutdown).
func (s *HTTPServer) Run(ctx context.Context) error {

	chErr := make(chan error, 1)

	go func() {

		s.log.Info(
			"starting http server",
			slog.String("Addr", s.sv.Addr),
		)

		if err := s.sv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			chErr <- err
		}

		close(chErr)
	}()

	select {
	case err := <-chErr:
		s.log.Error(
			"failed to start server",
			"err",
			err,
			slog.String("Addr", s.sv.Addr),
		)
		return fmt.Errorf("listen and serve: %w", err)
	case <-ctx.Done():
		s.log.Info(
			"server starting shutdown",
		)

		if err := s.shutdown(); err != nil {
			s.log.Error(
				"failed to shutdown",
				"err",
				err,
				slog.String("Addr", s.sv.Addr),
			)
			_ = s.sv.Close()
			return fmt.Errorf("failed to shutdown: %w", err)
		}
	}

	s.log.Info(
		"server successfully closed",
		slog.String("Addr", s.sv.Addr),
	)

	return nil
}

func (s *HTTPServer) shutdown() error {

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	if err := s.sv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	return nil

}
