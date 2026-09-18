package postgres

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	URI string `env:"DATABASE_URI"`
}

func RegisterFlags(fs *flag.FlagSet) *Config {

	postgresConfig := new(Config)

	fs.StringVar(
		&postgresConfig.URI,
		"d",
		"postgres://postgres:admin@localhost:5433/db",
		"database url connection",
	)

	return postgresConfig

}

func (c *Config) ParseEnv() error {
	if err := env.Parse(c); err != nil {
		return fmt.Errorf("parse env: %w", err)
	}
	return nil
}
