package main

import (
	"context"
	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/config"
	app2 "github.com/CimaCha/gophermart-loyal-service/internal/accrual/core/app"
	"github.com/CimaCha/gophermart-loyal-service/internal/shared/slogger"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sigCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// config
	cfg, err := config.Load(os.Args[1:])
	if err != nil {
		log.Printf("config load: %v\n", err)
		return
	}

	// initialize logger
	slog, closer, err := slogger.New(cfg.Logger)
	if err != nil {
		log.Printf("slogger: %v\n", err)
		return
	}

	defer func() {
		if err = closer.Close(); err != nil {
			log.Printf("logger close: %v\n", err)
		}
	}()

	// initialize app
	app, err := app2.New(sigCtx, cfg, slog)
	if err != nil {
		slog.Error(
			"failed to initialize application",
			"err", err,
		)
		return
	}

	// Run app
	if err = app.Run(sigCtx); err != nil {
		slog.Error(
			"failed to run server",
			"err", err,
		)
		return
	}
}
