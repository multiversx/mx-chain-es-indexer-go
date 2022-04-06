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

func createMockArgs() *ArgsLogsAndEventsProcessor {
	balanceConverter, _ := converters.NewBalanceConverter(10)
	return &ArgsLogsAndEventsProcessor{
		ShardCoordinator: &mock.ShardCoordinatorMock{},
		PubKeyConverter:  &mock.PubkeyConverterMock{},
		Marshalizer:      &mock.MarshalizerMock{},
		BalanceConverter: balanceConverter,
		Hasher:           &mock.HasherMock{},
		TxFeeCalculator:  &mock.EconomicsHandlerStub{},
	}
}

func TestNewLogsAndEventsProcessor(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	args.ShardCoordinator = nil
	_, err := NewLogsAndEventsProcessor(args)
	require.Equal(t, elasticIndexer.ErrNilShardCoordinator, err)

	args = createMockArgs()
	args.PubKeyConverter = nil
	_, err = NewLogsAndEventsProcessor(args)
	require.Equal(t, elasticIndexer.ErrNilPubkeyConverter, err)

	args = createMockArgs()
	args.Marshalizer = nil
	_, err = NewLogsAndEventsProcessor(args)
	require.Equal(t, elasticIndexer.ErrNilMarshalizer, err)

	args = createMockArgs()
	args.BalanceConverter = nil
	_, err = NewLogsAndEventsProcessor(args)
	require.Equal(t, elasticIndexer.ErrNilBalanceConverter, err)

	args = createMockArgs()
	args.Hasher = nil
	_, err = NewLogsAndEventsProcessor(args)
	require.Equal(t, elasticIndexer.ErrNilHasher, err)

	args = createMockArgs()
	args.TxFeeCalculator = nil
	_, err = NewLogsAndEventsProcessor(args)
	require.Equal(t, elasticIndexer.ErrNilTransactionFeeCalculator, err)

	args = createMockArgs()
	proc, err := NewLogsAndEventsProcessor(args)
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
		"h6": &transaction.Log{
			Address: []byte("contract-second"),
			Events: []*transaction.Event{
				{
					Address:    []byte("addr"),
					Identifier: []byte(delegateFunc),
					Topics:     [][]byte{big.NewInt(1000).Bytes(), big.NewInt(1000000000).Bytes(), big.NewInt(10).Bytes(), big.NewInt(1000000000).Bytes()},
				},
			},
		},
	}

	logsAndEventsSlice := make([]*coreData.LogData, 0)
	for hash, val := range logsAndEvents {
		logsAndEventsSlice = append(logsAndEventsSlice, &coreData.LogData{
			TxHash:     hash,
			LogHandler: val,
		})
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

	args := createMockArgs()
	balanceConverter, _ := converters.NewBalanceConverter(10)
	args.BalanceConverter = balanceConverter
	args.ShardCoordinator = &mock.ShardCoordinatorMock{
		SelfID: core.MetachainShardId,
	}
	proc, _ := NewLogsAndEventsProcessor(args)

	resLogs := proc.ExtractDataFromLogs(logsAndEventsSlice, res, 1000)
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
	}, resLogs.Delegators["61646472636f6e7472616374"])
	require.Equal(t, &data.Delegator{
		Address:        "61646472",
		Contract:       "636f6e74726163742d7365636f6e64",
		ActiveStakeNum: 0.1,
		ActiveStake:    "1000000000",
	}, resLogs.Delegators["61646472636f6e74726163742d7365636f6e64"])
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

	logsAndEventsSlice := make([]*coreData.LogData, 0)
	for hash, val := range logsAndEvents {
		logsAndEventsSlice = append(logsAndEventsSlice, &coreData.LogData{
			TxHash:     hash,
			LogHandler: val,
		})
	}

	args := createMockArgs()
	proc, _ := NewLogsAndEventsProcessor(args)

	_ = proc.ExtractDataFromLogs(nil, &data.PreparedResults{ScResults: []*data.ScResult{
		{
			Hash:           "747848617368",
			OriginalTxHash: "orignalHash",
		},
	}}, 1234)

	logsDB := proc.PrepareLogsForDB(logsAndEventsSlice, 1234)
	require.Equal(t, &data.Logs{
		ID:             "747848617368",
		Address:        "61646472657373",
		OriginalTxHash: "orignalHash",
		Timestamp:      time.Duration(1234),
		Events: []*data.Event{
			{
				Address:    "61646472",
				Identifier: core.BuiltInFunctionESDTNFTTransfer,
				Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).SetUint64(1).Bytes(), []byte("receiver")},
			},
		},
	}, logsDB[0])
}

func TestLogsAndEventsProcessor_ExtractDataFromLogsNFTBurn(t *testing.T) {
	t.Parallel()

	logsAndEventsSlice := make([]*coreData.LogData, 1)
	logsAndEventsSlice[0] = &coreData.LogData{
		LogHandler: &transaction.Log{
			Address: []byte("address"),
			Events: []*transaction.Event{
				{
					Address:    []byte("addr"),
					Identifier: []byte(core.BuiltInFunctionESDTNFTBurn),
					Topics:     [][]byte{[]byte("MY-NFT"), big.NewInt(2).Bytes(), big.NewInt(1).Bytes()},
				},
			},
		},
		TxHash: "h1",
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

	args := createMockArgs()
	balanceConverter, _ := converters.NewBalanceConverter(10)
	args.BalanceConverter = balanceConverter
	args.ShardCoordinator = &mock.ShardCoordinatorMock{
		SelfID: 0,
	}
	proc, _ := NewLogsAndEventsProcessor(args)

	resLogs := proc.ExtractDataFromLogs(logsAndEventsSlice, res, 1000)
	require.Equal(t, 1, resLogs.TokensSupply.Len())

	tokensSupply := resLogs.TokensSupply.GetAll()
	require.Equal(t, "MY-NFT", tokensSupply[0].Token)
	require.Equal(t, "MY-NFT-02", tokensSupply[0].Identifier)
}
