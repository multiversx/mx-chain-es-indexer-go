//go:build integrationtests

package integrationtests

import (
	"encoding/hex"
	"math/big"
	"testing"

	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	indexerData "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestTransactionWithSCCallFail(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	txHash := []byte("t")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}
	scrHash1 := []byte("txHashMetachain")
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

	address1 := "erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"
	address2 := "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqfhllllscrt56r"
	refundValueBig, _ := big.NewInt(0).SetString("5000000000000000000", 10)
	tx := &transaction.Transaction{
		Nonce:    46,
		SndAddr:  decodeAddress(address1),
		RcvAddr:  decodeAddress(address2),
		GasLimit: 12000000,
		GasPrice: 1000000000,
		Data:     []byte("delegate"),
		Value:    refundValueBig,
	}

	txInfo := &outport.TxInfo{
		Transaction: tx,
		FeeInfo: &outport.FeeInfo{
			GasUsed:        12000000,
			Fee:            big.NewInt(181380000000000),
			InitialPaidFee: big.NewInt(181380000000000),
		},
		ExecutionOrder: 0,
	}

	pool := &outport.TransactionPool{
		Transactions: map[string]*outport.TxInfo{
			hex.EncodeToString(txHash): txInfo,
		},
		SmartContractResults: map[string]*outport.SCRInfo{
			hex.EncodeToString(scrHash1): {SmartContractResult: &smartContractResult.SmartContractResult{
				Nonce:          46,
				Value:          refundValueBig,
				GasPrice:       0,
				SndAddr:        decodeAddress(address2),
				RcvAddr:        decodeAddress(address1),
				Data:           []byte("@75736572206572726f72"),
				PrevTxHash:     txHash,
				OriginalTxHash: txHash,
				ReturnMessage:  []byte("total delegation cap reached"),
			}, FeeInfo: &outport.FeeInfo{}},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerData.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/scCallIntraShard/sc-call-fail.json"),
		string(genericResponse.Docs[0].Source),
	)
}

func TestTransactionWithScCallSuccess(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	txHash := []byte("txHashClaimRewards")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}
	scrHash1 := []byte("scrHash1")
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

	address1 := "erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"
	address2 := "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqfhllllscrt56r"
	tx := &transaction.Transaction{
		Nonce:    101,
		SndAddr:  decodeAddress(address1),
		RcvAddr:  decodeAddress(address2),
		GasLimit: 250000000,
		GasPrice: 1000000000,
		Data:     []byte("claimRewards"),
		Value:    big.NewInt(0),
	}

	txInfo := &outport.TxInfo{
		Transaction: tx,
		FeeInfo: &outport.FeeInfo{
			GasUsed:        33891715,
			Fee:            big.NewInt(406237150000000),
			InitialPaidFee: big.NewInt(2567320000000000),
		},
		ExecutionOrder: 0,
	}

	refundValueBig, _ := big.NewInt(0).SetString("2161082850000000", 10)
	pool := &outport.TransactionPool{
		Transactions: map[string]*outport.TxInfo{
			hex.EncodeToString(txHash): txInfo,
		},
		SmartContractResults: map[string]*outport.SCRInfo{
			hex.EncodeToString(scrHash1): {SmartContractResult: &smartContractResult.SmartContractResult{
				Nonce:          102,
				Value:          refundValueBig,
				GasPrice:       1000000000,
				SndAddr:        decodeAddress(address2),
				RcvAddr:        decodeAddress(address1),
				Data:           []byte("@6f6b"),
				PrevTxHash:     txHash,
				OriginalTxHash: txHash,
			}, FeeInfo: &outport.FeeInfo{}},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerData.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/scCallIntraShard/claim-rewards.json"),
		string(genericResponse.Docs[0].Source),
	)
}
