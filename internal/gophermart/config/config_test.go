package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func argsWithConfig(path string) []string {
	return []string{
		"-d",
		"postgres://user:pass@localhost:5432/gophermart",
		"-a",
		":8080",
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
		"postgres://user:pass@localhost:5432/gophermart",
		cfg.DB.URI,
	)

	assert.Equal(t, ":8080", cfg.Server.Addr)

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
		"postgres://user:pass@localhost:5432/gophermart",
		"-a",
		":8080",
	})

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, 16, cfg.W.WorkerCount)
}

func TestLoad_EnvHasPriorityOverFlag(t *testing.T) {
	t.Setenv(configPathEnvVar, "")

	envConfig := writeYAML(t, validYAML)

	flagConfig := writeYAML(t, `
worker:
  polling_interval: 1s
  worker_count: 999
  jobs_queue_size: 1
  target_rps: 1
`)

	t.Setenv(configPathEnvVar, envConfig)

	cfg, err := Load(argsWithConfig(flagConfig))

	require.NoError(t, err)

	assert.Equal(t, 16, cfg.W.WorkerCount)
}

func TestLoad_ConfigFileNotFound(t *testing.T) {
	t.Setenv(configPathEnvVar, "")

	cfg, err := Load(
		argsWithConfig("/does/not/exist/config.yaml"),
	)

	require.Error(t, err)
	assert.Nil(t, cfg)

	assert.ErrorContains(t, err, "load yaml config")
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

func TestLoad_InvalidWorkerConfig(t *testing.T) {
	t.Setenv(configPathEnvVar, "")

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

	cfg, err := Load(argsWithConfig(path))

	require.Error(t, err)
	assert.Nil(t, cfg)

	assert.ErrorContains(t, err, "worker count must be at least 1")
}

func TestLoad_InvalidDatabaseURI(t *testing.T) {
	t.Setenv(configPathEnvVar, "")

	path := writeYAML(t, validYAML)

	cfg, err := Load([]string{
		"-d",
		"",
		"-a",
		":8080",
		"-config",
		path,
	})

	require.Error(t, err)
	assert.Nil(t, cfg)

	assert.ErrorContains(t, err, "database URI can't be empty")
}

func TestLoad_TargetRPSFloat(t *testing.T) {
	t.Setenv(configPathEnvVar, "")

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

	cfg, err := Load(argsWithConfig(path))

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.InDelta(
		t,
		0.17,
		cfg.W.TargetRPS,
		0.001,
	)

	assert.Equal(
		t,
		15*time.Second,
		cfg.W.PollingInterval,
	)
}
