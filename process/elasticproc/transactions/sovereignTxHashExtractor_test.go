package transactions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSovereignTxHashExtractor(t *testing.T) {
	t.Parallel()

	sthe := NewSovereignTxHashExtractor()
	require.False(t, sthe.IsInterfaceNil())
}

func TestSovereignTxHashExtractor_ExtractExecutedTxHashes(t *testing.T) {
	t.Parallel()

	sthe := NewSovereignTxHashExtractor()
	mbTxHashes := [][]byte{[]byte("hash1"), []byte("hash2")}
	txHashes := sthe.ExtractExecutedTxHashes(0, mbTxHashes, nil)
	require.Equal(t, mbTxHashes, txHashes)
}
