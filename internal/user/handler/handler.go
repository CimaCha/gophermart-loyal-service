package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/CimaCha/gophermart-loyal-service/internal/user/model"
	"github.com/CimaCha/gophermart-loyal-service/internal/user/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
)

type UserService interface {
	CreateUser(ctx context.Context, userLogin string, password string) (string, error)
	LoginUser(ctx context.Context, userLogin string, password string) (string, error)
}

type Handler struct {
	logger  slog.Logger
	service UserService
}

func New(logger slog.Logger, userService UserService) *Handler {
	return &Handler{
		logger:  logger,
		service: userService,
	}
}

// Публичные т.к. не требуют Auth middleware на них
// Вход в аккаунт - процесс аунтетификации

func (h *Handler) RegisterPublicAPI(r chi.Router) {
	r.With(middleware.AllowContentType("application/json")).
		Post("/api/user/register", h.RegisterUser)
	r.With(middleware.AllowContentType("application/json")).
		Post("/api/user/login", h.LoginUser)
}

func (h *Handler) RegisterUser(writer http.ResponseWriter, request *http.Request) {
	userLoginInfo := model.UserRegisterIn{}
	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(&userLoginInfo)
	if err != nil {
		h.logger.Error(err.Error())
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if userLoginInfo.Login == "" || userLoginInfo.Password == "" {
		h.logger.Error(err.Error())
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	jwtToken, err := h.service.CreateUser(request.Context(), userLoginInfo.Login, userLoginInfo.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			h.logger.Error(err.Error())
			http.Error(writer, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}
		if errors.Is(err, service.ErrInvalidCredentials) {
			h.logger.Error(err.Error())
			http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		h.logger.Error(err.Error())
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	h.setAuthCookie(writer, *request, jwtToken)

	writer.WriteHeader(http.StatusOK)
}

func (h *Handler) LoginUser(writer http.ResponseWriter, request *http.Request) {
	userLoginInfo := model.UserLoginIn{}
	decoder := json.NewDecoder(request.Body)
	err := decoder.Decode(&userLoginInfo)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	jwtToken, err := h.service.LoginUser(request.Context(), userLoginInfo.Login, userLoginInfo.Password)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	h.setAuthCookie(writer, *request, jwtToken)

	writer.WriteHeader(http.StatusOK)
}

func (h *Handler) setAuthCookie(w http.ResponseWriter, r http.Request, token string) {

	cookie := &http.Cookie{
		Name:     "jwt",
		Path:     "/api/user",
		Value:    token,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
	}

	http.SetCookie(w, cookie)
}
