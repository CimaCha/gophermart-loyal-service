package worker

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/stretchr/testify/assert/yaml"
)

// Config описывает параметры конфигурации для управления жизненным циклом и производительностью фоновых воркеров.
type Config struct {
	// PollingInterval - как часто воркер будет ходить в БД и проверять необходимые ему записи
	PollingInterval time.Duration `yaml:"polling_interval"`
	// WorkerCount - количество параллельно работающих воркеров
	WorkerCount int `yaml:"worker_count"`
	// JobsQueueSize - максимальный размер работ, которые может принять канал
	JobsQueueSize int `yaml:"jobs_queue_size"`
}

// LoadFromYAML загружает, считывает и десериализует параметры конфигурации воркеров из YAML-файла по указанному пути.
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

// Validate проверяет корректность установленных параметров конфигурации воркеров.
// Возвращает ошибку, если интервал поллинга меньше или равен нулю, либо количество воркеров и размер очереди меньше 1.
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

	return nil
}
