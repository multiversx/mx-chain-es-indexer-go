package transactions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArgumentsParserExtended_Split(t *testing.T) {
	t.Parallel()

	argsParserEx := newArgumentsParser()

	data1 := []byte("@aa@aa")
	res := argsParserEx.split(string(data1))
	require.Equal(t, 3, len(res))
	require.Equal(t, res[1], "aa")
	require.Equal(t, res[2], "aa")
}

func TestArgumentsParserExtended_HasOkPrefix(t *testing.T) {
	t.Parallel()

	argsParserEx := newArgumentsParser()

	require.True(t, argsParserEx.hasOKPrefix("@6f6b"))
	require.False(t, argsParserEx.hasOKPrefix("@"))
	require.False(t, argsParserEx.hasOKPrefix(""))
	require.False(t, argsParserEx.hasOKPrefix("aaa@aaa"))
	require.True(t, argsParserEx.hasOKPrefix("aaa@6f6b"))
	require.True(t, argsParserEx.hasOKPrefix("aaa@6f6b@aaa"))
}
