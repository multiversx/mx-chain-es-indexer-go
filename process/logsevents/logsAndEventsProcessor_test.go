package logsevents

import (
	"math/big"
	"testing"

	elasticIndexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go-core/core"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestNewLogsAndEventsProcessor(t *testing.T) {
	t.Parallel()

	_, err := NewLogsAndEventsProcessor(nil, &mock.PubkeyConverterMock{}, &mock.MarshalizerMock{})
	require.Equal(t, elasticIndexer.ErrNilShardCoordinator, err)

	_, err = NewLogsAndEventsProcessor(&mock.ShardCoordinatorMock{}, nil, &mock.MarshalizerMock{})
	require.Equal(t, elasticIndexer.ErrNilPubkeyConverter, err)

	_, err = NewLogsAndEventsProcessor(&mock.ShardCoordinatorMock{}, &mock.PubkeyConverterMock{}, nil)
	require.Equal(t, elasticIndexer.ErrNilMarshalizer, err)

	proc, err := NewLogsAndEventsProcessor(&mock.ShardCoordinatorMock{}, &mock.PubkeyConverterMock{}, &mock.MarshalizerMock{})
	require.NotNil(t, proc)
	require.Nil(t, err)
}

func TestLogsAndEventsProcessor_ExtractDataFromLogsAndPutInAltered(t *testing.T) {
	t.Parallel()

	logsAndEvents := map[string]coreData.LogHandler{
		"wrong": nil,
		"h3": &transaction.Log{
			Events: []*transaction.Event{
				{
					Address:    []byte("addr"),
					Identifier: []byte(core.SCDeployIdentifier),
					Topics:     [][]byte{[]byte("addr1"), []byte("addr2")},
				},
			},
		},

		"h1": &transaction.Log{
			Address: []byte("address"),
			Events: []*transaction.Event{
				{
					Address:    []byte("addr"),
					Identifier: []byte(core.BuiltInFunctionESDTNFTTransfer),
					Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).SetUint64(1).Bytes(), []byte("receiver")},
				},
			},
		},

		"h2": &transaction.Log{
			Events: []*transaction.Event{
				{
					Address:    []byte("addr"),
					Identifier: []byte(core.BuiltInFunctionESDTTransfer),
					Topics:     [][]byte{[]byte("esdt"), big.NewInt(0).SetUint64(100).Bytes(), []byte("receiver")},
				},
				nil,
			},
		},
	}

	altered := data.NewAlteredAccounts()
	res := &data.PreparedResults{
		Transactions: []*data.Transaction{
			{
				Hash: "6831",
			},
		},
		ScResults: []*data.ScResult{
			{
				Hash: "6832",
			},
		},
		AlteredAccts: altered,
	}
	proc, _ := NewLogsAndEventsProcessor(&mock.ShardCoordinatorMock{}, mock.NewPubkeyConverterMock(32), &mock.MarshalizerMock{})

	tokens, tagsCount, scDeploys := proc.ExtractDataFromLogsAndPutInAltered(logsAndEvents, res, 1000)
	require.NotNil(t, tokens)
	require.NotNil(t, tagsCount)
	require.Equal(t, "my-token-01", res.Transactions[0].EsdtTokenIdentifier)
	require.Equal(t, "esdt", res.ScResults[0].EsdtTokenIdentifier)

	require.Equal(t, &data.ScDeployInfo{
		TxHash:    "6833",
		Creator:   "6164647232",
		Timestamp: uint64(1000),
	}, scDeploys["6164647231"])
}

func TestLogsAndEventsProcessor_PrepareLogsForDB(t *testing.T) {
	t.Parallel()

	logsAndEvents := map[string]coreData.LogHandler{
		"wrong": nil,

		"txHash": &transaction.Log{
			Address: []byte("address"),
			Events: []*transaction.Event{
				{
					Address:    []byte("addr"),
					Identifier: []byte(core.BuiltInFunctionESDTNFTTransfer),
					Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).SetUint64(1).Bytes(), []byte("receiver")},
				},
			},
		},
	}

	proc, _ := NewLogsAndEventsProcessor(&mock.ShardCoordinatorMock{}, mock.NewPubkeyConverterMock(32), &mock.MarshalizerMock{})

	logsDB := proc.PrepareLogsForDB(logsAndEvents)
	require.Equal(t, &data.Logs{
		ID:      "747848617368",
		Address: "61646472657373",
		Events: []*data.Event{
			{
				Address:    "61646472",
				Identifier: core.BuiltInFunctionESDTNFTTransfer,
				Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).SetUint64(1).Bytes(), []byte("receiver")},
			},
		},
	}, logsDB[0])
}
