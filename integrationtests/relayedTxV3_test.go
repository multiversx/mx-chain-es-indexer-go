//go:build integrationtests

package integrationtests

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	indexerdata "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
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

func TestRelayedTxV3WithSignalErrorAndCompletedEvent(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	txHash := []byte("relayedTxV3WithSignalErrorAndCompletedEvent")
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
		Nonce:    1000,
		SndAddr:  decodeAddress("erd1ykqd64fxxpp4wsz0v7sjqem038wfpzlljhx4mhwx8w9lcxmdzcfszrp64a"),
		RcvAddr:  decodeAddress("erd1ykqd64fxxpp4wsz0v7sjqem038wfpzlljhx4mhwx8w9lcxmdzcfszrp64a"),
		GasLimit: 300_000,
		GasPrice: 1000000000,
		Value:    big.NewInt(0),
		InnerTransactions: []*transaction.Transaction{
			{
				Nonce:    5,
				SndAddr:  decodeAddress("erd10ksryjr065ad5475jcg82pnjfg9j9qtszjsrp24anl6ym7cmeddshwnru8"),
				RcvAddr:  decodeAddress("erd1aduqqezzw0u3j7tywlq3mrl0yn4z6f6vytdju8gg0neq38fauyzsa5yy6r"),
				GasLimit: 50_000,
				GasPrice: 1000000000,
				Value:    big.NewInt(10000000000000000),
			},
			{
				Nonce:    3,
				SndAddr:  decodeAddress("erd10ksryjr065ad5475jcg82pnjfg9j9qtszjsrp24anl6ym7cmeddshwnru8"),
				RcvAddr:  decodeAddress("erd1aduqqezzw0u3j7tywlq3mrl0yn4z6f6vytdju8gg0neq38fauyzsa5yy6r"),
				GasLimit: 50_000,
				GasPrice: 1000000000,
				Value:    big.NewInt(10000000000000000),
			},
			{
				Nonce:    4,
				SndAddr:  decodeAddress("erd10ksryjr065ad5475jcg82pnjfg9j9qtszjsrp24anl6ym7cmeddshwnru8"),
				RcvAddr:  decodeAddress("erd1aduqqezzw0u3j7tywlq3mrl0yn4z6f6vytdju8gg0neq38fauyzsa5yy6r"),
				GasLimit: 50_000,
				GasPrice: 1000000000,
				Value:    big.NewInt(10000000000000000),
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
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString(txHash),
				Log: &transaction.Log{
					Address: decodeAddress("erd1ykqd64fxxpp4wsz0v7sjqem038wfpzlljhx4mhwx8w9lcxmdzcfszrp64a"),
					Events: []*transaction.Event{
						{
							Address:    decodeAddress("erd1ykqd64fxxpp4wsz0v7sjqem038wfpzlljhx4mhwx8w9lcxmdzcfszrp64a"),
							Identifier: []byte(core.CompletedTxEventIdentifier),
							Topics:     [][]byte{[]byte("t1"), []byte("t2")},
						},
						nil,
					},
				},
			},
			{
				TxHash: hex.EncodeToString(txHash),
				Log: &transaction.Log{
					Address: decodeAddress("erd1ykqd64fxxpp4wsz0v7sjqem038wfpzlljhx4mhwx8w9lcxmdzcfszrp64a"),
					Events: []*transaction.Event{
						{
							Address:    decodeAddress("erd1ykqd64fxxpp4wsz0v7sjqem038wfpzlljhx4mhwx8w9lcxmdzcfszrp64a"),
							Identifier: []byte(core.SignalErrorOperation),
							Topics:     [][]byte{[]byte("t1"), []byte("t2")},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/relayedTxV3/relayed-tx-v3-with-events.json"),
		string(genericResponse.Docs[0].Source),
	)
}

func TestRelayedV3WithSCRCross(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	txHash := []byte("relayedTxV3WithScrCross")
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
		Nonce:    1000,
		SndAddr:  decodeAddress("erd1ykqd64fxxpp4wsz0v7sjqem038wfpzlljhx4mhwx8w9lcxmdzcfszrp64a"),
		RcvAddr:  decodeAddress("erd1ykqd64fxxpp4wsz0v7sjqem038wfpzlljhx4mhwx8w9lcxmdzcfszrp64a"),
		GasLimit: 300_000,
		GasPrice: 1000000000,
		Value:    big.NewInt(0),
		InnerTransactions: []*transaction.Transaction{
			{
				Nonce:    5,
				SndAddr:  decodeAddress("erd10ksryjr065ad5475jcg82pnjfg9j9qtszjsrp24anl6ym7cmeddshwnru8"),
				RcvAddr:  decodeAddress("erd1aduqqezzw0u3j7tywlq3mrl0yn4z6f6vytdju8gg0neq38fauyzsa5yy6r"),
				GasLimit: 50_000,
				GasPrice: 1000000000,
				Value:    big.NewInt(10000000000000000),
			},
			{
				Nonce:    3,
				SndAddr:  decodeAddress("erd10ksryjr065ad5475jcg82pnjfg9j9qtszjsrp24anl6ym7cmeddshwnru8"),
				RcvAddr:  decodeAddress("erd1aduqqezzw0u3j7tywlq3mrl0yn4z6f6vytdju8gg0neq38fauyzsa5yy6r"),
				GasLimit: 50_000,
				GasPrice: 1000000000,
				Value:    big.NewInt(10000000000000000),
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
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString(txHash),
				Log: &transaction.Log{
					Address: decodeAddress("erd1ykqd64fxxpp4wsz0v7sjqem038wfpzlljhx4mhwx8w9lcxmdzcfszrp64a"),
					Events: []*transaction.Event{
						{
							Address:    decodeAddress("erd1ykqd64fxxpp4wsz0v7sjqem038wfpzlljhx4mhwx8w9lcxmdzcfszrp64a"),
							Identifier: []byte(core.CompletedTxEventIdentifier),
							Topics:     [][]byte{[]byte("t1"), []byte("t2")},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/relayedTxV3/relayed-v3-execution-source.json"),
		string(genericResponse.Docs[0].Source),
	)

	// execute scr on destination
	header = &dataBlock.Header{
		Round:     60,
		TimeStamp: 6040,
	}

	scrInfo := &outport.SCRInfo{
		SmartContractResult: &smartContractResult.SmartContractResult{
			OriginalTxHash: txHash,
		},
		FeeInfo: &outport.FeeInfo{
			Fee:            big.NewInt(0),
			InitialPaidFee: big.NewInt(0),
		},
		ExecutionOrder: 0,
	}

	pool = &outport.TransactionPool{
		SmartContractResults: map[string]*outport.SCRInfo{
			hex.EncodeToString([]byte("scr")): scrInfo,
		},
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("scr")),
				Log: &transaction.Log{
					Address: decodeAddress("erd1ykqd64fxxpp4wsz0v7sjqem038wfpzlljhx4mhwx8w9lcxmdzcfszrp64a"),
					Events: []*transaction.Event{
						{
							Address:    decodeAddress("erd1ykqd64fxxpp4wsz0v7sjqem038wfpzlljhx4mhwx8w9lcxmdzcfszrp64a"),
							Identifier: []byte(core.SignalErrorOperation),
							Topics:     [][]byte{[]byte("t1"), []byte("t2")},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/relayedTxV3/relayed-v3-execution-scr-on-dest.json"),
		string(genericResponse.Docs[0].Source),
	)
}
