package logsevents

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestScDeploysProcessor(t *testing.T) {
	t.Parallel()

	scDeploysProc := newSCDeploysProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.SCDeployIdentifier),
		Topics:     [][]byte{[]byte("addr1"), []byte("addr2")},
	}

	scDeploys := map[string]*data.ScDeployInfo{}
	_, processed := scDeploysProc.processEvent(&argsProcessEvent{
		event:            event,
		timestamp:        1000,
		scDeploys:        scDeploys,
		txHashHexEncoded: "01020304",
	})
	require.True(t, processed)

	require.Equal(t, &data.ScDeployInfo{
		TxHash:    "01020304",
		Creator:   "6164647232",
		Timestamp: uint64(1000),
	}, scDeploys["6164647231"])
}
