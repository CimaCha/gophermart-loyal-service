package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validYAML = `
logger:
  directory: logs/accrual
  stdout:
    enabled: true
    format: text
    level: info
  files:
    - name: app
      enabled: true
      format: json
      level: debug

rate_limits:
  cleanup_interval: 30m

  routes:
    getorders:
      window: 1m
      max_requests: 1024
`

// writeYAML создаёт временный файл с содержимым и возвращает путь к нему
func writeYAML(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")

	require.NoError(
		t,
		os.WriteFile(path, []byte(content), 0o644),
	)

	return path
}

func TestLoadYAML_Success(t *testing.T) {
	path := writeYAML(t, validYAML)

	cfg, err := loadYAML(path)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Logger
	require.NotNil(t, cfg.Logger)
	assert.Equal(t, "logs/accrual", cfg.Logger.Directory)
	assert.True(t, cfg.Logger.Stdout.Enabled)
	require.Len(t, cfg.Logger.Files, 1)
	assert.Equal(t, "app", cfg.Logger.Files[0].Name)

	// RateLimits
	require.NotNil(t, cfg.RLS)
	assert.Equal(t, 30*time.Minute, cfg.RLS.CleanupInterval)

	_, ok := cfg.RLS.Routes["getorders"]
	assert.True(t, ok)

	assert.Equal(t, time.Minute, cfg.RLS.Routes["getorders"].Window)
	assert.Equal(t, 1024, cfg.RLS.Routes["getorders"].MaxRequests)
}

func TestLoadYAML_FileNotFound(t *testing.T) {
	cfg, err := loadYAML(filepath.Join(t.TempDir(), "does-not-exist.yaml"))

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.ErrorContains(t, err, "read config file")
}

func TestLoadYAML_InvalidSyntax(t *testing.T) {
	path := writeYAML(t, "logger: [this is not: a valid, mapping")

	cfg, err := loadYAML(path)

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.ErrorContains(t, err, "unmarshal config file")
}

// TestLoadYAML_MissingSection проверяет, что nil конфиг не валид
func TestLoadYAML_MissingSection(t *testing.T) {
	tests := []struct {
		name    string
		content string
		check   func(t *testing.T, cfg *yamlConfig)
	}{
		{
			name: "worker section missing",
			content: `
logger:
  directory: logs
  stdout:
    enabled: false
rate_limits:
  cleanup_interval: 30m

  routes:
    getorders:
      window: 1m
      max_requests: 1024
`,
			check: func(t *testing.T, cfg *yamlConfig) {
				assert.NotNil(t, cfg.Logger)
				assert.NotNil(t, cfg.RLS)
			},
		},
		{
			name:    "all sections missing (empty file)",
			content: ``,
			check: func(t *testing.T, cfg *yamlConfig) {
				assert.Nil(t, cfg.Logger)
				assert.Nil(t, cfg.RLS)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := writeYAML(t, tc.content)

			cfg, err := loadYAML(path)

			require.NoError(t, err)
			require.NotNil(t, cfg)
			tc.check(t, cfg)
		})
	}
}
