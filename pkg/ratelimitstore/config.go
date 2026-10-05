package ratelimitstore

import (
	"errors"
	"fmt"
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

// RateLimitConfig определяет параметры ограничения частоты запросов для конкретного маршрута.
type RateLimitConfig struct {
	// Window задает продолжительность временного окна (например, 1m, 1h).
	Window time.Duration `yaml:"window"`
	// MaxRequests определяет максимальное количество разрешенных запросов внутри одного окна.
	MaxRequests int `yaml:"max_requests"`
}

// Config описывает структуру конфигурации всей подсистемы Rate Limiting.
type Config struct {
	// CleanupInterval определяет периодичность очистки InMemory-хранилища от устаревших записей.
	CleanupInterval time.Duration `yaml:"cleanup_interval"`
	// Routes содержит карту соответствия путей (маршрутов) и их индивидуальных правил ограничения частоты запросов.
	Routes map[string]RateLimitConfig `yaml:"routes"`
}

// LoadFromYAML загружает, считывает и парсит конфигурацию Rate Limiting из указанного файла в формате YAML.
func LoadFromYAML(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("os read file: %w", err)
	}

	var cfg Config

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("yaml unmarshal: %w", err)
	}

	return &cfg, nil
}

// Validate проверяет корректность параметров отдельного правила ограничения запросов.
// Возвращает ошибку, если размер окна Window меньше или равен нулю, либо MaxRequests меньше 1.
func (c *RateLimitConfig) Validate() error {
	if c == nil {
		return errors.New("ratelimit config: section is required")
	}
	if c.Window <= 0 {
		return errors.New("window size must be greater than zero")
	}
	if c.MaxRequests < 1 {
		return errors.New("max requests must be at least 1")
	}
	return nil
}

// Validate осуществляет сквозную проверку всей конфигурации Rate Limiting.
// Возвращает ошибку, если интервал очистки невалиден или одно из зарегистрированных правил маршрутов содержит некорректные данные.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("ratelimit config: section is required")
	}
	if c.CleanupInterval <= 0 {
		return errors.New("cleanup interval must be greater than zero")
	}
	for name, rl := range c.Routes {
		if err := rl.Validate(); err != nil {
			return fmt.Errorf("route %q: %w", name, err)
		}
	}
	return nil
}
