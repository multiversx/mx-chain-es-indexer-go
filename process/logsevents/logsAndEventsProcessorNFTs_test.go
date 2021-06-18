package logsevents

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go/core"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestLogsAndEventsProcessor(t *testing.T) {
	t.Parallel()

	nonce := uint64(19)
	logsAndEvents := map[string]nodeData.LogHandler{
		"txHash": &transaction.Log{
			Events: []*transaction.Event{
				{
					Address:    []byte("addr"),
					Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
					Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).SetUint64(nonce).Bytes(), []byte(core.NonFungibleESDT)},
				},
			},
		},
	}

	logsProc, _ := NewLogsAndEventsProcessorNFT(&mock.ShardCoordinatorMock{}, &mock.PubkeyConverterMock{})

	altered := data.NewAlteredAccounts()

	logsProc.ProcessLogsAndEvents(logsAndEvents, altered)

	alteredAddr, ok := altered.Get("61646472")
	require.True(t, ok)
	require.Equal(t, &data.AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: "my-token",
		NFTNonce:        19,
		IsNFTCreate:     true,
		Type:            core.NonFungibleESDT,
	}, alteredAddr[0])
}

func TestLogsAndEventsProcessor_TransferNFT(t *testing.T) {
	t.Parallel()

	nonce := uint64(19)
	logsProc, _ := NewLogsAndEventsProcessorNFT(&mock.ShardCoordinatorMock{}, &mock.PubkeyConverterMock{})

	logsAndEvents := map[string]nodeData.LogHandler{
		"txHash": &transaction.Log{
			Events: []*transaction.Event{
				{
					Address:    []byte("addr"),
					Identifier: []byte(core.BuiltInFunctionESDTNFTTransfer),
					Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).SetUint64(nonce).Bytes(), []byte("receiver")},
				},
			},
		},
	}

	altered := data.NewAlteredAccounts()

	logsProc.ProcessLogsAndEvents(logsAndEvents, altered)

	alteredAddrSender, ok := altered.Get("61646472")
	require.True(t, ok)
	require.Equal(t, &data.AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: "my-token",
		NFTNonce:        19,
		IsNFTCreate:     false,
	}, alteredAddrSender[0])

	alteredAddrReceiver, ok := altered.Get("7265636569766572")
	require.True(t, ok)
	require.Equal(t, &data.AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: "my-token",
		NFTNonce:        19,
		IsNFTCreate:     false,
	}, alteredAddrReceiver[0])
}
