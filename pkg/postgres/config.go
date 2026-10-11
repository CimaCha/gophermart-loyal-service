package postgres

import (
	"errors"
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

// Config хранит конфигурационные параметры, необходимые для подключения к базе данных.
type Config struct {
	// URI содержит строку подключения (Data Source Name) к PostgreSQL,
	// включая хост, порт, пользователя, пароль и имя базы данных.
	URI string `env:"DATABASE_URI"`
}

// RegisterFlags регистрирует флаги командной строки для конфигурации PostgreSQL
// в переданном FlagSet и возвращает указатель на структуру Config со значениями по умолчанию.
func RegisterFlags(fs *flag.FlagSet) *Config {

	postgresConfig := new(Config)

	fs.StringVar(
		&postgresConfig.URI,
		"d",
		"postgres://postgres:admin@localhost:5433/db",
		"database url connection",
	)

	return postgresConfig
}

// ParseEnv считывает переменные окружения и переопределяет соответствующие поля конфигурации,
// если они были установлены (в данном случае переменную DATABASE_URI).
func (c *Config) ParseEnv() error {
	if err := env.Parse(c); err != nil {
		return fmt.Errorf("parse env: %w", err)
	}
	return nil
}

// Validate проверяет корректность заполнения конфигурации базы данных.
// Возвращает ошибку, если указатель на структуру равен nil или строка подключения URI пуста.
func (c *Config) Validate() error {

	if c == nil {
		return errors.New("postgres config: section is required")
	}

	if c.URI == "" {
		return errors.New("the database URI can't be empty")
	}
	return nil
}
