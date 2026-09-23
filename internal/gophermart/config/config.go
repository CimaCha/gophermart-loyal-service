package config

import (
	"flag"
	"fmt"

	"github.com/CimaCha/gophermart-loyal-service/internal/shared/db/postgres"
	"github.com/CimaCha/gophermart-loyal-service/internal/shared/httpserver"
	"github.com/CimaCha/gophermart-loyal-service/internal/shared/ratelimit"
	"github.com/CimaCha/gophermart-loyal-service/internal/shared/slogger"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Server *httpserver.Config
	DB     *postgres.Config
	Logger *slogger.Config
	RLS    *ratelimit.Config
}

// для получения config.yaml из флагов или env
type ConfigLoader struct {
	Path string `env:"GOPHERMART_CONFIG_PATH"`
}

func Load(args []string) (*Config, error) {

	fs := flag.NewFlagSet("config", flag.ContinueOnError)

	serverConfig := httpserver.RegisterFlags(fs)
	dbConfig := postgres.RegisterFlags(fs)
	configLoader := RegisterConfigPathFlag(fs)

	// Парсим флаги
	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}

	// Парсим env если есть, то они приоритет
	if err := configLoader.ParseEnv(); err != nil {
		return nil, fmt.Errorf("loader config: %w", err)
	}

	if err := serverConfig.ParseEnv(); err != nil {
		return nil, fmt.Errorf("server config: %w", err)
	}

	if err := dbConfig.ParseEnv(); err != nil {
		return nil, fmt.Errorf("db config: %w", err)
	}

	logConfig, err := slogger.LoadFromYAML(configLoader.Path)
	if err != nil {
		return nil, fmt.Errorf("slogger config: %w", err)
	}

	rateLimitStorage, err := ratelimit.LoadFromYAML(configLoader.Path)
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

func RegisterConfigPathFlag(fs *flag.FlagSet) *ConfigLoader {
	cfg := new(ConfigLoader)

	fs.StringVar(
		&cfg.Path,
		"config",
		"config/gophermart.yaml",
		"path to config file",
	)

	return cfg
}

func (c *ConfigLoader) ParseEnv() error {
	if err := env.Parse(c); err != nil {
		return fmt.Errorf("failed to parse env: %w", err)
	}
	return nil
}
