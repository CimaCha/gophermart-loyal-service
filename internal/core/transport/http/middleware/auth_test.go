package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CimaCha/gophermart-loyal-service/internal/core/transport/http/middleware/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAuthMiddleware(t *testing.T) {
	userID := uuid.New()
	errInvalid := errors.New("invalid token")
	tests := []struct {
		name           string
		cookie         *http.Cookie
		validatorErr   error
		wantStatus     int
		wantValidated  bool
		wantDownstream bool
	}{
		{name: "missing cookie", wantStatus: http.StatusUnauthorized},
		{name: "invalid token", cookie: &http.Cookie{Name: "jwt", Value: "invalid"}, validatorErr: errInvalid, wantStatus: http.StatusUnauthorized, wantValidated: true},
		{name: "valid token", cookie: &http.Cookie{Name: "jwt", Value: "signed-token"}, wantStatus: http.StatusNoContent, wantValidated: true, wantDownstream: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			downstream := false
			validator := mock.NewMockTokenValidator(gomock.NewController(t))
			if tt.wantValidated {
				validator.EXPECT().ValidateUserID(tt.cookie.Value).Return(userID, tt.validatorErr)
			} else {
				validator.EXPECT().ValidateUserID(gomock.Any()).Times(0)
			}
			middleware := AuthMiddleware(validator)
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				downstream = true
				gotID, ok := UserID(r.Context())
				require.True(t, ok)
				require.Equal(t, userID, gotID)
				w.WriteHeader(http.StatusNoContent)
			}))

			req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			require.Equal(t, tt.wantStatus, recorder.Code)
			require.Equal(t, tt.wantDownstream, downstream)
		})
	}
}

func TestUserID(t *testing.T) {
	userID := uuid.New()
	tests := []struct {
		name   string
		ctx    context.Context
		wantID uuid.UUID
		wantOK bool
	}{
		{name: "missing", ctx: context.Background(), wantID: uuid.Nil},
		{name: "present", ctx: context.WithValue(context.Background(), contextKey{}, userID), wantID: userID, wantOK: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, ok := UserID(tt.ctx)
			require.Equal(t, tt.wantID, gotID)
			require.Equal(t, tt.wantOK, ok)
		})
	}
}
