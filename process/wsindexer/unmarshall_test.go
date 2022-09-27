package wsindexer

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/process/dataindexer"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/receipt"
	"github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal/factory"
	outportData "github.com/ElrondNetwork/elrond-go-core/websocketOutportDriver/data"
	"github.com/stretchr/testify/require"
)

func TestGetHeaderAndBody(t *testing.T) {
	t.Parallel()

	marshaller, _ := factory.NewMarshalizer("json")
	nilDataIndexer := dataindexer.NewNilIndexer()

	di, _ := NewIndexer(marshaller, nilDataIndexer)

	arg := &outportData.ArgsRevertIndexedBlock{
		HeaderType: outportData.ShardHeaderV2,
		Header:     &block.HeaderV2{ScheduledRootHash: []byte("aaaaaa")},
		Body:       &block.Body{MiniBlocks: []*block.MiniBlock{{}}},
	}
	argBytes, _ := marshaller.Marshal(arg)

	body, header, err := di.getHeaderAndBody(argBytes)
	require.Nil(t, err)
	require.NotNil(t, body)
	require.NotNil(t, header)
}

func TestGetPool(t *testing.T) {
	t.Parallel()

	marshaller, _ := factory.NewMarshalizer("json")
	nilDataIndexer := dataindexer.NewNilIndexer()

	di, _ := NewIndexer(marshaller, nilDataIndexer)

	pool := &outport.Pool{
		Txs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			"txHash": outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
				Nonce: 1,
			}, 1, big.NewInt(100)),
		},
		Scrs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			"scrHash": outport.NewTransactionHandlerWithGasAndFee(&smartContractResult.SmartContractResult{
				Nonce: 2,
			}, 0, big.NewInt(0)),
		},
		Rewards: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			"reward": outport.NewTransactionHandlerWithGasAndFee(&rewardTx.RewardTx{
				Value: big.NewInt(10),
			}, 0, big.NewInt(0)),
		},
		Invalid: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			"invalid": outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
				Nonce: 3,
			}, 100, big.NewInt(1000)),
		},
		Receipts: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			"rec": outport.NewTransactionHandlerWithGasAndFee(&receipt.Receipt{
				Value: big.NewInt(300),
			}, 0, big.NewInt(0)),
		},
		Logs: []*coreData.LogData{
			{
				TxHash: "something",
				LogHandler: &transaction.Log{
					Address: []byte("addr"),
				},
			},
		},
	}

	argsSaveBlock := &outportData.ArgsSaveBlock{
		ArgsSaveBlockData: &outport.ArgsSaveBlockData{
			TransactionsPool: pool,
		},
	}

	argsSaveBlockBytes, _ := di.marshaller.Marshal(argsSaveBlock)

	resPool, err := di.getTxsPool(argsSaveBlockBytes)
	require.Nil(t, err)
	require.NotNil(t, resPool)
	require.Equal(t, 1, len(resPool.Txs))
	require.Equal(t, 1, len(resPool.Scrs))
	require.Equal(t, 1, len(resPool.Rewards))
	require.Equal(t, 1, len(resPool.Receipts))
	require.Equal(t, 1, len(resPool.Invalid))
	require.Equal(t, 1, len(resPool.Logs))
}
