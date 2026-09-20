package service

import (
	"strings"
	"testing"

	"github.com/alexedwards/argon2id"
	"github.com/stretchr/testify/require"
)

func TestArgon2Hasher(t *testing.T) {
	tests := []struct {
		name       string
		password   string
		candidate  string
		storedHash string
		wantMatch  bool
		wantErr    bool
	}{
		{name: "matching password", password: "correct horse battery staple", candidate: "correct horse battery staple", wantMatch: true},
		{name: "wrong password", password: "correct horse battery staple", candidate: "wrong password"},
		{name: "invalid encoded hash", candidate: "password", storedHash: "not-an-argon2-hash", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher := Argon2Hasher{}
			storedHash := tt.storedHash
			if storedHash == "" {
				var err error
				storedHash, err = hasher.Hash(tt.password)
				require.NoError(t, err)
				require.NotEqual(t, tt.password, storedHash)
				require.True(t, strings.HasPrefix(storedHash, "$argon2id$"))

				params, _, _, err := argon2id.DecodeHash(storedHash)
				require.NoError(t, err)
				require.Equal(t, uint32(19*1024), params.Memory)
				require.Equal(t, uint32(2), params.Iterations)
				require.Equal(t, uint8(1), params.Parallelism)
				require.Equal(t, uint32(16), params.SaltLength)
				require.Equal(t, uint32(32), params.KeyLength)
			}

			match, err := hasher.Compare(tt.candidate, storedHash)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantMatch, match)
		})
	}
}
