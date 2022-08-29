//go:build integrationtests

package integrationtests

import (
	"encoding/hex"
	"math/big"
	"testing"

	indexerdata "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestESDTTransferTooMuchGasProvided(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, shardCoordinator)
	require.Nil(t, err)

	txHash := []byte("esdtTransfer")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}
	scrHash2 := []byte("scrHash2ESDTTransfer")
	body := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.TxBlock,
				SenderShardID:   0,
				ReceiverShardID: 0,
				TxHashes:        [][]byte{txHash},
			},
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   0,
				ReceiverShardID: 1,
				TxHashes:        [][]byte{scrHash2},
			},
		},
	}

	txESDT := &transaction.Transaction{
		Nonce:    6,
		SndAddr:  []byte("erd1ef6470tjdtlgpa9f6g3ae4nsedmjg0gv6w73v32xtvhkfff993hq750xl9"),
		RcvAddr:  []byte("erd13u7zyekzvdvzek8768r5gau9p6677ufppsjuklu9e6t7yx7rhg4s68e2ze"),
		GasLimit: 104011,
		GasPrice: 1000000000,
		Data:     []byte("ESDTTransfer@54474e2d383862383366@0a"),
		Value:    big.NewInt(0),
	}

	scrHash1 := []byte("scrHash1ESDTTransfer")
	scr1 := &smartContractResult.SmartContractResult{
		Nonce:          7,
		GasPrice:       1000000000,
		SndAddr:        []byte("erd13u7zyekzvdvzek8768r5gau9p6677ufppsjuklu9e6t7yx7rhg4s68e2ze"),
		RcvAddr:        []byte("erd1ef6470tjdtlgpa9f6g3ae4nsedmjg0gv6w73v32xtvhkfff993hq750xl9"),
		Data:           []byte("@6f6b"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
		ReturnMessage:  []byte("@too much gas provided: gas needed = 372000, gas remained = 2250001"),
	}

	scr2 := &smartContractResult.SmartContractResult{
		Nonce:          7,
		GasPrice:       1000000000,
		SndAddr:        []byte("erd13u7zyekzvdvzek8768r5gau9p6677ufppsjuklu9e6t7yx7rhg4s68e2ze"),
		RcvAddr:        []byte("erd1ef6470tjdtlgpa9f6g3ae4nsedmjg0gv6w73v32xtvhkfff993hq750xl9"),
		Data:           []byte("@6f6b"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
	}

	initialPaidFee, _ := big.NewInt(0).SetString("104000110000000", 10)
	tx := outport.NewTransactionHandlerWithGasAndFee(txESDT, 104011, big.NewInt(104000110000000))
	tx.SetInitialPaidFee(initialPaidFee)

	pool := &outport.Pool{
		Txs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(txHash): tx,
		},
		Scrs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(scrHash2): outport.NewTransactionHandlerWithGasAndFee(scr2, 0, big.NewInt(0)),
			string(scrHash1): outport.NewTransactionHandlerWithGasAndFee(scr1, 0, big.NewInt(0)),
		},
	}
	err = esProc.SaveTransactions(body, header, pool, nil)
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t, readExpectedResult("./testdata/esdtTransfer/esdt-transfer.json"), string(genericResponse.Docs[0].Source))
}
