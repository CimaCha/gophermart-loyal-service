package main

import (
	"context"
	"github.com/CimaCha/gophermart-loyal-service/internal/config"
	postapiuserregistry "github.com/CimaCha/gophermart-loyal-service/internal/handlers/post-api-user-registry"
	"github.com/CimaCha/gophermart-loyal-service/internal/logger"
	"github.com/CimaCha/gophermart-loyal-service/internal/repository"
	"github.com/CimaCha/gophermart-loyal-service/internal/router"
	userservice "github.com/CimaCha/gophermart-loyal-service/internal/service/user-service"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func main() {
	initializedLogger, err := logger.Initialize("debug")
	if err != nil {
		log.Fatal("logger initialization error", err.Error())
	}
	defer initializedLogger.Sync()

	if err = run(*initializedLogger); err != nil {
		initializedLogger.Fatal("application stopped", zap.Error(err))
	}
}

func run(log zap.Logger) error {
	ctx := context.Background()
	cfg, err := config.New()
	if err != nil {
		log.Error("cannot parse config")
		return err
	}

	storage, err := repository.NewDatabaseStorage(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer storage.Close()

	userService := userservice.NewService(storage)

	userRegisterHandler := postapiuserregistry.NewRegisterHandler(*log.With(zap.String("handler", "shorten URL")), userService)

	apiRouter := router.New(log.With(zap.String("layer", "router")), userRegisterHandler)

	err = http.ListenAndServe(cfg.Address, apiRouter)
	if err != nil {
		log.Error("HTTP server stopped", zap.Error(err))
		return err
	}
	return nil
}
