// Package config отвечает за сборку, парсинг и валидацию глобальной конфигурации
// приложения из различных источников: флагов командной строки, переменных окружения и YAML-файла.
package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"

	"github.com/CimaCha/gophermart-loyal-service/internal/accrual/worker"
	"github.com/CimaCha/gophermart-loyal-service/pkg/httpserver"
	"github.com/CimaCha/gophermart-loyal-service/pkg/postgres"
	"github.com/CimaCha/gophermart-loyal-service/pkg/ratelimitstore"
	"github.com/CimaCha/gophermart-loyal-service/pkg/slogger"
)

// Config объединяет в себе все конфигурационные подсистемы приложения:
// HTTP-сервер, базу данных, систему логирования, лимитер запросов и воркеры.
type Config struct {
	// Server содержит параметры запуска HTTP-сервера.
	Server *httpserver.Config
	// DB хранит строку подключения и параметры для PostgreSQL.
	DB *postgres.Config
	// Logger управляет конфигурацией вывода логов (консоль, файлы, уровни).
	Logger *slogger.Config
	// RLS содержит настройки подсистемы Rate Limit Store.
	RLS *ratelimitstore.Config
	// W содержит параметры конфигурации фоновых воркеров обработки.
	W *worker.Config
}

type validator interface {
	Validate() error
}

const (
	defaultConfigPath = "config/accrual.yaml"
	configPathEnvVar  = "ACCRUAL_CONFIG_PATH"
)

// Load инициализирует, считывает и валидирует полную конфигурацию приложения.
// Источники обрабатываются в следующем приоритете (от высшего к низшему):
// 1. Явно переданные флаги командной строки (CLI Flags)
// 2. Переменные окружения (Environment Variables)
// 3. Значения из конфигурационного YAML-файла
// 4. Дефолтные значения подсистем.
func Load(args []string) (*Config, error) {

	fs := flag.NewFlagSet("config", flag.ContinueOnError)

	configPath := defaultConfigPath
	if v := os.Getenv(configPathEnvVar); v != "" {
		configPath = v
	}
	configPathFlag := fs.String("config", configPath, "path to config file")
	serverConfig := httpserver.RegisterFlags(fs)
	dbConfig := postgres.RegisterFlags(fs)

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("parse flags: %w", err)
	}

	environment := env.ToMap(os.Environ())
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			delete(environment, "RUN_ADDRESS")
		case "d":
			delete(environment, "DATABASE_URI")
		}
	})
	for _, p := range []any{serverConfig, dbConfig} {
		if err := env.ParseWithOptions(p, env.Options{Environment: environment}); err != nil {
			return nil, fmt.Errorf("parse env: %w", err)
		}
	}

	yc, err := loadYAML(*configPathFlag)
	if err != nil {
		return nil, fmt.Errorf("load yaml config: %w", err)
	}

	for _, v := range []validator{yc.Logger, yc.RLS, dbConfig, serverConfig, yc.Worker} {
		if err := v.Validate(); err != nil {
			return nil, fmt.Errorf("validate: %w", err)
		}
	}

	return &Config{
		Server: serverConfig,
		DB:     dbConfig,
		Logger: yc.Logger,
		RLS:    yc.RLS,
		W:      yc.Worker,
	}, nil

}
