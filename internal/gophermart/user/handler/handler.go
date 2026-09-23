package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/user/model"
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/user/service"
)

var errInvalidCredentialsRequest = errors.New("invalid credentials request")

//go:generate go tool mockgen -source=handler.go -destination=mock/user_service_gen.go -package=mock

type UserService interface {
	CreateUser(ctx context.Context, userLogin string, password string) (string, error)
	LoginUser(ctx context.Context, userLogin string, password string) (string, error)
}

type Handler struct {
	logger  *slog.Logger
	service UserService
}

func New(logger *slog.Logger, userService UserService) *Handler {
	return &Handler{
		logger:  logger,
		service: userService,
	}
}

func (h *Handler) RegisterUser(writer http.ResponseWriter, request *http.Request) {
	credentials, err := decodeCredentials(request.Body)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	jwtToken, err := h.service.CreateUser(request.Context(), credentials.Login, credentials.Password)
	if err != nil {
		h.writeServiceError(writer, err)
		return
	}
	h.setAuthCookie(writer, request, jwtToken)

	writer.WriteHeader(http.StatusOK)
}

func (h *Handler) LoginUser(writer http.ResponseWriter, request *http.Request) {
	credentials, err := decodeCredentials(request.Body)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	jwtToken, err := h.service.LoginUser(request.Context(), credentials.Login, credentials.Password)
	if err != nil {
		h.writeServiceError(writer, err)
		return
	}
	h.setAuthCookie(writer, request, jwtToken)

	writer.WriteHeader(http.StatusOK)
}

func decodeCredentials(r io.Reader) (model.Credentials, error) {
	var credentials model.Credentials
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&credentials); err != nil {
		return model.Credentials{}, err
	}
	if credentials.Login == "" || credentials.Password == "" {
		return model.Credentials{}, errInvalidCredentialsRequest
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return model.Credentials{}, errInvalidCredentialsRequest
	}
	return credentials, nil
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, service.ErrUserAlreadyExists):
		status = http.StatusConflict
	case errors.Is(err, service.ErrInvalidCredentials):
		status = http.StatusUnauthorized
	default:
		h.logger.Error("user request failed", "err", err)
	}
	http.Error(w, http.StatusText(status), status)
}

func (h *Handler) setAuthCookie(w http.ResponseWriter, r *http.Request, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Path:     "/api/user",
		Value:    token,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
	})
}
