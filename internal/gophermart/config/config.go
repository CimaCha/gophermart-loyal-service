package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/CimaCha/gophermart-loyal-service/internal/shared/db/postgres"
	"github.com/CimaCha/gophermart-loyal-service/internal/shared/httpserver"
	"github.com/CimaCha/gophermart-loyal-service/internal/shared/ratelimit"
	"github.com/CimaCha/gophermart-loyal-service/internal/shared/slogger"
)

type Config struct {
	Server *httpserver.Config
	DB     *postgres.Config
	Logger *slogger.Config
	RLS    *ratelimit.Config
}

type envParser interface {
	ParseEnv() error
}

const (
	defaultConfigPath = "config/gophermart.yaml"
	configPathEnvVar  = "GOPHERMART_CONFIG_PATH"
)

func Load(args []string) (*Config, error) {

	fs := flag.NewFlagSet("config", flag.ContinueOnError)

	configPathFlag := fs.String("config", defaultConfigPath, "path to config file")
	serverConfig := httpserver.RegisterFlags(fs)
	dbConfig := postgres.RegisterFlags(fs)

	// Парсим флаги
	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("parse flags: %w", err)
	}

	configPath := *configPathFlag
	if v := os.Getenv(configPathEnvVar); v != "" {
		configPath = v
	}

	// Парсим env если есть, то они приоритет
	for _, p := range []envParser{serverConfig, dbConfig} {
		if err := p.ParseEnv(); err != nil {
			return nil, fmt.Errorf("parse env: %w", err)
		}
	}

	logConfig, err := slogger.LoadFromYAML(configPath)
	if err != nil {
		return nil, fmt.Errorf("slogger config: %w", err)
	}

	rateLimitStorage, err := ratelimit.LoadFromYAML(configPath)
	if err != nil {
		return nil, fmt.Errorf("rate limit config: %w", err)
	}

	return &Config{
		Server: serverConfig,
		DB:     dbConfig,
		Logger: logConfig,
		RLS:    rateLimitStorage,
	}, nil

}
