package converters

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComputeTokenIdentifier(t *testing.T) {
	t.Parallel()

	require.Equal(t, "", ComputeTokenIdentifier("", 0))
	require.Equal(t, "", ComputeTokenIdentifier("token", 0))
	require.Equal(t, "my-token-01", ComputeTokenIdentifier("my-token", 1))
}
