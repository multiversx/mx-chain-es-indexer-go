//go:build integrationtests

package integrationtests

import (
	"context"
	"encoding/hex"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	indexerdata "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func TestRelayedTxV3(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	txHash := []byte("relayedTxV3")
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

	address1 := "erd1k7j6ewjsla4zsgv8v6f6fe3dvrkgv3d0d9jerczw45hzedhyed8sh2u34u"
	address2 := "erd14eyayfrvlrhzfrwg5zwleua25mkzgncggn35nvc6xhv5yxwml2es0f3dht"
	initialTx := &transaction.Transaction{
		Nonce:    1000,
		SndAddr:  decodeAddress(address1),
		RcvAddr:  decodeAddress(address2),
		GasLimit: 15406000,
		GasPrice: 1000000000,
		Value:    big.NewInt(0),
		InnerTransactions: []*transaction.Transaction{
			{
				Nonce:    10,
				SndAddr:  decodeAddress(address1),
				RcvAddr:  decodeAddress(address2),
				GasLimit: 15406000,
				GasPrice: 1000000000,
				Value:    big.NewInt(0),
			},
			{
				Nonce:    20,
				SndAddr:  decodeAddress(address1),
				RcvAddr:  decodeAddress(address2),
				GasLimit: 15406000,
				GasPrice: 1000000000,
				Value:    big.NewInt(1000),
			},
		},
	}

	txInfo := &outport.TxInfo{
		Transaction: initialTx,
		FeeInfo: &outport.FeeInfo{
			GasUsed:        10556000,
			Fee:            big.NewInt(2257820000000000),
			InitialPaidFee: big.NewInt(2306320000000000),
		},
		ExecutionOrder: 0,
	}

	pool := &outport.TransactionPool{
		Transactions: map[string]*outport.TxInfo{
			hex.EncodeToString(txHash): txInfo,
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/relayedTxV3/relayed-tx-v3.json"),
		string(genericResponse.Docs[0].Source),
	)
}
