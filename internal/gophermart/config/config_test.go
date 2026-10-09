package config

import (
	"fmt"
	"strings"
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
		"-r",
		"http://localhost:8081",
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
	require.NotNil(t, cfg.Accrual)
	assert.Equal(t, "http://localhost:8081", cfg.Accrual.Address)

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
		"-r",
		"http://localhost:8081",
	})

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, 16, cfg.W.WorkerCount)
	assert.Equal(t, "http://localhost:8081", cfg.Accrual.Address)
}

func TestLoad_AccrualFlagHasPriorityOverEnv(t *testing.T) {
	t.Setenv(configPathEnvVar, "")

	path := writeYAML(t, validYAML)

	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://env:8081")

	cfg, err := Load([]string{
		"-d", "postgres://user:pass@localhost:5432/gophermart",
		"-a", ":8080",
		"-r", "http://flag:8081",
		"-config", path,
	})

	require.NoError(t, err)

	assert.Equal(t, "http://flag:8081", cfg.Accrual.Address)
}

func TestLoad_ConfigFlagHasPriorityOverEnv(t *testing.T) {
	envConfig := writeYAML(t, validYAML)
	flagConfig := writeYAML(t, strings.Replace(validYAML, "worker_count: 16", "worker_count: 999", 1))
	t.Setenv(configPathEnvVar, envConfig)
	cfg, err := Load(argsWithConfig(flagConfig))
	require.NoError(t, err)
	assert.Equal(t, 999, cfg.W.WorkerCount)
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

func TestLoad_FlagAndEnvPriority(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantAddr string
		wantURI  string
		wantErr  bool
	}{
		{name: "explicit flags override env", args: []string{"-a", ":9000", "-d", "postgres://flag"}, wantAddr: ":9000", wantURI: "postgres://flag"},
		{name: "env used without flags", wantAddr: ":9001", wantURI: "postgres://env"},
		{name: "only supplied flag overrides env", args: []string{"-a", ":9000"}, wantAddr: ":9000", wantURI: "postgres://env"},
		{name: "explicit default value overrides env", args: []string{"-a", "localhost:8080"}, wantAddr: "localhost:8080", wantURI: "postgres://env"},
		{name: "explicit empty flag overrides env", args: []string{"-d", ""}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(configPathEnvVar, writeYAML(t, validYAML))
			t.Setenv("RUN_ADDRESS", ":9001")
			t.Setenv("DATABASE_URI", "postgres://env")
			t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://env:8081")
			t.Setenv("ACCRUAL_SYSTEM_TIMEOUT", "10s")
			cfg, err := Load(tt.args)
			if tt.wantErr {
				require.ErrorContains(t, err, "database URI can't be empty")
				require.Nil(t, cfg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantAddr, cfg.Server.Addr)
			assert.Equal(t, tt.wantURI, cfg.DB.URI)
			assert.Equal(t, "logs/test", cfg.Logger.Directory)
		})
	}
}

func TestLoad_AccrualTimeoutFlagHasPriorityOverEnv(t *testing.T) {
	t.Setenv("ACCRUAL_SYSTEM_TIMEOUT", "30s")
	args := append(argsWithConfig(writeYAML(t, validYAML)), "-accrual-timeout", "3s")
	cfg, err := Load(args)
	require.NoError(t, err)
	assert.Equal(t, 3*time.Second, cfg.Accrual.Timeout)
}

func TestLoad_InvalidTimeoutEnvPriority(t *testing.T) {
	for _, withFlag := range []bool{false, true} {
		t.Run(fmt.Sprintf("flag=%t", withFlag), func(t *testing.T) {
			t.Setenv("ACCRUAL_SYSTEM_TIMEOUT", "bad-duration")
			args := argsWithConfig(writeYAML(t, validYAML))
			if withFlag {
				args = append(args, "-accrual-timeout", "3s")
			}
			cfg, err := Load(args)
			if withFlag {
				require.NoError(t, err)
				assert.Equal(t, 3*time.Second, cfg.Accrual.Timeout)
			} else {
				require.ErrorContains(t, err, "bad-duration")
				require.Nil(t, cfg)
			}
		})
	}
}
