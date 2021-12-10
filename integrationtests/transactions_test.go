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
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

const moveBalanceTransaction = `{"miniBlockHash":"24c374c9405540e88a36959ea83eede6ad50f6872f82d2e2a2280975615e1811","nonce":1,"round":50,"value":"1234","receiver":"7265636569766572","sender":"73656e646572","receiverShard":0,"senderShard":0,"gasPrice":1000000000,"gasLimit":70000,"gasUsed":62000,"fee":"62000000000000","data":"dHJhbnNmZXI=","signature":"","timestamp":5040,"status":"success","searchOrder":0}`

func TestElasticIndexerSaveTransactions(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
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
	pool := &indexer.Pool{
		Txs: map[string]coreData.TransactionHandler{
			string(txHash): &transaction.Transaction{
				Nonce:    1,
				SndAddr:  []byte("sender"),
				RcvAddr:  []byte("receiver"),
				GasLimit: 70000,
				GasPrice: 1000000000,
				Data:     []byte("transfer"),
				Value:    big.NewInt(1234),
			},
		},
	}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerData.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, []byte(moveBalanceTransaction), genericResponse.Docs[0].Source)
}
