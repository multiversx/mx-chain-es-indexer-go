//go:build integrationtests

package integrationtests

import (
	"encoding/hex"
	"math/big"
	"testing"

	indexerData "github.com/ElrondNetwork/elastic-indexer-go/process/dataindexer"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestElasticIndexerSaveTransactions(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
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
	tx := outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
		Nonce:    1,
		SndAddr:  []byte("sender"),
		RcvAddr:  []byte("receiver"),
		GasLimit: 70000,
		GasPrice: 1000000000,
		Data:     []byte("transfer"),
		Value:    big.NewInt(1234),
	}, 62000, big.NewInt(62000000000000))
	tx.SetInitialPaidFee(big.NewInt(62080000000000))
	pool := &outport.Pool{
		Txs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(txHash): tx,
		},
	}
	err = esProc.SaveTransactions(body, header, pool, nil, false, testsNumOfShards)
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerData.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/transactions/move-balance.json"),
		string(genericResponse.Docs[0].Source),
	)
}
