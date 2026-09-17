package slogger

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"go.yaml.in/yaml/v3"
)

type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

type Config struct {
	// Директория с log файлами
	Directory string `yaml:"directory"`

	// Структура с конфигурацией для вывода в stdout (консоль)
	Stdout StdoutConfig `yaml:"stdout"`
	// Слайс структур с конфигурациями для файлов логов
	Files []FileConfig `yaml:"files"`
}

type StdoutConfig struct {
	// Enabled даёт возможность выключить вывод логов через конфигурацию
	Enabled bool `yaml:"enabled"`
	// Format задаёт формат вывода в виде json или text
	Format Format `yaml:"format"`

	Level Level `yaml:"level"`
	// Можно использовать любой Writer + удобно для unit тестов
	Writer io.Writer `yaml:"-"`
}

type FileConfig struct {
	Name string `yaml:"name"`

	Enabled bool   `yaml:"enabled"`
	Format  Format `yaml:"format"`
	Level   Level  `yaml:"level"`
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

	// по умолчанию os.Stdout
	cfg.Stdout.Writer = os.Stdout

	return &cfg, nil
}

func (l Level) SlogLevel() slog.Level {
	switch l {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
