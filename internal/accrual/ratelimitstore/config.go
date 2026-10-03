package ratelimitstore

import (
	"errors"
	"fmt"
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

type RateLimitConfig struct {
	Window      time.Duration `yaml:"window"`
	MaxRequests int           `yaml:"max_requests"`
}

type Config struct {
	CleanupInterval time.Duration              `yaml:"cleanup_interval"`
	Routes          map[string]RateLimitConfig `yaml:"routes"`
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
	if c.Window <= 0 {
		return errors.New("window size must be greater than zero")
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
		return errors.New("cleanup interval must be greater than zero")
	}
	for name, rl := range c.Routes {
		if err := rl.Validate(); err != nil {
			return fmt.Errorf("route %q: %w", name, err)
		}
	}
	return nil
}
