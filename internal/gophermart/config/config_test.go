package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// baseArgs — минимальный набор флагов
func baseArgs(configPath string) []string {
	return []string{
		"-d", "postgres://user:pass@localhost:5432/gophermart",
		"-a", ":8080",
		"-config", configPath,
	}
}

func TestLoad_Success(t *testing.T) {
	path := writeYAML(t, validYAML)

	cfg, err := Load(baseArgs(path))

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "postgres://user:pass@localhost:5432/gophermart", cfg.DB.URI)
	require.NotNil(t, cfg.Logger)
	require.NotNil(t, cfg.RLS)
	require.NotNil(t, cfg.W)
	assert.Equal(t, 16, cfg.W.WorkerCount)
}

func TestLoad_MissingConfigFile(t *testing.T) {
	args := baseArgs("/no/such/path/config.yaml")

	cfg, err := Load(args)

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.ErrorContains(t, err, "load yaml config")
}

func TestLoad_UnknownFlag(t *testing.T) {
	cfg, err := Load([]string{"-this-flag-does-not-exist"})

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.ErrorContains(t, err, "parse flags")
}

func TestLoad_ConfigPathFromEnv(t *testing.T) {
	path := writeYAML(t, validYAML)
	t.Setenv(configPathEnvVar, path)

	cfg, err := Load([]string{
		"-d", "postgres://user:pass@localhost:5432/gophermart",
		"-a", ":8080",
	})

	require.NoError(t, err)
	require.NotNil(t, cfg)
}

func TestLoad_ValidationFailure_MissingYAMLSection(t *testing.T) {
	path := writeYAML(t, `
logger:
  directory: logs
  stdout:
    enabled: false
rate_limits:
  cleanup_interval: 1m
  register:
    key_prefix: register
    window: 1m
    max_requests: 1
  login:
    key_prefix: login
    window: 1m
    max_requests: 1
`)

	cfg, err := Load(baseArgs(path))

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.ErrorContains(t, err, "section is required")
}

func TestLoad_ValidationFailure_InvalidField(t *testing.T) {
	path := writeYAML(t, `
logger:
  directory: logs
  stdout:
    enabled: false
rate_limits:
  cleanup_interval: 1m
  register:
    key_prefix: register
    window: 1m
    max_requests: 1
  login:
    key_prefix: login
    window: 1m
    max_requests: 1
worker:
  polling_interval: 15s
  worker_count: 0
  jobs_queue_size: 1
  target_rps: 1
`)

	cfg, err := Load(baseArgs(path))

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.ErrorContains(t, err, "worker count must be at least 1")
}

func TestLoad_ValidationFailure_EmptyDatabaseURI(t *testing.T) {
	path := writeYAML(t, validYAML)

	cfg, err := Load([]string{"-d", "", "-a", ":8080", "-config", path})

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.ErrorContains(t, err, "database URI can't be empty")
}

func TestLoad_TargetRPS_BelowOne(t *testing.T) {
	path := writeYAML(t, `
logger:
  directory: logs
  stdout:
    enabled: false
rate_limits:
  cleanup_interval: 1m
  register:
    key_prefix: register
    window: 1m
    max_requests: 1
  login:
    key_prefix: login
    window: 1m
    max_requests: 1
worker:
  polling_interval: 15s
  worker_count: 1
  jobs_queue_size: 1
  target_rps: 0.17
`)

	cfg, err := Load(baseArgs(path))

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.InDelta(t, 0.17, cfg.W.TargetRPS, 0.001)
}

func TestLoad_EnvOverridesFlag(t *testing.T) {
	path := writeYAML(t, validYAML)
	t.Setenv("DATABASE_URI", "postgres://from-env/db")

	cfg, err := Load([]string{
		"-d", "postgres://from-flag/db", // передан явно
		"-a", ":8080",
		"-config", path,
	})

	require.NoError(t, err)
	require.NotNil(t, cfg)
	// Env приоритет
	assert.Equal(t, "postgres://from-env/db", cfg.DB.URI)
}

func TestLoad_EnvUsedWhenFlagNotProvided(t *testing.T) {
	path := writeYAML(t, validYAML)
	t.Setenv("DATABASE_URI", "postgres://from-env/db")

	cfg, err := Load([]string{"-a", ":8080", "-config", path})

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "postgres://from-env/db", cfg.DB.URI)
}
