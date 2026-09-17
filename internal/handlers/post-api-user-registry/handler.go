package post_api_user_registry

import (
	"context"
	"go.uber.org/zap"
	"net/http"
)

type Register interface {
	RegisterUser(ctx context.Context, userID string, password string) error
}

type Handler struct {
	log      zap.Logger
	register Register
}

func (h Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me
	panic("implement me")
}

func NewRegisterHandler(logger zap.Logger, register Register) Handler {
	return Handler{log: logger, register: register}
}
