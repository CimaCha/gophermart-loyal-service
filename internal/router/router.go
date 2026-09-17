package router

import (
	"github.com/CimaCha/gophermart-loyal-service/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"net/http"
)

func New(log *zap.Logger, userRegisterHandler http.Handler) http.Handler {
	router := chi.NewRouter()
	router.Use(logger.RequestLogger(log))
	router.With(middleware.AllowContentType("application/json")).
		Method(http.MethodPost, "/api/user/register", userRegisterHandler)
	return router
}
