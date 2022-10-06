package factory

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrepareIndices(t *testing.T) {
	t.Parallel()

	available := []string{"index1", "index2"}
	disabled := []string{"index2", "index3"}

	res := prepareIndices(available, disabled)
	require.Equal(t, []string{"index1"}, res)

	available = []string{"index1", "index2"}
	disabled = []string{}

	res = prepareIndices(available, disabled)
	require.Equal(t, []string{"index1", "index2"}, res)
}
