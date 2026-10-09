// Package authentication предоставляет сервисы для генерации, парсинга
// и валидации токенов аутентификации (JWT) пользователей.
package authentication

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// TokenService инкапсулирует секретный ключ подписи и предоставляет
// методы для работы с токенами аутентификации.
type TokenService struct {
	secretKey []byte
}

var (
	// ErrExpiredToken возвращается, если срок действия токена истек.
	ErrExpiredToken = errors.New("token is expired")
	// ErrInvalidToken возвращается, если токен поврежден, имеет неверную подпись или структуру.
	ErrInvalidToken = errors.New("token is invalid")
	// ErrMissingUserID возвращается, если в декодированных утверждениях (claims) токена отсутствует UUID пользователя.
	ErrMissingUserID = errors.New("user id is missing from token claims")
)

// New создает и инициализирует новый экземпляр TokenService с переданным секретным ключом.
func New(secretKey []byte) *TokenService {
	return &TokenService{
		secretKey: secretKey,
	}
}

// Claims объединяет зарегистрированные стандартные утверждения JWT (RegisteredClaims)
// и кастомное поле идентификатора аутентифицированного пользователя (UserID).
type Claims struct {
	jwt.RegisteredClaims
	// UserID содержит уникальный UUID пользователя, которому принадлежит токен.
	UserID uuid.UUID `json:"user_id"`
}

// BuildJWTString генерирует новый криптографический JWT-токен на основе алгоритма подписи HS256
// для указанного userID со сроком действия 365 дней и возвращает его в виде строки.
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

// ValidateToken выполняет парсинг строки токена, проверяет валидность криптографической подписи HS256
// и срок его действия. В случае успеха извлекает и возвращает UUID пользователя (UserID).
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
