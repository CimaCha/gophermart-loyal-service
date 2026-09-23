package authentication

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestParser_ValidateUserID(t *testing.T) {
	secret := []byte("test-secret")
	userID := uuid.New()
	zeroPrefixID := uuid.MustParse("00000000-0000-4000-8000-000000000001")

	tests := []struct {
		name    string
		token   func() string
		secret  []byte
		wantID  uuid.UUID
		wantErr error
	}{
		{
			name: "valid token",
			token: func() string {
				return signClaims(t, jwt.SigningMethodHS256, secret, Claims{UserID: userID})
			},
			secret: secret,
			wantID: userID,
		},
		{
			name: "wrong secret",
			token: func() string {
				return signClaims(t, jwt.SigningMethodHS256, secret, Claims{UserID: userID})
			},
			secret:  []byte("wrong-secret"),
			wantErr: ErrInvalidToken,
		},
		{
			name: "tampered signature",
			token: func() string {
				return signClaims(t, jwt.SigningMethodHS256, secret, Claims{UserID: userID}) + "x"
			},
			secret:  secret,
			wantErr: ErrInvalidToken,
		},
		{
			name: "unsupported algorithm",
			token: func() string {
				return signClaims(t, jwt.SigningMethodHS384, secret, Claims{UserID: userID})
			},
			secret:  secret,
			wantErr: ErrInvalidToken,
		},
		{
			name: "missing user id",
			token: func() string {
				return signClaims(t, jwt.SigningMethodHS256, secret, Claims{})
			},
			secret:  secret,
			wantErr: ErrMissingUserID,
		},
		{
			name: "expired token",
			token: func() string {
				return signClaims(t, jwt.SigningMethodHS256, secret, Claims{
					RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour))},
					UserID:           userID,
				})
			},
			secret:  secret,
			wantErr: ErrExpiredToken,
		},
		{
			name: "uuid with zero prefix",
			token: func() string {
				return signClaims(t, jwt.SigningMethodHS256, secret, Claims{UserID: zeroPrefixID})
			},
			secret: secret,
			wantID: zeroPrefixID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewUserIDParser(tt.secret).ValidateUserID(tt.token())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Equal(t, uuid.Nil, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantID, got)
		})
	}
}

func signClaims(t *testing.T, method jwt.SigningMethod, secret []byte, claims Claims) string {
	t.Helper()
	token, err := jwt.NewWithClaims(method, claims).SignedString(secret)
	require.NoError(t, err)
	return token
}
