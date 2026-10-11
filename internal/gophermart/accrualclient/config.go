package accrualclient

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config хранит параметры подключения и ограничения по времени для внешнего HTTP-клиента системы начисления.
type Config struct {
	// Address определяет базовый сетевой URL-адрес внешней системы начисления баллов (например, "http://localhost:8081").
	Address string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	// Timeout задает максимальное время ожидания ответа на HTTP-запросы к системе начисления.
	Timeout time.Duration `env:"ACCRUAL_SYSTEM_TIMEOUT"`
}

const defaultTimeout = 10 * time.Second

// RegisterFlags регистрирует флаги командной строки для настройки адреса и таймаута клиента
// в переданном FlagSet и возвращает указатель на структуру Config со значениями по умолчанию.
func RegisterFlags(fs *flag.FlagSet) *Config {
	cfg := &Config{Timeout: defaultTimeout}

	fs.StringVar(&cfg.Address, "r", "", "accrual system address")
	fs.DurationVar(&cfg.Timeout, "accrual-timeout", defaultTimeout, "accrual request timeout")

	return cfg
}

// ParseEnv считывает переменные окружения операционной системы и переопределяет соответствующие поля конфигурации,
// если они были установлены (переменные ACCRUAL_SYSTEM_ADDRESS и ACCRUAL_SYSTEM_TIMEOUT).
func (c *Config) ParseEnv() error {
	if err := env.Parse(c); err != nil {
		return fmt.Errorf("parse env: %w", err)
	}
	return nil
}

// Validate выполняет проверку корректности и достаточности заполненных параметров конфигурации.
// Возвращает ошибку, если адрес системы начисления остался пустым или таймаут равен либо меньше нуля.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("accrual client config: section is required")
	}
	if c.Address == "" {
		return errors.New("accrual system address can't be empty")
	}
	if c.Timeout <= 0 {
		return errors.New("accrual system timeout must be greater than zero")
	}
	return nil
}
