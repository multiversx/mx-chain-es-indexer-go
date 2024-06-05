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

func TestESDTTransferTooMuchGasProvided(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	txHash := []byte("esdtTransfer")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
		ShardID:   0,
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

	address1 := "erd1ef6470tjdtlgpa9f6g3ae4nsedmjg0gv6w73v32xtvhkfff993hq750xl9"
	address2 := "erd13u7zyekzvdvzek8768r5gau9p6677ufppsjuklu9e6t7yx7rhg4s68e2ze"
	txESDT := &transaction.Transaction{
		Nonce:    6,
		SndAddr:  decodeAddress(address1),
		RcvAddr:  decodeAddress(address2),
		GasLimit: 104011,
		GasPrice: 1000000000,
		Data:     []byte("ESDTTransfer@54474e2d383862383366@0a"),
		Value:    big.NewInt(0),
	}

	scrHash1 := []byte("scrHash1ESDTTransfer")
	scr1 := &smartContractResult.SmartContractResult{
		Nonce:          7,
		GasPrice:       1000000000,
		SndAddr:        decodeAddress(address2),
		RcvAddr:        decodeAddress(address1),
		Data:           []byte("@6f6b"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
		ReturnMessage:  []byte("@too much gas provided: gas needed = 372000, gas remained = 2250001"),
	}

	scr2 := &smartContractResult.SmartContractResult{
		Nonce:          7,
		GasPrice:       1000000000,
		SndAddr:        decodeAddress(address2),
		RcvAddr:        decodeAddress(address1),
		Data:           []byte("@6f6b"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
	}

	initialPaidFee, _ := big.NewInt(0).SetString("104000110000000", 10)
	txInfo := &outport.TxInfo{
		Transaction: txESDT,
		FeeInfo: &outport.FeeInfo{
			GasUsed:        104011,
			Fee:            initialPaidFee,
			InitialPaidFee: big.NewInt(104000110000000),
		},
		ExecutionOrder: 0,
	}

	pool := &outport.TransactionPool{
		Transactions: map[string]*outport.TxInfo{
			hex.EncodeToString(txHash): txInfo,
		},
		SmartContractResults: map[string]*outport.SCRInfo{
			hex.EncodeToString(scrHash2): {SmartContractResult: scr2, FeeInfo: &outport.FeeInfo{}},
			hex.EncodeToString(scrHash1): {SmartContractResult: scr1, FeeInfo: &outport.FeeInfo{}},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t, readExpectedResult("./testdata/esdtTransfer/esdt-transfer.json"), string(genericResponse.Docs[0].Source))
}

func TestESDTTransferCrossShardWithRefundOnSourceShard(t *testing.T) {
	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	txHash := []byte("esdtTransferCross")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 10101,
		ShardID:   2,
	}

	scrHash1 := []byte("scrHash1ESDTTransferOnSourceShard")
	body := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.TxBlock,
				SenderShardID:   2,
				ReceiverShardID: 1,
				TxHashes:        [][]byte{txHash},
			},
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   2,
				ReceiverShardID: 2,
				TxHashes:        [][]byte{scrHash1},
			},
		},
	}

	address1 := "erd1ef6470tjdtlgpa9f6g3ae4nsedmjg0gv6w73v32xtvhkfff993hq750xl9"
	address2 := "erd13u7zyekzvdvzek8768r5gau9p6677ufppsjuklu9e6t7yx7rhg4s68e2ze"

	txESDT := outport.TxInfo{
		Transaction: &transaction.Transaction{
			Nonce:    1,
			SndAddr:  decodeAddress(address1),
			RcvAddr:  decodeAddress(address2),
			GasLimit: 500_000,
			GasPrice: 1000000000,
			Data:     []byte("ESDTTransfer@54474e2d383862383366@0a"),
			Value:    big.NewInt(0),
		},
		FeeInfo: &outport.FeeInfo{
			GasUsed:        334_000,
			Fee:            big.NewInt(136000000000000),
			InitialPaidFee: big.NewInt(137660000000000),
		},
	}

	refundValueBig, _ := big.NewInt(0).SetString("1660000000000", 10)
	scrRefund := &smartContractResult.SmartContractResult{
		Nonce:          20,
		SndAddr:        decodeAddress(address2),
		RcvAddr:        decodeAddress(address1),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
		Value:          refundValueBig,
		Data:           []byte("@6f6b"),
	}

	pool := &outport.TransactionPool{
		Transactions: map[string]*outport.TxInfo{
			hex.EncodeToString(txHash): &txESDT,
		},
		SmartContractResults: map[string]*outport.SCRInfo{
			hex.EncodeToString(scrHash1): {SmartContractResult: scrRefund, FeeInfo: &outport.FeeInfo{}},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.OperationsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t, readExpectedResult("./testdata/esdtTransfer/esdt-transfer-cross-shard-on-source.json"), string(genericResponse.Docs[0].Source))

	header = &dataBlock.Header{
		Round:     55,
		TimeStamp: 10102,
		ShardID:   1,
	}
	txESDT.FeeInfo = &outport.FeeInfo{
		GasUsed:        500_000,
		Fee:            big.NewInt(137660000000000),
		InitialPaidFee: big.NewInt(137660000000000),
	}

	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.OperationsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t, readExpectedResult("./testdata/esdtTransfer/esdt-transfer-cross-shard-on-destination.json"), string(genericResponse.Docs[0].Source))
}
