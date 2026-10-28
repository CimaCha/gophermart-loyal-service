package accrualclient

import (
	"flag"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr string
	}{
		{
			name: "valid config",
			cfg: &Config{
				Address: "http://localhost:8081",
				Timeout: time.Second,
			},
		},
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: "accrual client config: section is required",
		},
		{
			name: "empty address",
			cfg: &Config{
				Address: "",
				Timeout: time.Second,
			},
			wantErr: "accrual system address can't be empty",
		},
		{
			name: "zero timeout",
			cfg: &Config{
				Address: "http://localhost:8081",
				Timeout: 0,
			},
			wantErr: "accrual system timeout must be greater than zero",
		},
		{
			name: "negative timeout",
			cfg: &Config{
				Address: "http://localhost:8081",
				Timeout: -time.Second,
			},
			wantErr: "accrual system timeout must be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()

			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}

			require.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestRegisterFlags(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	cfg := RegisterFlags(fs)

	require.Equal(t, defaultTimeout, cfg.Timeout)
	require.Empty(t, cfg.Address)

	err := fs.Parse([]string{
		"-r", "http://localhost:8081",
		"-accrual-timeout", "5s",
	})

	require.NoError(t, err)

	require.Equal(t, "http://localhost:8081", cfg.Address)
	require.Equal(t, 5*time.Second, cfg.Timeout)
}
