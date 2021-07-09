package converters

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/stretchr/testify/require"
)

func TestConvertTxsSliceIntoMap(t *testing.T) {
	t.Parallel()

	tx1 := &data.Transaction{
		Hash:  "h1",
		Nonce: 1,
	}
	tx2 := &data.Transaction{
		Hash:  "h2",
		Nonce: 2,
	}
	txsSlice := []*data.Transaction{tx1, tx2}
	txsMap := ConvertTxsSliceIntoMap(txsSlice)

	txsMap["h1"].Nonce = 2
	txsMap["h2"].Nonce = 10
	require.True(t, tx1 == txsMap["h1"]) // pointer testing
	require.True(t, tx2 == txsMap["h2"]) // pointer testing
	require.Equal(t, len(txsSlice), len(txsMap))
}

func TestConvertScrsSliceIntoMap(t *testing.T) {
	t.Parallel()

	scr1 := &data.ScResult{
		Hash:  "h1",
		Nonce: 1,
	}
	scr2 := &data.ScResult{
		Hash:  "h2",
		Nonce: 2,
	}
	scrSlice := []*data.ScResult{scr1, scr2}
	scrsMap := ConvertScrsSliceIntoMap(scrSlice)

	scrsMap["h1"].Nonce = 5
	scrsMap["h2"].Nonce = 10
	require.True(t, scr1 == scrsMap["h1"]) // pointer testing
	require.True(t, scr2 == scrsMap["h2"]) // pointer testing
	require.Equal(t, len(scrSlice), len(scrsMap))
}
