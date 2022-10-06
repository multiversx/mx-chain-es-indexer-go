package logsevents

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestProcessLogsAndEventsESDT_IntraShard(t *testing.T) {
	t.Parallel()

	fungibleProc := newFungibleESDTProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.BuiltInFunctionESDTTransfer),
		Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).Bytes(), big.NewInt(0).SetUint64(100).Bytes(), []byte("receiver")},
	}
	altered := data.NewAlteredAccounts()

	fungibleProc.processEvent(&argsProcessEvent{
		event:       event,
		accounts:    altered,
		selfShardID: 2,
		numOfShards: 3,
	})

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
}

func TestProcessLogsAndEventsESDT_CrossShardOnSource(t *testing.T) {
	t.Parallel()

	receiverAddr := []byte("a")
	fungibleProc := newFungibleESDTProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(core.BuiltInFunctionESDTTransfer),
		Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).Bytes(), big.NewInt(0).SetUint64(100).Bytes(), receiverAddr},
	}

	altered := data.NewAlteredAccounts()

	res := fungibleProc.processEvent(&argsProcessEvent{
		event:       event,
		accounts:    altered,
		selfShardID: 2,
		numOfShards: 3,
	})
	require.Equal(t, "my-token", res.identifier)
	require.Equal(t, "100", res.value)
	require.Equal(t, true, res.processed)

	alteredAddrSender, ok := altered.Get("61646472")
	require.True(t, ok)
	require.Equal(t, &data.AlteredAccount{
		IsESDTOperation: true,
		TokenIdentifier: "my-token",
	}, alteredAddrSender[0])

	_, ok = altered.Get("61")
	require.False(t, ok)
}

func TestProcessLogsAndEventsESDT_CrossShardOnDestination(t *testing.T) {
	t.Parallel()

	senderAddr := []byte("a")
	receiverAddr := []byte("receiver")
	fungibleProc := newFungibleESDTProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    senderAddr,
		Identifier: []byte(core.BuiltInFunctionESDTTransfer),
		Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).Bytes(), big.NewInt(0).SetUint64(100).Bytes(), receiverAddr},
	}

	altered := data.NewAlteredAccounts()

	res := fungibleProc.processEvent(&argsProcessEvent{
		event:       event,
		accounts:    altered,
		selfShardID: 2,
		numOfShards: 3,
	})
	require.Equal(t, "my-token", res.identifier)
	require.Equal(t, "100", res.value)
	require.Equal(t, true, res.processed)

	alteredAddrSender, ok := altered.Get("7265636569766572")
	require.True(t, ok)
	require.Equal(t, &data.AlteredAccount{
		IsESDTOperation: true,
		TokenIdentifier: "my-token",
	}, alteredAddrSender[0])

	_, ok = altered.Get("61")
	require.False(t, ok)
}
