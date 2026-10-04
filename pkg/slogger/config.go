// Package slogger предоставляет структуры и функции для конфигурации
// многопоточного логирования с поддержкой вывода в консоль (stdout) и файлы на основе YAML-файлов.
package slogger

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"go.yaml.in/yaml/v3"
)

// Format определяет строковый тип для представления формата вывода логов.
type Format string

const (
	// FormatText представляет текстовый формат логирования (через slog.NewTextHandler).
	FormatText Format = "text"
	// FormatJSON представляет формат логирования в виде JSON-структур (через slog.NewJSONHandler).
	FormatJSON Format = "json"
)

// Level определяет строковый тип для представления уровней важности логов.
type Level string

const (
	// LevelDebug обозначает отладочный уровень логирования.
	LevelDebug Level = "debug"
	// LevelInfo обозначает стандартный информационный уровень логирования.
	LevelInfo Level = "info"
	// LevelWarn обозначает уровень логирования для предупреждений.
	LevelWarn Level = "warn"
	// LevelError обозначает уровень логирования для ошибок.
	LevelError Level = "error"
)

// Config описывает корневую структуру файла конфигурации логгера.
type Config struct {
	// Directory задает путь к папке, в которой будут создаваться файлы логов.
	Directory string `yaml:"directory"`

	// Stdout содержит настройки для вывода логов в консоль.
	Stdout StdoutConfig `yaml:"stdout"`
	// Files содержит коллекцию настроек для файловых логгеров.
	Files []FileConfig `yaml:"files"`
}

// StdoutConfig описывает параметры вывода логов в стандартный поток вывода (консоль).
type StdoutConfig struct {
	// Enabled активирует или деактивирует вывод логов в консоль.
	Enabled bool `yaml:"enabled"`
	// Format задает структуру вывода (json или text).
	Format Format `yaml:"format"`
	// Level определяет минимальный уровень логирования для консоли.
	Level Level `yaml:"level"`
	// Writer абстрагирует целевой поток вывода (по умолчанию os.Stdout, полезно для тестов).
	Writer io.Writer `yaml:"-"`
}

// FileConfig описывает параметры вывода логов в отдельный файл.
type FileConfig struct {
	// Name задает уникальное имя или относительный путь для файла лога.
	Name string `yaml:"name"`
	// Enabled активирует или деактивирует ведение данного файла логов.
	Enabled bool `yaml:"enabled"`
	// Format задает структуру вывода (json или text) для файла.
	Format Format `yaml:"format"`
	// Level определяет минимальный уровень логирования для файла.
	Level Level `yaml:"level"`
}

// Validate проверяет валидность всей конфигурации логирования, включая
// параметры консоли, файлов и отсутствие дублирования имен конфигураций файлов.
func (c *Config) Validate() error {

	if c == nil {
		return errors.New("slogger config: section is required")
	}

	if c.Stdout.Enabled {
		if err := c.Stdout.Validate(); err != nil {
			return fmt.Errorf("stdout: %w", err)
		}
	}

	names := make(map[string]struct{})

	for i, file := range c.Files {
		if err := file.Validate(); err != nil {
			return fmt.Errorf("files[%d]: %w", i, err)
		}

		if _, ok := names[file.Name]; ok {
			return fmt.Errorf("duplicate file logger name %q", file.Name)
		}

		names[file.Name] = struct{}{}
	}

	return nil
}

// Validate проверяет корректность параметров конфигурации вывода в консоль.
func (c *StdoutConfig) Validate() error {

	if c == nil {
		return errors.New("slogger config: stdout config is required")
	}

	if !c.Enabled {
		return nil
	}

	if !c.Format.Valid() {
		return fmt.Errorf("invalid format %q", c.Format)
	}

	if !c.Level.Valid() {
		return fmt.Errorf("invalid level %q", c.Level)
	}

	return nil
}

// Validate проверяет корректность параметров конфигурации вывода в файл.
func (c *FileConfig) Validate() error {

	if c == nil {
		return errors.New("slogger config: file config is required")
	}

	if !c.Enabled {
		return nil
	}

	if c.Name == "" {
		return errors.New("name is required")
	}

	if !c.Format.Valid() {
		return fmt.Errorf("invalid format %q", c.Format)
	}

	if !c.Level.Valid() {
		return fmt.Errorf("invalid level %q", c.Level)
	}

	return nil
}

// Valid проверяет, соответствует ли строковое значение формата одному из допустимых типов (text, json).
func (f Format) Valid() bool {
	switch f {
	case FormatText, FormatJSON:
		return true
	default:
		return false
	}
}

// Valid проверяет, входит ли строковое значение уровня в список поддерживаемых (debug, info, warn, error).
func (l Level) Valid() bool {
	switch l {
	case LevelDebug, LevelInfo, LevelWarn, LevelError:
		return true
	default:
		return false
	}
}

// LoadFromYAML считывает файл по указанному пути, декодирует его структуру из формата YAML
// и инициализирует дефолтный Writer для потока stdout.
func LoadFromYAML(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("os read file: %w", err)
	}

	var cfg Config

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("yaml unmarshal: %w", err)
	}

	// По умолчанию используется стандартный вывод операционной системы
	cfg.Stdout.Writer = os.Stdout

	return &cfg, nil
}

// SlogLevel сопоставляет внутренний строковый тип Level с соответствующей константой уровня из пакета log/slog.
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
