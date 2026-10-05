// Package model содержит структуры данных, доменные модели и DTO
// для управления профилями, учетными записями и учетными данными пользователей.
package model

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Credentials описывает входящую JSON-структуру (DTO) для передачи
// аутентификационных данных (логина и незахешированного пароля) при регистрации или входе.
type Credentials struct {
	// Login содержит уникальное строковое имя (логин) пользователя.
	Login string `json:"login"`
	// Password содержит открытый текстовый пароль, присланный клиентом.
	Password string `json:"password"`
}

var (
	// ErrEmptyLogin возвращается, если логин отсутствует или состоит только из пробелов.
	ErrEmptyLogin = errors.New("login is required")

	// ErrEmptyPassword возвращается, если пароль отсутствует или состоит только из пробелов.
	ErrEmptyPassword = errors.New("password is required")
)

// UserInfo представляет доменную модель учетной записи пользователя в системе,
// инкапсулирующую его уникальный UUID, логин, криптографический хеш пароля и дату регистрации.
type UserInfo struct {
	// UUID хранит уникальный криптографический идентификатор пользователя (RFC 4122).
	UUID uuid.UUID
	// Login содержит строковое имя (логин) учетной записи.
	Login string
	// PasswordHash содержит вычисленный криптографический хеш пароля (например, Argon2id).
	PasswordHash string
	// CreatedAt фиксирует точную дату и время регистрации аккаунта в системе.
	CreatedAt time.Time
}

// Validate проверяет корректность учетных данных пользователя.
//
// Метод убеждается, что логин и пароль заданы и не состоят
// исключительно из пробельных символов.
func (c Credentials) Validate() error {
	if strings.TrimSpace(c.Login) == "" {
		return ErrEmptyLogin
	}

	if strings.TrimSpace(c.Password) == "" {
		return ErrEmptyPassword
	}

	return nil
}
