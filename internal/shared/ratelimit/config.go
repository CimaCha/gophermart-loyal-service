package ratelimit

import (
	"errors"
	"fmt"
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

type RateLimitConfig struct {
	KeyPrefix   string        `yaml:"key_prefix"`
	Window      time.Duration `yaml:"window"`
	MaxRequests int           `yaml:"max_requests"`
}

type Config struct {
	CleanupInterval time.Duration `yaml:"cleanup_interval"`

	Register RateLimitConfig `yaml:"register"`
	Login    RateLimitConfig `yaml:"login"`
}

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

func (c *RateLimitConfig) Validate() error {

	if c == nil {
		return errors.New("ratelimit config: section is required")
	}

	if c.KeyPrefix == "" {
		return errors.New("the key prefix can't be empty")
	}

	if c.Window < 1 {
		return errors.New("window size must be at least 1")
	}

	if c.MaxRequests < 1 {
		return errors.New("max requests must be at least 1")
	}

	return nil
}

func (c *Config) Validate() error {

	if c == nil {
		return errors.New("ratelimit config: section is required")
	}

	if c.CleanupInterval <= 0 {
		return errors.New("polling interval must be greater than zero")
	}

	if err := c.Register.Validate(); err != nil {
		return fmt.Errorf("register rate limit validate: %w", err)
	}

	if err := c.Login.Validate(); err != nil {
		return fmt.Errorf("login rate limit validate: %w", err)
	}
	return nil
}
