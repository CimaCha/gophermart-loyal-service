package worker

import (
	"errors"
	"fmt"
	"os"
	"time"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	// PollingInterval - как часто воркер будет ходить в БД и проверять необходимые ему записи
	PollingInterval time.Duration `yaml:"polling_interval"`
	// WorkerCount - количество параллельно работающих воркеров
	WorkerCount int `yaml:"worker_count"`
	// JobsQueueSize - максимальный размер работ, которые может принять канал
	JobsQueueSize int `yaml:"jobs_queue_size"`
	// TargetRPS - условно сколько запросов в секунду можем делать к accrual
	TargetRPS float64 `yaml:"target_rps"`
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

func (c *Config) Validate() error {

	if c == nil {
		return errors.New("worker config: section is required")
	}

	if c.PollingInterval <= 0 {
		return errors.New("polling interval must be greater than zero")
	}
	if c.WorkerCount < 1 {
		return errors.New("worker count must be at least 1")
	}
	if c.JobsQueueSize < 1 {
		return errors.New("jobs queue size must be at least 1")
	}
	if c.TargetRPS <= 0 {
		return errors.New("target RPS must be greater than zero")
	}
	return nil
}
