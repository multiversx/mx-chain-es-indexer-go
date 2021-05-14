package transactions

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/parsers"
	"github.com/stretchr/testify/require"
)

func TestArgumentsParserExtended_Split(t *testing.T) {
	t.Parallel()

	argsParserEx := newArgumentsParser(parsers.NewCallArgsParser())

	data1 := []byte("@aa@aa")
	res := argsParserEx.split(string(data1))
	require.Equal(t, 3, len(res))
	require.Equal(t, res[1], "aa")
	require.Equal(t, res[2], "aa")
}

func TestArgumentsParserExtended_HasOkPrefix(t *testing.T) {
	t.Parallel()

	argsParserEx := newArgumentsParser(parsers.NewCallArgsParser())

	require.True(t, argsParserEx.hasOKPrefix("@6f6b"))
	require.False(t, argsParserEx.hasOKPrefix("@"))
	require.False(t, argsParserEx.hasOKPrefix(""))
}
