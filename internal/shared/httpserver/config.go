package httpserver

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Addr string `env:"RUN_ADDRESS"`
}

func RegisterFlags(fs *flag.FlagSet) *Config {

	serverCfg := new(Config)

	fs.StringVar(
		&serverCfg.Addr,
		"a",
		"localhost:8080",
		"Server address",
	)

	return serverCfg

}

func (c *Config) ParseEnv() error {
	if err := env.Parse(c); err != nil {
		return fmt.Errorf("failed to parse env: %w", err)
	}
	return nil
}
