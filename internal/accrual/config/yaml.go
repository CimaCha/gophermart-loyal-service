package config

import (
	"fmt"
	"os"

	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/ratelimitstore"
	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/worker"
	"github.com/CimaCha/gophermart-loyal-service/pkg/slogger"
	"github.com/stretchr/testify/assert/yaml"
)

type yamlConfig struct {
	Logger *slogger.Config        `yaml:"logger"`
	RLS    *ratelimitstore.Config `yaml:"rate_limits"`
	Worker *worker.Config         `yaml:"worker"`
}

func loadYAML(path string) (*yamlConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg yamlConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config file: %w", err)
	}
	return &cfg, nil
}
