package ratelimit

import (
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
