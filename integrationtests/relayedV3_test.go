//go:build integrationtests

package integrationtests

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"

	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	indexerdata "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestRelayedV3TransactionWithMultipleRefunds(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	txHash := []byte("relayedTxV3WithMultipleRefunds")
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

	initialTx := &transaction.Transaction{
		Nonce:            1000,
		SndAddr:          decodeAddress("erd1ykqd64fxxpp4wsz0v7sjqem038wfpzlljhx4mhwx8w9lcxmdzcfszrp64a"),
		RcvAddr:          decodeAddress("erd1qqqqqqqqqqqqqpgqak8zt22wl2ph4tswtyc39namqx6ysa2sd8ss4xmlj3"),
		RelayerAddr:      decodeAddress("erd10ksryjr065ad5475jcg82pnjfg9j9qtszjsrp24anl6ym7cmeddshwnru8"),
		Signature:        []byte("d"),
		RelayerSignature: []byte("a"),
		GasLimit:         500_000_000,
		GasPrice:         1000000000,
		Value:            big.NewInt(0),
		Data:             []byte("doSomething"),
	}

	txInfo := &outport.TxInfo{
		Transaction: initialTx,
		FeeInfo: &outport.FeeInfo{
			GasUsed:        180_150_000,
			Fee:            big.NewInt(2864760000000000),
			InitialPaidFee: big.NewInt(2864760000000000),
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
		readExpectedResult("./testdata/relayedTxV3/relayed-v3-no-refund.json"),
		string(genericResponse.Docs[0].Source),
	)

	//  execute first SCR with refund
	pool = &outport.TransactionPool{
		SmartContractResults: map[string]*outport.SCRInfo{
			"scrHash": {
				SmartContractResult: &smartContractResult.SmartContractResult{
					OriginalTxHash: txHash,
				},
				FeeInfo: &outport.FeeInfo{
					GasRefunded: 9_692_000,
					Fee:         big.NewInt(96920000000000),
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids = []string{hex.EncodeToString(txHash)}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/relayedTxV3/relayed-v3-with-one-refund.json"),
		string(genericResponse.Docs[0].Source),
	)

	//  execute second SCR with refund
	pool = &outport.TransactionPool{
		SmartContractResults: map[string]*outport.SCRInfo{
			"scrHash": {
				SmartContractResult: &smartContractResult.SmartContractResult{
					OriginalTxHash: txHash,
				},
				FeeInfo: &outport.FeeInfo{
					GasRefunded: 9_692_000,
					Fee:         big.NewInt(96920000000000),
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids = []string{hex.EncodeToString(txHash)}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/relayedTxV3/relayed-v3-with-two-refunds.json"),
		string(genericResponse.Docs[0].Source),
	)
}
