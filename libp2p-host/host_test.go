package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHost(t *testing.T) {
	ctx := context.Background()
	h1, err := createHost1(ctx)
	require.NoError(t, err)

	h2, err := createHost2(ctx)
	require.NoError(t, err)

	require.NotEmpty(t, h1.ID())
	require.NotEmpty(t, h2.ID())
}
