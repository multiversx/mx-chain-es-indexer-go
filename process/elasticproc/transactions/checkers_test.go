package transactions

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	elasticIndexer "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
)

func createMockArgs() *ArgsTransactionProcessor {
	bc, _ := converters.NewBalanceConverter(18)
	return &ArgsTransactionProcessor{
		AddressPubkeyConverter: &mock.PubkeyConverterMock{},
		Hasher:                 &mock.HasherMock{},
		Marshalizer:            &mock.MarshalizerMock{},
		BalanceConverter:       bc,
		TxHashExtractor:        &mock.TxHashExtractorMock{},
		RewardTxData:           &mock.RewardTxDataMock{},
	}
}

func TestNewTransactionsProcessor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		args  func() *ArgsTransactionProcessor
		exErr error
	}{
		{
			name: "NilAddressPubkeyConvertor",
			args: func() *ArgsTransactionProcessor {
				args := createMockArgs()
				args.AddressPubkeyConverter = nil
				return args
			},
			exErr: elasticIndexer.ErrNilPubkeyConverter,
		},
		{
			name: "NilMarshalizer",
			args: func() *ArgsTransactionProcessor {
				args := createMockArgs()
				args.Marshalizer = nil
				return args
			},
			exErr: elasticIndexer.ErrNilMarshalizer,
		},
		{
			name: "NilHasher",
			args: func() *ArgsTransactionProcessor {
				args := createMockArgs()
				args.Hasher = nil
				return args
			},
			exErr: elasticIndexer.ErrNilHasher,
		},
		{
			name: "NilBalanceConverter",
			args: func() *ArgsTransactionProcessor {
				args := createMockArgs()
				args.BalanceConverter = nil
				return args
			},
			exErr: elasticIndexer.ErrNilBalanceConverter,
		},
		{
			name: "NilTxHashExtractor",
			args: func() *ArgsTransactionProcessor {
				args := createMockArgs()
				args.TxHashExtractor = nil
				return args
			},
			exErr: ErrNilTxHashExtractor,
		},
		{
			name: "NilRewardTxDataHandler",
			args: func() *ArgsTransactionProcessor {
				args := createMockArgs()
				args.RewardTxData = nil
				return args
			},
			exErr: ErrNilRewardTxDataHandler,
		},
	}

	for _, tt := range tests {
		_, err := NewTransactionsProcessor(tt.args())
		require.Equal(t, tt.exErr, err)
	}
}

func TestCheckTxsProcessorArg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		args  func() (header coreData.HeaderHandler, pool *outport.TransactionPool)
		exErr error
	}{
		{
			name: "NilHeaderHandler",
			args: func() (header coreData.HeaderHandler, pool *outport.TransactionPool) {
				return nil, &outport.TransactionPool{}
			},
			exErr: elasticIndexer.ErrNilHeaderHandler,
		},
		{
			name: "NilPool",
			args: func() (header coreData.HeaderHandler, pool *outport.TransactionPool) {
				return &block.Header{}, nil
			},
			exErr: elasticIndexer.ErrNilPool,
		},
	}

	for _, tt := range tests {
		err := checkPrepareTransactionForDatabaseArguments(tt.args())
		require.Equal(t, tt.exErr, err)
	}
}

func TestIsScResultSuccessful(t *testing.T) {
	t.Parallel()

	scResultData := []byte("@" + vmcommon.Ok.String())
	require.True(t, isScResultSuccessful(scResultData))

	scResultData = []byte("user error")
	require.False(t, isScResultSuccessful(scResultData))

	scResultData = []byte("@" + hex.EncodeToString([]byte(vmcommon.Ok.String())))
	require.True(t, isScResultSuccessful(scResultData))
}

func TestStringValueToBigInt(t *testing.T) {
	t.Parallel()

	str1 := "10000"
	require.Equal(t, big.NewInt(10000), stringValueToBigInt(str1))

	str2 := "aaaa"
	require.Equal(t, big.NewInt(0), stringValueToBigInt(str2))
}

func TestIsRelayedTx(t *testing.T) {
	t.Parallel()

	tx1 := &data.Transaction{
		Data:                 []byte(core.RelayedTransaction + "@aaaaaa"),
		SmartContractResults: []*data.ScResult{{}},
	}

	require.True(t, isRelayedTx(tx1))

	tx2 := &data.Transaction{
		Data:                 []byte(core.RelayedTransaction + "@aaaaaa"),
		SmartContractResults: []*data.ScResult{},
	}

	require.False(t, isRelayedTx(tx2))
}

func TestIsCrossShardSourceMe(t *testing.T) {
	t.Parallel()

	tx1 := &data.Transaction{SenderShard: 2, ReceiverShard: 1}
	require.True(t, isCrossShardOnSourceShard(tx1, 2))

	tx2 := &data.Transaction{SenderShard: 1, ReceiverShard: 1}
	require.False(t, isCrossShardOnSourceShard(tx2, 1))
}

func TestAreESDTValuesOK(t *testing.T) {
	t.Parallel()

	values := []string{"10000", "1", "10"}
	require.True(t, areESDTValuesOK(values))

	values = []string{"10000", "1", "1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"}
	require.False(t, areESDTValuesOK(values))
}
