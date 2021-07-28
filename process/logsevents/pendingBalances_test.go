package logsevents

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/stretchr/testify/require"
)

func TestPendingBalancesProcessor(t *testing.T) {
	t.Parallel()

	pp := newPendingBalancesProcessor()

	pp.addInfo("receiver", "token", 10, "5")

	res := pp.getAll()
	require.Len(t, res, 1)
	require.Equal(t, &data.AccountInfo{
		Address:         "pending-receiver",
		Balance:         "5",
		TokenIdentifier: "token-0a",
		TokenName:       "token",
		TokenNonce:      10,
	}, res["pending-receiver-token-0a"])
}
