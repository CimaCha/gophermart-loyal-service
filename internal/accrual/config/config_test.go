package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func argsWithConfig(path string) []string {
	return []string{
		"-d",
		"postgres://user:pass@localhost:5433/accrual",
		"-a",
		":8081",
		"-config",
		path,
	}
}

func TestLoad_Success(t *testing.T) {
	t.Setenv(configPathEnvVar, "")

	path := writeYAML(t, validYAML)

	cfg, err := Load(argsWithConfig(path))

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(
		t,
		"postgres://user:pass@localhost:5433/accrual",
		cfg.DB.URI,
	)

	assert.Equal(t, ":8081", cfg.Server.Addr)

	require.NotNil(t, cfg.Logger)
	require.NotNil(t, cfg.RLS)
	require.NotNil(t, cfg.W)

	assert.Equal(t, 16, cfg.W.WorkerCount)
}

func TestLoad_ConfigPathFromEnv(t *testing.T) {
	path := writeYAML(t, validYAML)

	t.Setenv(configPathEnvVar, path)

	cfg, err := Load([]string{
		"-d",
		"postgres://user:pass@localhost:5433/accrual",
		"-a",
		":8081",
	})

	require.NoError(t, err)
	require.NotNil(t, cfg)
}

func TestLoad_InvalidYAML(t *testing.T) {
	t.Setenv(configPathEnvVar, "")

	path := writeYAML(
		t,
		"logger: [invalid",
	)

	cfg, err := Load(argsWithConfig(path))

	require.Error(t, err)
	assert.Nil(t, cfg)

	assert.ErrorContains(t, err, "unmarshal config file")
}

func TestLoad_MissingSection(t *testing.T) {
	t.Setenv(configPathEnvVar, "")

	path := writeYAML(t, `
logger:
  directory: logs
  stdout:
    enabled: false
`)

	cfg, err := Load(argsWithConfig(path))

	require.Error(t, err)
	assert.Nil(t, cfg)

	assert.ErrorContains(t, err, "section is required")
}

func TestLoad_InvalidDatabaseURI(t *testing.T) {
	t.Setenv(configPathEnvVar, "")

	path := writeYAML(t, validYAML)

	cfg, err := Load([]string{
		"-d",
		"",
		"-a",
		":8081",
		"-config",
		path,
	})

	require.Error(t, err)
	assert.Nil(t, cfg)

	assert.ErrorContains(t, err, "database URI can't be empty")
}
