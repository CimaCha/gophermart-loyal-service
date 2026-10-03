package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/accrualclient"
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/ratelimitstore"
	"github.com/CimaCha/gophermart-loyal-service/internal/gophermart/worker"
	"github.com/CimaCha/gophermart-loyal-service/pkg/httpserver"
	"github.com/CimaCha/gophermart-loyal-service/pkg/postgres"
	"github.com/CimaCha/gophermart-loyal-service/pkg/slogger"
)

type Config struct {
	Server  *httpserver.Config
	Accrual *accrualclient.Config
	DB      *postgres.Config
	Logger  *slogger.Config
	RLS     *ratelimitstore.Config
	W       *worker.Config
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
	accrualConfig := accrualclient.RegisterFlags(fs)

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("parse flags: %w", err)
	}

	configPath := *configPathFlag
	if v := os.Getenv(configPathEnvVar); v != "" {
		configPath = v
	}

	for _, p := range []envParser{serverConfig, dbConfig, accrualConfig} {
		if err := p.ParseEnv(); err != nil {
			return nil, fmt.Errorf("parse env: %w", err)
		}
	}

	yc, err := loadYAML(configPath)
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
