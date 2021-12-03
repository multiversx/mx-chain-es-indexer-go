//go:build integration

package integrationtests

import (
	"encoding/hex"
	"math/big"
	"testing"

	indexerData "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go-core/data/receipt"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

const (
	moveBalanceTransaction     = `{"miniBlockHash":"24c374c9405540e88a36959ea83eede6ad50f6872f82d2e2a2280975615e1811","nonce":1,"round":50,"value":"1234","receiver":"7265636569766572","sender":"73656e646572","receiverShard":0,"senderShard":0,"gasPrice":1000000000,"gasLimit":70000,"gasUsed":62000,"fee":"62000000000000","data":"dHJhbnNmZXI=","signature":"","timestamp":5040,"status":"success","searchOrder":0,"receipt":{"value":"1000","data":"gasRefund"}}`
	transactionSCRS            = `{"miniBlockHash":"","nonce":0,"round":0,"value":"","receiver":"","sender":"","receiverShard":0,"senderShard":0,"gasPrice":0,"gasLimit":0,"gasUsed":0,"fee":"","data":null,"signature":"","timestamp":0,"status":"","searchOrder":0,"scresults":{"736f6c6f2d7363722d31":{"miniBlockHash":"b23fbb9a009a84d2a70ef8bfd329af3c0def00b484eadb3cf509aab4f2d7d3a7","nonce":1,"gasLimit":70000,"gasPrice":1000000000,"value":"1234","sender":"73656e646572","receiver":"7265636569766572","senderShard":2,"receiverShard":0,"data":"dHJhbnNmZXI=","prevTxHash":"74782d68617368","originalTxHash":"74782d68617368","callType":"0","timestamp":5040},"736f6c6f2d7363722d32":{"prevTxHash":"74782d68617368","receiver":"7265636569766572","data":"ZG8tc29tZXRoaW5n","nonce":12,"callType":"0","gasLimit":0,"originalTxHash":"74782d68617368","miniBlockHash":"b23fbb9a009a84d2a70ef8bfd329af3c0def00b484eadb3cf509aab4f2d7d3a7","sender":"73656e646572","receiverShard":0,"senderShard":2,"value":"0","gasPrice":0,"timestamp":5040}}}`
	transactionWithDataAndSCRS = `{"receiver":"7265636569766572","data":"ZG8tY2FsbA==","signature":"7369676e6174757265","fee":"60595000000","nonce":12,"gasLimit":70000,"gasUsed":70000,"miniBlockHash":"544b737692c0d07c4adc0404ed34984dfbf2c89e6f44957d1ea67bf6883df0a3","round":50,"hasScResults":true,"sender":"73656e646572","receiverShard":0,"senderShard":0,"scresults":{"736f6c6f2d7363722d31":{"prevTxHash":"74782d68617368","receiver":"7265636569766572","data":"dHJhbnNmZXI=","nonce":1,"callType":"0","gasLimit":70000,"originalTxHash":"74782d68617368","miniBlockHash":"b23fbb9a009a84d2a70ef8bfd329af3c0def00b484eadb3cf509aab4f2d7d3a7","sender":"73656e646572","receiverShard":0,"senderShard":2,"value":"1234","gasPrice":1000000000,"timestamp":5040},"736f6c6f2d7363722d32":{"prevTxHash":"74782d68617368","receiver":"7265636569766572","data":"ZG8tc29tZXRoaW5n","nonce":12,"callType":"0","gasLimit":0,"originalTxHash":"74782d68617368","miniBlockHash":"b23fbb9a009a84d2a70ef8bfd329af3c0def00b484eadb3cf509aab4f2d7d3a7","sender":"73656e646572","receiverShard":0,"senderShard":2,"value":"0","gasPrice":0,"timestamp":5040},"7363722d33":{"prevTxHash":"74782d68617368","receiver":"7265636569766572","data":"ZG8tdGhpbmdz","nonce":111,"callType":"0","gasLimit":0,"originalTxHash":"74782d68617368","miniBlockHash":"1934d6538d33c94476a45083799384ccbd92c5a70e9971aeefc00f029b038844","sender":"73656e646572","receiverShard":0,"senderShard":0,"value":"0","gasPrice":0,"timestamp":5040}},"value":"0","gasPrice":1000000,"timestamp":5040,"status":"fail","searchOrder":0}`
)

func TestElasticIndexerSaveTransactions(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	txHash := []byte("hash")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}
	body := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.TxBlock,
				SenderShardID:   0,
				ReceiverShardID: 0,
				TxHashes:        [][]byte{txHash},
			},
		},
	}
	pool := &indexer.Pool{
		Txs: map[string]coreData.TransactionHandler{
			string(txHash): &transaction.Transaction{
				Nonce:    1,
				SndAddr:  []byte("sender"),
				RcvAddr:  []byte("receiver"),
				GasLimit: 70000,
				GasPrice: 1000000000,
				Data:     []byte("transfer"),
				Value:    big.NewInt(1234),
			},
		},
		Receipts: map[string]coreData.TransactionHandler{
			string("recHash"): &receipt.Receipt{
				SndAddr: []byte("sender"),
				Value:   big.NewInt(1000),
				Data:    []byte("gasRefund"),
				TxHash:  txHash,
			},
		},
	}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerData.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, []byte(moveBalanceTransaction), genericResponse.Docs[0].Source)
}

func TestTransactionCrossShardIndexSCRFirstThanTransaction(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	txHash := []byte("tx-hash")
	scrHash1 := []byte("solo-scr-1")
	scrHash2 := []byte("solo-scr-2")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}
	body := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   2,
				ReceiverShardID: 0,
				TxHashes:        [][]byte{scrHash1, scrHash2},
			},
		},
	}
	pool := &indexer.Pool{
		Scrs: map[string]coreData.TransactionHandler{
			string(scrHash1): &smartContractResult.SmartContractResult{
				Nonce:          1,
				SndAddr:        []byte("sender"),
				RcvAddr:        []byte("receiver"),
				GasLimit:       70000,
				GasPrice:       1000000000,
				Data:           []byte("transfer"),
				Value:          big.NewInt(1234),
				OriginalTxHash: txHash,
				PrevTxHash:     txHash,
			},
			string(scrHash2): &smartContractResult.SmartContractResult{
				Nonce:          12,
				SndAddr:        []byte("sender"),
				RcvAddr:        []byte("receiver"),
				Data:           []byte("do-something"),
				Value:          big.NewInt(0),
				OriginalTxHash: txHash,
				PrevTxHash:     txHash,
			},
		},
	}

	// index on shard 2
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerData.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, []byte(transactionSCRS), genericResponse.Docs[0].Source)

	// index transaction on shard 0
	scrHash3 := []byte("scr-3")
	pool = &indexer.Pool{
		Scrs: map[string]coreData.TransactionHandler{
			string(scrHash3): &smartContractResult.SmartContractResult{
				Nonce:          111,
				SndAddr:        []byte("sender"),
				RcvAddr:        []byte("receiver"),
				Data:           []byte("do-things"),
				Value:          big.NewInt(0),
				OriginalTxHash: txHash,
				PrevTxHash:     txHash,
			},
		},
		Txs: map[string]coreData.TransactionHandler{
			string(txHash): &transaction.Transaction{
				Nonce:     12,
				SndAddr:   []byte("sender"),
				RcvAddr:   []byte("receiver"),
				Data:      []byte("do-call"),
				Value:     big.NewInt(0),
				Signature: []byte("signature"),
				GasPrice:  1000000,
				GasLimit:  70000,
			},
		},
	}

	body = &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:     dataBlock.SmartContractResultBlock,
				TxHashes: [][]byte{scrHash3},
			},
			{
				Type:     dataBlock.TxBlock,
				TxHashes: [][]byte{txHash},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexerData.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, []byte(transactionWithDataAndSCRS), genericResponse.Docs[0].Source)
}
