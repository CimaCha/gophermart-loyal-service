package authentication

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type JWTBuilder struct {
	SecretKey []byte
}

func NewJWTBuilder(secretKey []byte) *JWTBuilder {
	return &JWTBuilder{
		SecretKey: secretKey,
	}
}

// Claims includes registered JWT claims and the authenticated user ID.
type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
}

// BuildJWTString создаёт токен и возвращает его в виде строки.
func (b JWTBuilder) BuildJWTString(userID uuid.UUID) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		// собственное утверждение
		UserID: userID,
	})

	// создаём строку токена
	return token.SignedString(b.SecretKey)
}
