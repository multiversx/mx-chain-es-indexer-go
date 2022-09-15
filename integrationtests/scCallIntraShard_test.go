//go:build integrationtests

package integrationtests

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	indexerData "github.com/ElrondNetwork/elastic-indexer-go/process/dataindexer"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestTransactionWithSCCallFail(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, shardCoordinator)
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

	refundValueBig, _ := big.NewInt(0).SetString("5000000000000000000", 10)
	tx := outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
		Nonce:    46,
		SndAddr:  []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
		RcvAddr:  []byte("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqfhllllscrt56r"),
		GasLimit: 12000000,
		GasPrice: 1000000000,
		Data:     []byte("delegate"),
		Value:    refundValueBig,
	}, 12000000, big.NewInt(181380000000000))
	tx.SetInitialPaidFee(big.NewInt(181380000000000))

	pool := &outport.Pool{
		Txs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(txHash): tx,
		},
		Scrs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(scrHash1): outport.NewTransactionHandlerWithGasAndFee(&smartContractResult.SmartContractResult{
				Nonce:          46,
				Value:          refundValueBig,
				GasPrice:       0,
				SndAddr:        []byte("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqfhllllscrt56r"),
				RcvAddr:        []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				Data:           []byte("@75736572206572726f72"),
				PrevTxHash:     txHash,
				OriginalTxHash: txHash,
				ReturnMessage:  []byte("total delegation cap reached"),
			}, 0, big.NewInt(0)),
		},
	}
	err = esProc.SaveTransactions(body, header, pool, nil, false)
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

	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, shardCoordinator)
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

	tx := outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
		Nonce:    101,
		SndAddr:  []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
		RcvAddr:  []byte("erd1qqqqqqqqqqqqqpgqxwakt2g7u9atsnr03gqcgmhcv38pt7mkd94q6shuwt"),
		GasLimit: 250000000,
		GasPrice: 1000000000,
		Data:     []byte("claimRewards"),
		Value:    big.NewInt(0),
	}, 33891715, big.NewInt(406237150000000))
	tx.SetInitialPaidFee(big.NewInt(2567320000000000))

	refundValueBig, _ := big.NewInt(0).SetString("2161082850000000", 10)
	pool := &outport.Pool{
		Txs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(txHash): tx,
		},
		Scrs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(scrHash1): outport.NewTransactionHandlerWithGasAndFee(&smartContractResult.SmartContractResult{
				Nonce:          102,
				Value:          refundValueBig,
				GasPrice:       1000000000,
				SndAddr:        []byte("erd1qqqqqqqqqqqqqpgqxwakt2g7u9atsnr03gqcgmhcv38pt7mkd94q6shuwt"),
				RcvAddr:        []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				Data:           []byte("@6f6b"),
				PrevTxHash:     txHash,
				OriginalTxHash: txHash,
			}, 0, big.NewInt(0)),
		},
	}
	err = esProc.SaveTransactions(body, header, pool, nil, false)
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
