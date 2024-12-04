package transactions

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"
)

func TestNewTxHashExtractor(t *testing.T) {
	t.Parallel()

	the := NewTxHashExtractor()
	require.False(t, the.IsInterfaceNil())
}

func TestTxHashExtractor_ExtractExecutedTxHashes(t *testing.T) {
	t.Parallel()

	the := NewTxHashExtractor()
	mbTxHashes := [][]byte{[]byte("hash1"), []byte("hash2")}
	txHashes := the.ExtractExecutedTxHashes(0, mbTxHashes, &block.Header{})
	require.Equal(t, mbTxHashes, txHashes)
}
