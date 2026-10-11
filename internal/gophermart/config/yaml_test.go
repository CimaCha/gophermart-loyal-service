package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validYAML = `
logger:
  directory: logs/test
  stdout:
    enabled: true
    format: text
    level: info

rate_limits:
  cleanup_interval: 5m
  routes:
    register:
      window: 1m
      max_requests: 5

worker:
  polling_interval: 15s
  worker_count: 16
  jobs_queue_size: 100
  target_rps: 50
`

func writeYAML(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")

	require.NoError(
		t,
		os.WriteFile(path, []byte(content), 0o644),
	)

	return path
}

func TestLoadYAML_FileSuccess(t *testing.T) {
	path := writeYAML(t, validYAML)

	cfg, err := loadYAML(path)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.NotNil(t, cfg.Logger)
	assert.Equal(t, "logs/test", cfg.Logger.Directory)

	require.NotNil(t, cfg.RLS)
	assert.NotNil(t, cfg.RLS.Routes["register"])

	require.NotNil(t, cfg.Worker)
	assert.Equal(t, 16, cfg.Worker.WorkerCount)
	assert.Equal(t, 100, cfg.Worker.JobsQueueSize)
	assert.Equal(t, float64(50), cfg.Worker.TargetRPS)
}

func TestLoadYAML_FileNotFoundUsesEmbeddedConfig(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"not-exists.yaml",
	)

	cfg, err := loadYAML(path)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.NotNil(t, cfg.Logger)
	assert.NotNil(t, cfg.RLS)
	assert.NotNil(t, cfg.Worker)
}

func TestLoadYAML_EmptyFileUsesEmbeddedConfig(t *testing.T) {
	path := writeYAML(t, "")

	cfg, err := loadYAML(path)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.NotNil(t, cfg.Logger)
	assert.NotNil(t, cfg.RLS)
	assert.NotNil(t, cfg.Worker)
}

func TestLoadYAML_InvalidSyntax(t *testing.T) {
	path := writeYAML(
		t,
		`
logger:
  - invalid
`,
	)

	cfg, err := loadYAML(path)

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.ErrorContains(t, err, "unmarshal config file")
}

func TestLoad_DefaultEmbeddedConfig(t *testing.T) {
	t.Setenv("RUN_ADDRESS", "localhost:8080")
	t.Setenv("DATABASE_URI", "postgres://test")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://localhost:8081")
	t.Setenv("ACCRUAL_SYSTEM_TIMEOUT", "10s")

	cfg, err := Load(nil)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "localhost:8080", cfg.Server.Addr)
	assert.Equal(t, "postgres://test", cfg.DB.URI)
	assert.Equal(t, "http://localhost:8081", cfg.Accrual.Address)

	assert.NotNil(t, cfg.Logger)
	assert.NotNil(t, cfg.RLS)
	assert.NotNil(t, cfg.W)
}

func TestLoad_ConfigPathFlag(t *testing.T) {
	path := writeYAML(t, validYAML)

	t.Setenv("RUN_ADDRESS", "localhost:8080")
	t.Setenv("DATABASE_URI", "postgres://test")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://localhost:8081")
	t.Setenv("ACCRUAL_SYSTEM_TIMEOUT", "10s")

	cfg, err := Load([]string{
		"-config",
		path,
	})

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "logs/test", cfg.Logger.Directory)
	assert.Equal(t, 16, cfg.W.WorkerCount)
}

func TestLoad_ConfigPathEnv(t *testing.T) {
	path := writeYAML(t, validYAML)

	t.Setenv(configPathEnvVar, path)

	t.Setenv("RUN_ADDRESS", "localhost:8080")
	t.Setenv("DATABASE_URI", "postgres://test")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://localhost:8081")
	t.Setenv("ACCRUAL_SYSTEM_TIMEOUT", "10s")

	cfg, err := Load(nil)

	require.NoError(t, err)

	assert.Equal(t, "logs/test", cfg.Logger.Directory)
	assert.Equal(t, 16, cfg.W.WorkerCount)
}
