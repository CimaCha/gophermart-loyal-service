package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCredentials_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cred    Credentials
		wantErr error
	}{
		{
			name: "valid",
			cred: Credentials{
				Login:    "sayfargo",
				Password: "qwerty123",
			},
		},
		{
			name: "empty login",
			cred: Credentials{
				Login:    "",
				Password: "qwerty123",
			},
			wantErr: ErrEmptyLogin,
		},
		{
			name: "blank login",
			cred: Credentials{
				Login:    "   ",
				Password: "qwerty123",
			},
			wantErr: ErrEmptyLogin,
		},
		{
			name: "empty password",
			cred: Credentials{
				Login:    "sayfargo",
				Password: "",
			},
			wantErr: ErrEmptyPassword,
		},
		{
			name: "blank password",
			cred: Credentials{
				Login:    "sayfargo",
				Password: "   ",
			},
			wantErr: ErrEmptyPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cred.Validate()

			if tt.wantErr == nil {
				require.NoError(t, err)
				return
			}

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
