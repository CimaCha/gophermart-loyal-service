package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad_Success(t *testing.T) {
	t.Setenv(configPathEnvVar, "")

	cfg, err := Load([]string{
		"-config", "testdata/gophermart.yaml",
	})

	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.NotNil(t, cfg.Server)
	require.NotNil(t, cfg.DB)
	require.NotNil(t, cfg.Logger)
	require.NotNil(t, cfg.RLS)
}

func TestLoad_UsesEnvConfigPath(t *testing.T) {
	t.Setenv(configPathEnvVar, "testdata/gophermart.yaml")

	cfg, err := Load([]string{
		"-config", "does-not-exist.yaml",
	})

	require.NoError(t, err)
	require.NotNil(t, cfg)
}

func TestLoad_InvalidFlag(t *testing.T) {
	_, err := Load([]string{
		"-unknown",
	})

	require.Error(t, err)
	require.ErrorContains(t, err, "parse flags")
}

func TestLoad_ConfigNotFound(t *testing.T) {
	t.Setenv(configPathEnvVar, "")

	_, err := Load([]string{
		"-config", "testdata/not-found.yaml",
	})

	require.Error(t, err)
	require.ErrorContains(t, err, "slogger config")
}
