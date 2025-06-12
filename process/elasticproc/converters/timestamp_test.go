package converters

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMillisecondsToSeconds(t *testing.T) {
	t.Parallel()

	require.Equal(t, uint64(1), MillisecondsToSeconds(1000))
	require.Equal(t, uint64(0), MillisecondsToSeconds(0))
	require.Equal(t, uint64(12), MillisecondsToSeconds(12_000))
	require.Equal(t, uint64(12), MillisecondsToSeconds(12_123))
}
