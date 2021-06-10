package transactions

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
	logsProc := newLogsAndEventsProcessorNFT(&mock.ShardCoordinatorMock{}, &mock.PubkeyConverterMock{}, &mock.MarshalizerMock{})
	logsProc.setLogsAndEventsHandler(&mock.TxLogProcessorStub{
		GetLogFromCacheCalled: func(txHash []byte) (nodeData.LogHandler, bool) {
			switch string(txHash) {
			case "my-hash":
				return &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).SetUint64(nonce).Bytes()},
						},
					},
				}, true
			default:
				return nil, false
			}
		},
	})

	tx1 := &data.Transaction{
		Hash: "6d792d68617368",
	}
	txs := []*data.Transaction{
		tx1,
		{},
	}

	altered := data.NewAlteredAccounts()

	logsProc.processLogsTransactions(txs, altered)
	require.Equal(t, tx1.EsdtTokenIdentifier, "my-token")

	alteredAddr, ok := altered.Get("61646472")
	require.True(t, ok)
	require.Equal(t, &data.AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: "my-token",
		NFTNonce:        19,
		IsCreate:        true,
	}, alteredAddr[0])
}

func TestLogsAndEventsProcessor_TransferNFT(t *testing.T) {
	t.Parallel()

	nonce := uint64(19)
	logsProc := newLogsAndEventsProcessorNFT(&mock.ShardCoordinatorMock{}, &mock.PubkeyConverterMock{}, &mock.MarshalizerMock{})
	logsProc.setLogsAndEventsHandler(&mock.TxLogProcessorStub{
		GetLogFromCacheCalled: func(txHash []byte) (nodeData.LogHandler, bool) {
			switch string(txHash) {
			case "my-hash":
				return &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionESDTNFTTransfer),
							Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).SetUint64(nonce).Bytes(), []byte("receiver")},
						},
					},
				}, true
			default:
				return nil, false
			}
		},
	})

	scr1 := &data.ScResult{
		Hash: "6d792d68617368",
	}
	scrs := []*data.ScResult{
		scr1,
		{},
	}

	altered := data.NewAlteredAccounts()

	logsProc.processLogsScrs(scrs, altered)
	require.Equal(t, scr1.EsdtTokenIdentifier, "my-token")

	alteredAddrSender, ok := altered.Get("61646472")
	require.True(t, ok)
	require.Equal(t, &data.AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: "my-token",
		NFTNonce:        19,
		IsCreate:        true,
	}, alteredAddrSender[0])

	alteredAddrReceiver, ok := altered.Get("7265636569766572")
	require.True(t, ok)
	require.Equal(t, &data.AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: "my-token",
		NFTNonce:        19,
		IsCreate:        false,
	}, alteredAddrReceiver[0])
}
