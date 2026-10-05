package worker

import (
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
				PollingInterval: time.Second,
				WorkerCount:     1,
				JobsQueueSize:   1,
				TargetRPS:       1,
			},
		},
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: "worker config: section is required",
		},
		{
			name: "zero polling interval",
			cfg: &Config{
				PollingInterval: 0,
				WorkerCount:     1,
				JobsQueueSize:   1,
				TargetRPS:       1,
			},
			wantErr: "polling interval must be greater than zero",
		},
		{
			name: "negative polling interval",
			cfg: &Config{
				PollingInterval: -time.Second,
				WorkerCount:     1,
				JobsQueueSize:   1,
				TargetRPS:       1,
			},
			wantErr: "polling interval must be greater than zero",
		},
		{
			name: "zero worker count",
			cfg: &Config{
				PollingInterval: time.Second,
				WorkerCount:     0,
				JobsQueueSize:   1,
				TargetRPS:       1,
			},
			wantErr: "worker count must be at least 1",
		},
		{
			name: "negative worker count",
			cfg: &Config{
				PollingInterval: time.Second,
				WorkerCount:     -1,
				JobsQueueSize:   1,
				TargetRPS:       1,
			},
			wantErr: "worker count must be at least 1",
		},
		{
			name: "zero jobs queue size",
			cfg: &Config{
				PollingInterval: time.Second,
				WorkerCount:     1,
				JobsQueueSize:   0,
				TargetRPS:       1,
			},
			wantErr: "jobs queue size must be at least 1",
		},
		{
			name: "negative jobs queue size",
			cfg: &Config{
				PollingInterval: time.Second,
				WorkerCount:     1,
				JobsQueueSize:   -1,
				TargetRPS:       1,
			},
			wantErr: "jobs queue size must be at least 1",
		},
		{
			name: "zero target RPS",
			cfg: &Config{
				PollingInterval: time.Second,
				WorkerCount:     1,
				JobsQueueSize:   1,
				TargetRPS:       0,
			},
			wantErr: "target RPS must be greater than zero",
		},
		{
			name: "negative target RPS",
			cfg: &Config{
				PollingInterval: time.Second,
				WorkerCount:     1,
				JobsQueueSize:   1,
				TargetRPS:       -1,
			},
			wantErr: "target RPS must be greater than zero",
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
