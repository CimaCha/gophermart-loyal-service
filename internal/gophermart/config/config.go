package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/worker"
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
	W      *worker.Config
}

type envParser interface {
	ParseEnv() error
}

type validator interface {
	Validate() error
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

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("parse flags: %w", err)
	}

	fmt.Println("ARGS:", args)
	fmt.Println("CONFIG FLAG:", *configPathFlag)

	configPath := *configPathFlag

	fmt.Println("BEFORE ENV:", configPath)

	if v := os.Getenv(configPathEnvVar); v != "" {
		configPath = v
	}

	fmt.Println("AFTER ENV:", configPath)

	for _, p := range []envParser{serverConfig, dbConfig} {
		if err := p.ParseEnv(); err != nil {
			return nil, fmt.Errorf("parse env: %w", err)
		}
	}

	yc, err := loadYAML(configPath)
	if err != nil {
		return nil, fmt.Errorf("load yaml config: %w", err)
	}

	for _, v := range []validator{yc.Logger, yc.RLS, yc.Worker, dbConfig, serverConfig} {
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
