package httpserver

import (
	"errors"
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Config хранит конфигурационные параметры для запуска HTTP-сервера.
type Config struct {
	// Addr определяет сетевой адрес и порт, на котором сервер будет принимать запросы (например, "localhost:8080").
	Addr string `env:"RUN_ADDRESS"`
}

// RegisterFlags регистрирует флаги командной строки для конфигурации сервера
// в переданном FlagSet и возвращает указатель на структуру Config.
func RegisterFlags(fs *flag.FlagSet) *Config {

	serverCfg := new(Config)

	fs.StringVar(
		&serverCfg.Addr,
		"a",
		"localhost:8080",
		"Server address",
	)

	return serverCfg
}

// ParseEnv считывает переменные окружения и переопределяет соответствующие поля конфигурации,
// если они были установлены (в данном случае переменную RUN_ADDRESS).
func (c *Config) ParseEnv() error {
	if err := env.Parse(c); err != nil {
		return fmt.Errorf("failed to parse env: %w", err)
	}
	return nil
}

// Validate проверяет корректность заполнения полей конфигурации.
// Возвращает ошибку, если адрес запуска сервера остался пустым.
func (c Config) Validate() error {
	if c.Addr == "" {
		return errors.New("the run address can't be empty")
	}
	return nil
}
