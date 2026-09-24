package handler

import (
	"crypto/tls"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CimaCha/gophermart-loyal-service/internal/user/service"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_RegisterUser(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		serviceErr error
		wantStatus int
		wantCalled bool
		wantSecure bool
		requestTLS bool
	}{
		{name: "malformed json", body: `{"login":`, wantStatus: http.StatusBadRequest},
		{name: "empty login", body: `{"login":"","password":"secret"}`, wantStatus: http.StatusBadRequest},
		{name: "empty password", body: `{"login":"alice","password":""}`, wantStatus: http.StatusBadRequest},
		{name: "multiple json values", body: `{"login":"alice","password":"secret"}{}`, wantStatus: http.StatusBadRequest},
		{name: "duplicate user", body: `{"login":"alice","password":"secret"}`, serviceErr: service.ErrUserAlreadyExists, wantStatus: http.StatusConflict, wantCalled: true},
		{name: "internal error", body: `{"login":"alice","password":"secret"}`, serviceErr: errors.New("storage failed"), wantStatus: http.StatusInternalServerError, wantCalled: true},
		{name: "success over http", body: `{"login":"alice","password":"secret"}`, wantStatus: http.StatusOK, wantCalled: true},
		{name: "success over https", body: `{"login":"alice","password":"secret"}`, wantStatus: http.StatusOK, wantCalled: true, wantSecure: true, requestTLS: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceMock := NewMockUserService(t)
			if tt.wantCalled {
				serviceMock.EXPECT().
					CreateUser(testifymock.Anything, "alice", "secret").
					Return("signed-token", tt.serviceErr)
			}
			h := New(slog.New(slog.NewTextHandler(io.Discard, nil)), serviceMock)

			req := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(tt.body))
			if tt.requestTLS {
				req.URL.Scheme = "https"
				req.TLS = &tlsState
			}
			recorder := httptest.NewRecorder()

			h.RegisterUser(recorder, req)

			require.Equal(t, tt.wantStatus, recorder.Code)
			if tt.wantStatus == http.StatusOK {
				cookies := recorder.Result().Cookies()
				require.Len(t, cookies, 1)
				require.Equal(t, "jwt", cookies[0].Name)
				require.Equal(t, "signed-token", cookies[0].Value)
				require.Equal(t, "/api/user", cookies[0].Path)
				require.True(t, cookies[0].HttpOnly)
				require.Equal(t, tt.wantSecure, cookies[0].Secure)
			}
		})
	}
}

func TestHandler_LoginUser(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		serviceErr error
		wantStatus int
		wantCalled bool
	}{
		{name: "malformed json", body: `{"login":`, wantStatus: http.StatusBadRequest},
		{name: "empty login", body: `{"login":"","password":"secret"}`, wantStatus: http.StatusBadRequest},
		{name: "empty password", body: `{"login":"alice","password":""}`, wantStatus: http.StatusBadRequest},
		{name: "multiple json values", body: `{"login":"alice","password":"secret"}{}`, wantStatus: http.StatusBadRequest},
		{name: "invalid credentials", body: `{"login":"alice","password":"secret"}`, serviceErr: service.ErrInvalidCredentials, wantStatus: http.StatusUnauthorized, wantCalled: true},
		{name: "internal error", body: `{"login":"alice","password":"secret"}`, serviceErr: errors.New("storage failed"), wantStatus: http.StatusInternalServerError, wantCalled: true},
		{name: "success", body: `{"login":"alice","password":"secret"}`, wantStatus: http.StatusOK, wantCalled: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceMock := NewMockUserService(t)
			if tt.wantCalled {
				serviceMock.EXPECT().
					LoginUser(testifymock.Anything, "alice", "secret").
					Return("signed-token", tt.serviceErr)
			}
			h := New(slog.New(slog.NewTextHandler(io.Discard, nil)), serviceMock)

			recorder := httptest.NewRecorder()
			h.LoginUser(recorder, httptest.NewRequest(http.MethodPost, "/api/user/login", strings.NewReader(tt.body)))

			require.Equal(t, tt.wantStatus, recorder.Code)
		})
	}
}

var tlsState = tls.ConnectionState{}
