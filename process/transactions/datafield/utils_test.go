package datafield

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsASCIIString(t *testing.T) {
	t.Parallel()

	require.True(t, isASCIIString("hello"))
	require.True(t, isASCIIString("TOKEN-abcd"))
	require.False(t, isASCIIString(string([]byte{12, 255})))
	require.False(t, isASCIIString(string([]byte{12, 188})))
}
