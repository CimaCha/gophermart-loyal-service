// Package config отвечает за сборку, парсинг и валидацию глобальной конфигурации
// основного сервиса gophermart из флагов командной строки, переменных окружения и YAML-файла.
package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/accrualclient"
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/worker"
	"github.com/CimaCha/gophermart-loyal-service/pkg/httpserver"
	"github.com/CimaCha/gophermart-loyal-service/pkg/postgres"
	"github.com/CimaCha/gophermart-loyal-service/pkg/ratelimitstore"
	"github.com/CimaCha/gophermart-loyal-service/pkg/slogger"
)

// Config объединяет в себе конфигурационные параметры всех подсистем gophermart:
// HTTP-сервера, клиента системы начислений, базы данных, логгера, лимитера запросов и воркеров.
type Config struct {
	// Server содержит параметры запуска HTTP-сервера gophermart.
	Server *httpserver.Config
	// Accrual хранит настройки подключения к внешнему микросервису расчета баллов.
	Accrual *accrualclient.Config
	// DB содержит строку подключения и параметры пула для PostgreSQL.
	DB *postgres.Config
	// Logger управляет конфигурацией вывода структурированных логов.
	Logger *slogger.Config
	// RLS содержит настройки InMemory-хранилища лимитов частоты запросов.
	RLS *ratelimitstore.Config
	// W определяет параметры производительности и таймингов фоновых воркеров.
	W *worker.Config
}

type validator interface {
	Validate() error
}

const (
	defaultConfigPath = "config/gophermart.yaml"
	configPathEnvVar  = "GOPHERMART_CONFIG_PATH"
)

// Load выполняет пошаговую инициализацию, слияние и валидацию полной конфигурации сервиса.
// Приоритет применения источников (от высшего к низшему):
// 1. Явно переданные флаги командной строки (CLI Flags)
// 2. Переменные окружения (Environment Variables)
// 3. Значения из файла конфигурации YAML (если путь передан)
// 4. Дефолтные значения внутренних структур.
func Load(args []string) (*Config, error) {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)

	configPath := defaultConfigPath
	if v := os.Getenv(configPathEnvVar); v != "" {
		configPath = v
	}
	configPathFlag := fs.String("config", configPath, "path to config file")
	serverConfig := httpserver.RegisterFlags(fs)
	dbConfig := postgres.RegisterFlags(fs)
	accrualConfig := accrualclient.RegisterFlags(fs)

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
		case "r":
			delete(environment, "ACCRUAL_SYSTEM_ADDRESS")
		case "accrual-timeout":
			delete(environment, "ACCRUAL_SYSTEM_TIMEOUT")
		}
	})
	for _, p := range []any{serverConfig, dbConfig, accrualConfig} {
		if err := env.ParseWithOptions(p, env.Options{Environment: environment}); err != nil {
			return nil, fmt.Errorf("parse env: %w", err)
		}
	}

	yc, err := loadYAML(*configPathFlag)
	if err != nil {
		return nil, fmt.Errorf("load yaml config: %w", err)
	}

	for _, v := range []validator{yc.Logger, yc.RLS, yc.Worker, dbConfig, serverConfig, accrualConfig} {
		if err := v.Validate(); err != nil {
			return nil, fmt.Errorf("validate: %w", err)
		}
	}

	return &Config{
		Server:  serverConfig,
		Accrual: accrualConfig,
		DB:      dbConfig,
		Logger:  yc.Logger,
		RLS:     yc.RLS,
		W:       yc.Worker,
	}, nil

}
