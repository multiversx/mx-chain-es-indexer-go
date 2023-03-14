package validators

import (
	"testing"

	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestNewValidatorsProcessor(t *testing.T) {
	t.Parallel()

	vp, err := NewValidatorsProcessor(nil, 0)
	require.Nil(t, vp)
	require.Equal(t, dataindexer.ErrNilPubkeyConverter, err)
}

// TODO add unit tests
