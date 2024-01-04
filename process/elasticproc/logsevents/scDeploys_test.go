package logsevents

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/stretchr/testify/require"
)

func TestScDeploysProcessor(t *testing.T) {
	t.Parallel()

	scDeploysProc := newSCDeploysProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.SCDeployIdentifier),
		Topics:     [][]byte{[]byte("addr1"), []byte("addr2"), []byte("codeHash")},
	}

	scDeploys := map[string]*data.ScDeployInfo{}
	res := scDeploysProc.processEvent(&argsProcessEvent{
		event:            event,
		timestamp:        1000,
		scDeploys:        scDeploys,
		txHashHexEncoded: "01020304",
	})
	require.True(t, res.processed)

	require.Equal(t, &data.ScDeployInfo{
		TxHash:       "01020304",
		Creator:      "6164647232",
		Timestamp:    uint64(1000),
		CurrentOwner: "6164647232",
		CodeHash:     []byte("codeHash"),
	}, scDeploys["6164647231"])
}

func TestScDeploysProcessorChangeOwner(t *testing.T) {
	event := &transaction.Event{
		Address:    []byte("contractAddr"),
		Identifier: []byte(core.BuiltInFunctionChangeOwnerAddress),
		Topics:     [][]byte{[]byte("newOwner")},
	}

	scDeploysProc := newSCDeploysProcessor(&mock.PubkeyConverterMock{})

	changeOwnerOperations := map[string]*data.OwnerData{}
	res := scDeploysProc.processEvent(&argsProcessEvent{
		event:                 event,
		changeOwnerOperations: changeOwnerOperations,
		timestamp:             2000,
		txHashHexEncoded:      "01020304",
	})
	require.True(t, res.processed)

	require.Equal(t, &data.OwnerData{
		TxHash:    "01020304",
		Address:   "6e65774f776e6572",
		Timestamp: 2000,
	}, changeOwnerOperations["636f6e747261637441646472"])
}
