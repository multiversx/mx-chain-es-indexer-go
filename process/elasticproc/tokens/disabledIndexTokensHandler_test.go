package tokens

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewIndexTokensHandler(t *testing.T) {
	t.Parallel()

	ith := NewDisabledIndexTokensHandler()
	require.False(t, ith.IsInterfaceNil())
}

func TestIndexTokensHandler_IndexCrossChainTokens(t *testing.T) {
	t.Parallel()

	ith := NewDisabledIndexTokensHandler()
	err := ith.IndexCrossChainTokens(nil, nil, nil)
	require.NoError(t, err)
}
