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
	require.Equal(t, uint64(2), tx1.Nonce)
	require.Equal(t, uint64(10), tx2.Nonce)
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
	scrMap := ConvertScrsSliceIntoMap(scrSlice)

	scrMap["h1"].Nonce = 5
	scrMap["h2"].Nonce = 10
	require.Equal(t, uint64(5), scr1.Nonce)
	require.Equal(t, uint64(10), scr2.Nonce)
}
