package transactions

import (
	"encoding/hex"
	"math/big"
	"testing"

	elasticIndexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/indexer"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func createMockArgs() *ArgsTransactionProcessor {
	return &ArgsTransactionProcessor{
		AddressPubkeyConverter: &mock.PubkeyConverterMock{},
		TxFeeCalculator:        &mock.EconomicsHandlerStub{},
		ShardCoordinator:       &mock.ShardCoordinatorMock{},
		Hasher:                 &mock.HasherMock{},
		Marshalizer:            &mock.MarshalizerMock{},
		IsInImportMode:         false,
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
			name: "NilTxFeeCalculator",
			args: func() *ArgsTransactionProcessor {
				args := createMockArgs()
				args.TxFeeCalculator = nil
				return args
			},
			exErr: elasticIndexer.ErrNilTransactionFeeCalculator,
		},
		{
			name: "NilShardCoordinator",
			args: func() *ArgsTransactionProcessor {
				args := createMockArgs()
				args.ShardCoordinator = nil
				return args
			},
			exErr: elasticIndexer.ErrNilShardCoordinator,
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
		args  func() (body *block.Body, header nodeData.HeaderHandler, pool *indexer.Pool)
		exErr error
	}{
		{
			name: "NilBlockBody",
			args: func() (body *block.Body, header nodeData.HeaderHandler, pool *indexer.Pool) {
				return nil, &block.Header{}, &indexer.Pool{}
			},
			exErr: elasticIndexer.ErrNilBlockBody,
		},
		{
			name: "NilHeaderHandler",
			args: func() (body *block.Body, header nodeData.HeaderHandler, pool *indexer.Pool) {
				return &block.Body{}, nil, &indexer.Pool{}
			},
			exErr: elasticIndexer.ErrNilHeaderHandler,
		},
		{
			name: "NilPool",
			args: func() (body *block.Body, header nodeData.HeaderHandler, pool *indexer.Pool) {
				return &block.Body{}, &block.Header{}, nil
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

func TestIsCrossShardDstMe(t *testing.T) {
	t.Parallel()

	tx1 := &data.Transaction{SenderShard: 2, ReceiverShard: 1}
	require.True(t, isCrossShardDstMe(tx1, 1))

	tx2 := &data.Transaction{SenderShard: 1, ReceiverShard: 1}
	require.False(t, isCrossShardDstMe(tx2, 1))
}

func TestIsIntraShardOrInvalid(t *testing.T) {
	t.Parallel()

	tx1 := &data.Transaction{SenderShard: 1, ReceiverShard: 1}
	require.True(t, isIntraShardOrInvalid(tx1, 1))

	tx2 := &data.Transaction{SenderShard: 2, ReceiverShard: 1}
	require.False(t, isIntraShardOrInvalid(tx2, 1))

	tx3 := &data.Transaction{SenderShard: 2, ReceiverShard: 1, Status: transaction.TxStatusInvalid.String()}
	require.True(t, isIntraShardOrInvalid(tx3, 1))
}
