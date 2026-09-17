package config

import (
	"flag"
	"github.com/caarlos0/env/v11"
	"os"
)

type Config struct {
	Address          string `env:"RUN_ADDRESS"`
	DatabaseURL      string `env:"DATABASE_URI"`
	AccrualSystemURL string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func New() (*Config, error) {
	address := flag.String("a", "localhost:8080", "address of service")
	databaseURL := flag.String("d", "http://localhost:8080", "url fordatabase")
	accrualSystemURL := flag.String("r", "", "path to the storage file")
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	config := Config{
		Address:          *address,
		DatabaseURL:      *databaseURL,
		AccrualSystemURL: *accrualSystemURL,
	}
	if err := env.Parse(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
