package integrationtests

import (
	"encoding/hex"
	"math/big"
	"testing"

	indexerdata "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

const (
	expectedTxNoSCR   = `{"miniBlockHash":"0aea7125ebcf1b8811521336d63ae511f9c8469dcb469ab16d873d8acd5192a0","nonce":6,"round":50,"value":"0","receiver":"657264313375377a79656b7a7664767a656b38373638723567617539703636373775667070736a756b6c7539653674377978377268673473363865327a65","sender":"65726431656636343730746a64746c67706139663667336165346e7365646d6a6730677636773733763332787476686b6666663939336871373530786c39","receiverShard":0,"senderShard":0,"gasPrice":1000000000,"gasLimit":104011,"gasUsed":59000,"fee":"59000000000000","data":"c2NDYWxs","signature":"","timestamp":5040,"status":"success","searchOrder":0}`
	expectedTxWithSCR = `{"miniBlockHash":"0aea7125ebcf1b8811521336d63ae511f9c8469dcb469ab16d873d8acd5192a0","nonce":6,"round":50,"value":"0","receiver":"657264313375377a79656b7a7664767a656b38373638723567617539703636373775667070736a756b6c7539653674377978377268673473363865327a65","sender":"65726431656636343730746a64746c67706139663667336165346e7365646d6a6730677636773733763332787476686b6666663939336871373530786c39","receiverShard":0,"senderShard":0,"gasPrice":1000000000,"gasLimit":104011,"gasUsed":59000,"fee":"59000000000000","data":"c2NDYWxs","signature":"","timestamp":5040,"status":"success","searchOrder":0,"scresults":{"7363724861736831455344545472616e73666572":{"prevTxHash":"736343616c6c536f6c6f","receiver":"65726431656636343730746a64746c67706139663667336165346e7365646d6a6730677636773733763332787476686b6666663939336871373530786c39","data":"QDZmNmI=","returnMessage":"@too much gas provided: gas needed = 372000, gas remained = 2250001","nonce":7,"callType":"0","gasLimit":0,"originalTxHash":"736343616c6c536f6c6f","sender":"657264313375377a79656b7a7664767a656b38373638723567617539703636373775667070736a756b6c7539653674377978377268673473363865327a65","receiverShard":0,"senderShard":0,"value":"<nil>","gasPrice":1000000000,"timestamp":5040}}}`
)

func TestTransactionWithSoloSCR(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	txHash := []byte("scCallSolo")
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

	tx := &transaction.Transaction{
		Nonce:    6,
		SndAddr:  []byte("erd1ef6470tjdtlgpa9f6g3ae4nsedmjg0gv6w73v32xtvhkfff993hq750xl9"),
		RcvAddr:  []byte("erd13u7zyekzvdvzek8768r5gau9p6677ufppsjuklu9e6t7yx7rhg4s68e2ze"),
		GasLimit: 104011,
		GasPrice: 1000000000,
		Data:     []byte("scCall"),
		Value:    big.NewInt(0),
	}

	pool := &indexer.Pool{
		Txs: map[string]coreData.TransactionHandler{
			string(txHash): tx,
		},
	}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, []byte(expectedTxNoSCR), genericResponse.Docs[0].Source)

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

	// save solo SCR
	pool = &indexer.Pool{
		Scrs: map[string]coreData.TransactionHandler{
			string(scrHash1): scr1,
		},
	}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids = []string{hex.EncodeToString(txHash)}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, []byte(expectedTxWithSCR), genericResponse.Docs[0].Source)
}
