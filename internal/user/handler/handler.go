package handler

import (
	"context"
	"encoding/json"
	authentication "github.com/CimaCha/gophermart-loyal-service/internal/auth"
	"github.com/CimaCha/gophermart-loyal-service/internal/user/model"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
)

type contextKey struct{}

type UserService interface {
	CreateUser(userLogin string, password string) error
	LoginUser(userLogin string, password string) error
}

type Handler struct {
	logger     slog.Logger
	service    UserService
	jwtBuilder authentication.JWTBuilder
}

func New(logger slog.Logger, userService UserService, jwtBuilder authentication.JWTBuilder) *Handler {
	return &Handler{
		logger:     logger,
		service:    userService,
		jwtBuilder: jwtBuilder,
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
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	err = h.service.CreateUser(userLoginInfo.Login, userLoginInfo.Password)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer, request, err = h.setUserLoginCookie(userLoginInfo.Login, &writer, request)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
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
	err = h.service.LoginUser(userLoginInfo.Login, userLoginInfo.Password)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer, request, err = h.setUserLoginCookie(userLoginInfo.Login, &writer, request)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func getUserLogin(ctx context.Context) string {
	id, _ := ctx.Value(contextKey{}).(string)
	return id
}

// UserLogin returns the authenticated user login, or an empty string if absent.
func UserLogin(ctx context.Context) string {
	return getUserLogin(ctx)
}

func (h *Handler) setUserLoginCookie(userLogin string, writer *http.ResponseWriter, request *http.Request) (http.ResponseWriter, *http.Request, error) {
	jwtString, err := h.jwtBuilder.BuildJWTString(userLogin)
	if err != nil {
		return nil, nil, err
	}

	request = request.WithContext(context.WithValue(request.Context(), contextKey{}, userLogin))

	cookie := &http.Cookie{
		Name:     "jwt",
		Value:    jwtString,
		HttpOnly: true,
		Secure:   request.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(*writer, cookie)
	return *writer, request, nil
}
