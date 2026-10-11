package testenv

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestSetup(t *testing.T) {
	ctx := context.Background()

	fs := fstest.MapFS{
		"001_init.sql": {
			Data: []byte(`
-- +goose Up

CREATE TABLE test (
    id SERIAL PRIMARY KEY
);

-- +goose Down

DROP TABLE test;
`),
		},
	}

	env := Setup(ctx, fs)

	require.NotNil(t, env)
	require.NotNil(t, env.Pool)

	defer env.Close(ctx)

	err := env.Pool.Ping(ctx)
	require.NoError(t, err)
}

func TestEnvironment_Close(t *testing.T) {
	ctx := context.Background()

	env := &Environment{}

	require.NotPanics(t, func() {
		env.Close(ctx)
	})
}
