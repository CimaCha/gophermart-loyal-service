package handler

import "github.com/go-chi/chi/v5"

// type UserService interface...

type Handler struct {
	// service
	// logger
}

func New() *Handler {
	return &Handler{}
}

// Публичные т.к. не требуют Auth middleware на них
// Вход в аккаунт - процесс аунтетификации

func (h *Handler) RegisterPublicAPI(r chi.Router) {
	// r.Post("/api/user/register", h.RegisterUser)
	// r.Post("/api/user/login", h.LoginUser)
}
