package logsevents

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go/core"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestProcessLogsAndEventsESDT_IntraShard(t *testing.T) {
	t.Parallel()

	fungibleProc := newFungibleESDTProcessor(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{})

	logsAndEvents := map[string]nodeData.LogHandler{
		"txHash": &transaction.Log{
			Events: []*transaction.Event{
				{
					Address:    []byte("addr"),
					Identifier: []byte(core.BuiltInFunctionESDTTransfer),
					Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).SetUint64(100).Bytes(), []byte("receiver")},
				},
			},
		},
	}

	altered := data.NewAlteredAccounts()

	txs := make(map[string]*data.Transaction)
	encodedHash := hex.EncodeToString([]byte("txHash"))
	txs[encodedHash] = &data.Transaction{}
	scrs := make(map[string]*data.ScResult)
	fungibleProc.processLogsAndEventsESDT(logsAndEvents, altered, txs, scrs)

	alteredAddrSender, ok := altered.Get("61646472")
	require.True(t, ok)
	require.Equal(t, &data.AlteredAccount{
		IsESDTOperation: true,
		TokenIdentifier: "my-token",
	}, alteredAddrSender[0])

	alteredAddrReceiver, ok := altered.Get("7265636569766572")
	require.True(t, ok)
	require.Equal(t, &data.AlteredAccount{
		IsESDTOperation: true,
		TokenIdentifier: "my-token",
	}, alteredAddrReceiver[0])

	require.Equal(t, "my-token", txs[encodedHash].EsdtTokenIdentifier)
}

func TestProcessLogsAndEventsESDT_CrossShardOnSource(t *testing.T) {
	t.Parallel()

	receiverAddr := []byte("receiver")
	fungibleProc := newFungibleESDTProcessor(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{
		ComputeIdCalled: func(address []byte) uint32 {
			if bytes.Equal(address, receiverAddr) {
				return 1
			}
			return 0
		},
	})

	logsAndEvents := map[string]nodeData.LogHandler{
		"txHash": &transaction.Log{
			Events: []*transaction.Event{
				{
					Address:    []byte("addr"),
					Identifier: []byte(core.BuiltInFunctionESDTTransfer),
					Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).SetUint64(100).Bytes(), receiverAddr},
				},
			},
		},
	}

	altered := data.NewAlteredAccounts()

	txs := make(map[string]*data.Transaction)
	scrs := make(map[string]*data.ScResult)
	encodedHash := hex.EncodeToString([]byte("txHash"))
	scrs[encodedHash] = &data.ScResult{}
	fungibleProc.processLogsAndEventsESDT(logsAndEvents, altered, txs, scrs)

	alteredAddrSender, ok := altered.Get("61646472")
	require.True(t, ok)
	require.Equal(t, &data.AlteredAccount{
		IsESDTOperation: true,
		TokenIdentifier: "my-token",
	}, alteredAddrSender[0])

	_, ok = altered.Get("7265636569766572")
	require.False(t, ok)

	require.Equal(t, "my-token", scrs[encodedHash].EsdtTokenIdentifier)
}

func TestProcessLogsAndEventsESDT_CrossShardOnDestination(t *testing.T) {
	t.Parallel()

	senderAddr := []byte("addr")
	receiverAddr := []byte("receiver")
	fungibleProc := newFungibleESDTProcessor(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{
		ComputeIdCalled: func(address []byte) uint32 {
			if bytes.Equal(address, senderAddr) {
				return 1
			}
			return 0
		},
	})

	logsAndEvents := map[string]nodeData.LogHandler{
		"txHash": &transaction.Log{
			Events: []*transaction.Event{
				{
					Address:    senderAddr,
					Identifier: []byte(core.BuiltInFunctionESDTTransfer),
					Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).SetUint64(100).Bytes(), receiverAddr},
				},
			},
		},
	}

	altered := data.NewAlteredAccounts()

	txs := make(map[string]*data.Transaction)
	scrs := make(map[string]*data.ScResult)
	encodedHash := hex.EncodeToString([]byte("txHash"))
	scrs[encodedHash] = &data.ScResult{}
	fungibleProc.processLogsAndEventsESDT(logsAndEvents, altered, txs, scrs)

	alteredAddrSender, ok := altered.Get("7265636569766572")
	require.True(t, ok)
	require.Equal(t, &data.AlteredAccount{
		IsESDTOperation: true,
		TokenIdentifier: "my-token",
	}, alteredAddrSender[0])

	_, ok = altered.Get("61646472")
	require.False(t, ok)

	require.Equal(t, "my-token", scrs[encodedHash].EsdtTokenIdentifier)
}
