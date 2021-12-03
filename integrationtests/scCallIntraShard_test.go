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
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

const (
	claimRewardsTx = `{"miniBlockHash":"60b38b11110d28d1b361359f9688bb041bb9180219a612a83ff00dcc0db4d607","nonce":101,"round":50,"value":"0","receiver":"65726431717171717171717171717171717067717877616b7432673775396174736e723033677163676d68637633387074376d6b64393471367368757774","sender":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","receiverShard":0,"senderShard":0,"gasPrice":1000000000,"gasLimit":250000000,"gasUsed":33891715,"fee":"406237150000000","data":"Y2xhaW1SZXdhcmRz","signature":"","timestamp":5040,"status":"success","searchOrder":0,"hasScResults":true,"scresults":{"7363724861736831":{"nonce":102,"gasLimit":0,"gasPrice":1000000000,"value":"2161082850000000","sender":"65726431717171717171717171717171717067717877616b7432673775396174736e723033677163676d68637633387074376d6b64393471367368757774","receiver":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","senderShard":0,"receiverShard":0,"data":"QDZmNmI=","prevTxHash":"747848617368436c61696d52657761726473","originalTxHash":"747848617368436c61696d52657761726473","callType":"0","timestamp":5040}}}`
	scCallFailTx   = `{"receiver":"6572643171717171717171717171717171717170717171717171717171717171717171717171717171717171717171717166686c6c6c6c73637274353672","data":"ZGVsZWdhdGU=","signature":"","fee":"181380000000000","nonce":46,"gasLimit":12000000,"gasUsed":12000000,"miniBlockHash":"5d04f80b044352bfbbde123702323eae07fdd8ca77f24f256079006058b6e7b4","round":50,"hasScResults":true,"sender":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","receiverShard":0,"senderShard":0,"scresults":{"7478486173684d657461636861696e":{"prevTxHash":"74","receiver":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","data":"QHVzZXIgZXJyb3I=","returnMessage":"total delegation cap reached","nonce":46,"callType":"0","gasLimit":0,"originalTxHash":"74","sender":"6572643171717171717171717171717171717170717171717171717171717171717171717171717171717171717171717166686c6c6c6c73637274353672","receiverShard":0,"senderShard":0,"value":"5000000000000000000","gasPrice":0,"timestamp":5040}},"value":"5000000000000000000","gasPrice":1000000000,"timestamp":5040,"status":"fail","searchOrder":0}`
)

func TestTransactionWithSCCallFail(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
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
	pool := &indexer.Pool{
		Txs: map[string]coreData.TransactionHandler{
			string(txHash): &transaction.Transaction{
				Nonce:    46,
				SndAddr:  []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				RcvAddr:  []byte("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqfhllllscrt56r"),
				GasLimit: 12000000,
				GasPrice: 1000000000,
				Data:     []byte("delegate"),
				Value:    refundValueBig,
			},
		},
		Scrs: map[string]coreData.TransactionHandler{
			string(scrHash1): &smartContractResult.SmartContractResult{
				Nonce:          46,
				Value:          refundValueBig,
				GasPrice:       0,
				SndAddr:        []byte("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqfhllllscrt56r"),
				RcvAddr:        []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				Data:           []byte("@user error"),
				PrevTxHash:     txHash,
				OriginalTxHash: txHash,
				ReturnMessage:  []byte("total delegation cap reached"),
			},
		},
	}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerData.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, []byte(scCallFailTx), genericResponse.Docs[0].Source)
}

func TestTransactionWithScCallSuccess(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
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

	refundValueBig, _ := big.NewInt(0).SetString("2161082850000000", 10)
	pool := &indexer.Pool{
		Txs: map[string]coreData.TransactionHandler{
			string(txHash): &transaction.Transaction{
				Nonce:    101,
				SndAddr:  []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				RcvAddr:  []byte("erd1qqqqqqqqqqqqqpgqxwakt2g7u9atsnr03gqcgmhcv38pt7mkd94q6shuwt"),
				GasLimit: 250000000,
				GasPrice: 1000000000,
				Data:     []byte("claimRewards"),
				Value:    big.NewInt(0),
			},
		},
		Scrs: map[string]coreData.TransactionHandler{
			string(scrHash1): &smartContractResult.SmartContractResult{
				Nonce:          102,
				Value:          refundValueBig,
				GasPrice:       1000000000,
				SndAddr:        []byte("erd1qqqqqqqqqqqqqpgqxwakt2g7u9atsnr03gqcgmhcv38pt7mkd94q6shuwt"),
				RcvAddr:        []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				Data:           []byte("@6f6b"),
				PrevTxHash:     txHash,
				OriginalTxHash: txHash,
			},
		},
	}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerData.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, []byte(claimRewardsTx), genericResponse.Docs[0].Source)
}
