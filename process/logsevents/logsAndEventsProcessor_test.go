package logsevents

import (
	"math/big"
	"testing"
	"time"

	elasticIndexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go-core/core"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestNewLogsAndEventsProcessor(t *testing.T) {
	t.Parallel()

	balanceConverter, _ := converters.NewBalanceConverter(10)
	_, err := NewLogsAndEventsProcessor(nil, &mock.PubkeyConverterMock{}, &mock.MarshalizerMock{}, balanceConverter, &mock.HasherMock{}, &mock.EconomicsHandlerStub{})
	require.Equal(t, elasticIndexer.ErrNilShardCoordinator, err)

	_, err = NewLogsAndEventsProcessor(&mock.ShardCoordinatorMock{}, nil, &mock.MarshalizerMock{}, balanceConverter, &mock.HasherMock{}, &mock.EconomicsHandlerStub{})
	require.Equal(t, elasticIndexer.ErrNilPubkeyConverter, err)

	_, err = NewLogsAndEventsProcessor(&mock.ShardCoordinatorMock{}, &mock.PubkeyConverterMock{}, nil, balanceConverter, &mock.HasherMock{}, &mock.EconomicsHandlerStub{})
	require.Equal(t, elasticIndexer.ErrNilMarshalizer, err)

	_, err = NewLogsAndEventsProcessor(&mock.ShardCoordinatorMock{}, &mock.PubkeyConverterMock{}, &mock.MarshalizerMock{}, nil, &mock.HasherMock{}, &mock.EconomicsHandlerStub{})
	require.Equal(t, elasticIndexer.ErrNilBalanceConverter, err)

	_, err = NewLogsAndEventsProcessor(&mock.ShardCoordinatorMock{}, &mock.PubkeyConverterMock{}, &mock.MarshalizerMock{}, balanceConverter, nil, &mock.EconomicsHandlerStub{})
	require.Equal(t, elasticIndexer.ErrNilHasher, err)

	_, err = NewLogsAndEventsProcessor(&mock.ShardCoordinatorMock{}, &mock.PubkeyConverterMock{}, &mock.MarshalizerMock{}, balanceConverter, &mock.HasherMock{}, nil)
	require.Equal(t, elasticIndexer.ErrNilTransactionFeeCalculator, err)

	proc, err := NewLogsAndEventsProcessor(&mock.ShardCoordinatorMock{}, &mock.PubkeyConverterMock{}, &mock.MarshalizerMock{}, balanceConverter, &mock.HasherMock{}, &mock.EconomicsHandlerStub{})
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
					Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).SetUint64(1).Bytes(), big.NewInt(100).Bytes(), []byte("receiver")},
				},
			},
		},

		"h2": &transaction.Log{
			Events: []*transaction.Event{
				{
					Address:    []byte("addr"),
					Identifier: []byte(core.BuiltInFunctionESDTTransfer),
					Topics:     [][]byte{[]byte("esdt"), big.NewInt(0).Bytes(), big.NewInt(0).SetUint64(100).Bytes(), []byte("receiver")},
				},
				nil,
			},
		},
		"h4": &transaction.Log{
			Events: []*transaction.Event{
				{
					Address:    []byte("addr"),
					Identifier: []byte(issueSemiFungibleESDTFunc),
					Topics:     [][]byte{[]byte("SEMI-abcd"), []byte("semi-token"), []byte("SEMI"), []byte(core.SemiFungibleESDT)},
				},
				nil,
			},
		},
		"h5": &transaction.Log{
			Address: []byte("contract"),
			Events: []*transaction.Event{
				{
					Address:    []byte("addr"),
					Identifier: []byte(delegateFunc),
					Topics:     [][]byte{big.NewInt(1000).Bytes(), big.NewInt(1000000000).Bytes(), big.NewInt(10).Bytes(), big.NewInt(1000000000).Bytes()},
				},
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

	balanceConverter, _ := converters.NewBalanceConverter(10)
	proc, _ := NewLogsAndEventsProcessor(&mock.ShardCoordinatorMock{
		SelfID: core.MetachainShardId,
	}, mock.NewPubkeyConverterMock(32), &mock.MarshalizerMock{}, balanceConverter, &mock.HasherMock{}, &mock.EconomicsHandlerStub{})

	resLogs := proc.ExtractDataFromLogs(logsAndEvents, res, 1000)
	require.NotNil(t, resLogs.Tokens)
	require.NotNil(t, resLogs.TagsCount)
	require.True(t, res.Transactions[0].HasOperations)
	require.True(t, res.ScResults[0].HasOperations)

	require.Equal(t, &data.ScDeployInfo{
		TxHash:    "6833",
		Creator:   "6164647232",
		Timestamp: uint64(1000),
	}, resLogs.ScDeploys["6164647231"])

	require.Equal(t, &data.TokenInfo{
		Name:         "semi-token",
		Ticker:       "SEMI",
		Token:        "SEMI-abcd",
		Type:         core.SemiFungibleESDT,
		Timestamp:    1000,
		Issuer:       "61646472",
		CurrentOwner: "61646472",
		OwnersHistory: []*data.OwnerData{
			{
				Address:   "61646472",
				Timestamp: 1000,
			},
		},
	}, resLogs.TokensInfo[0])

	require.Equal(t, &data.Delegator{
		Address:        "61646472",
		Contract:       "636f6e7472616374",
		ActiveStakeNum: 0.1,
		ActiveStake:    "1000000000",
	}, resLogs.Delegators["61646472"])
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

	balanceConverter, _ := converters.NewBalanceConverter(10)
	proc, _ := NewLogsAndEventsProcessor(&mock.ShardCoordinatorMock{}, mock.NewPubkeyConverterMock(32), &mock.MarshalizerMock{}, balanceConverter, &mock.HasherMock{}, &mock.EconomicsHandlerStub{})

	logsDB := proc.PrepareLogsForDB(logsAndEvents, 1234)
	require.Equal(t, &data.Logs{
		ID:        "747848617368",
		Address:   "61646472657373",
		Timestamp: time.Duration(1234),
		Events: []*data.Event{
			{
				Address:    "61646472",
				Identifier: core.BuiltInFunctionESDTNFTTransfer,
				Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).SetUint64(1).Bytes(), []byte("receiver")},
			},
		},
	}, logsDB[0])
}
