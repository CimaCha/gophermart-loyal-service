package authentication

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type TokenService struct {
	secretKey []byte
}

var (
	ErrExpiredToken  = errors.New("token is expired")
	ErrInvalidToken  = errors.New("token is invalid")
	ErrMissingUserID = errors.New("user id is missing from token claims")
)

func New(secretKey []byte) *TokenService {
	return &TokenService{
		secretKey: secretKey,
	}
}

// Claims includes registered JWT claims and the authenticated user ID.
type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
}

// BuildJWTString создаёт токен и возвращает его в виде строки.
func (b TokenService) BuildJWTString(userID uuid.UUID) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims

	claims := Claims{
		UserID: userID,

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(365 * 24 * time.Hour)),

			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// создаём строку токена
	return token.SignedString(b.secretKey)
}

func (b TokenService) ValidateToken(tokenString string) (uuid.UUID, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, ErrInvalidToken
			}
			return b.secretKey, nil
		})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return uuid.UUID{}, ErrExpiredToken
		}
		return uuid.UUID{}, ErrInvalidToken
	}

	if token == nil || !token.Valid {
		return uuid.UUID{}, ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return uuid.UUID{}, ErrInvalidToken
	}

	if claims.UserID == uuid.Nil {
		return uuid.UUID{}, ErrMissingUserID
	}
	return claims.UserID, nil
}
