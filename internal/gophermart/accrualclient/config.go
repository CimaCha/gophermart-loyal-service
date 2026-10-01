package accrualclient

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Address string        `env:"ACCRUAL_SYSTEM_ADDRESS"`
	Timeout time.Duration `env:"ACCRUAL_SYSTEM_TIMEOUT"`
}

const defaultTimeout = 10 * time.Second

func RegisterFlags(fs *flag.FlagSet) *Config {
	cfg := &Config{Timeout: defaultTimeout}

	fs.StringVar(&cfg.Address, "r", "", "accrual system address")
	fs.DurationVar(&cfg.Timeout, "accrual-timeout", defaultTimeout, "accrual request timeout")

	return cfg
}

func (c *Config) ParseEnv() error {
	if err := env.Parse(c); err != nil {
		return fmt.Errorf("parse env: %w", err)
	}
	return nil
}

func (c *Config) Validate() error {
	if c == nil {
		return errors.New("accrual client config: section is required")
	}
	if c.Address == "" {
		return errors.New("accrual system address can't be empty")
	}
	if c.Timeout <= 0 {
		return errors.New("accrual system timeout must be greater than zero")
	}
	return nil
}
