package converters

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/stretchr/testify/require"
)

func TestMergeAccountsInfoMaps(t *testing.T) {
	t.Parallel()

	m1 := map[string]*data.AccountInfo{
		"k1": {
			Balance: "1",
		},
		"k2": {
			Balance: "2",
		},
	}

	m2 := map[string]*data.AccountInfo{
		"k3": {
			Balance: "3",
		},
		"k4": {
			Balance: "4",
		},
	}

	res := MergeAccountsInfoMaps(m1, m2)
	require.Len(t, res, 4)
}
