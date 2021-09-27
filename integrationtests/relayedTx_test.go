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
	expectedRelayedTxSource      = `{"miniBlockHash":"fed7c174a849c30b88c36a26453407f1b95970941d0872e603e641c5c804104a","nonce":1196667,"round":50,"value":"0","receiver":"6572643134657961796672766c72687a66727767357a776c65756132356d6b7a676e6367676e33356e766336786876357978776d6c326573306633646874","sender":"657264316b376a3665776a736c61347a73677638763666366665336476726b677633643064396a6572637a773435687a6564687965643873683275333475","receiverShard":0,"senderShard":0,"gasPrice":1000000000,"gasLimit":16610000,"gasUsed":16610000,"fee":"1760000000000000","data":"cmVsYXllZFR4QDdiMjI2ZTZmNmU2MzY1MjIzYTMyMmMyMjc2NjE2Yzc1NjUyMjNhMzAyYzIyNzI2NTYzNjU2OTc2NjU3MjIyM2EyMjQxNDE0MTQxNDE0MTQxNDE0MTQxNDE0NjQxNDk3NDY3MzczODM1MmY3MzZjNzM1NTQxNDg2ODZiNTczMzQ1Njk2MjRjNmU0NzUyNGI3NjQ5NmY0ZTRkM2QyMjJjMjI3MzY1NmU2NDY1NzIyMjNhMjI3MjZiNmU1MzRhNDc3YTM0Mzc2OTUzNGU3OTRiNDM2NDJmNTA0ZjcxNzA3NTc3NmI1NDc3Njg0NTM0MzA2ZDdhNDc2YTU4NWE1MTY4NmU2MjJiNzI0ZDNkMjIyYzIyNjc2MTczNTA3MjY5NjM2NTIyM2EzMTMwMzAzMDMwMzAzMDMwMzAzMDJjMjI2NzYxNzM0YzY5NmQ2OTc0MjIzYTMxMzUzMDMwMzAzMDMwMzAyYzIyNjQ2MTc0NjEyMjNhMjI2MzMyNDYzMjVhNTU0NjMwNjQ0NzU2N2E2NDQ3NDYzMDYxNTczOTc1NTE0NDQ2Njg1OTdhNDkzMTRkNmE1OTM1NTk2ZDUxMzM1YTQ0NDk3NzU5MzI0YTY5NTk1NDRkMzE1OTZkNTY2YzRmNDQ1OTMxNGQ0NDY0Njg0ZjU3NGU2YTRlN2E2NzdhNWE0NzU1Nzc0ZjQ0NWE2OTRlNDQ0NTMzNGU1NDZiMzQ1YTU0NTE3YTU5NTQ0ZTZiNWE2YTU2NmE1OTMyNDU3OTVhNTQ2ODY4NGQ2YTZjNDE0ZDZhNTEzNDRlNTQ2NzdhNGQ1NzRlNmQ0ZDU0NDUzMDRkNTQ1NjZkNTk2YTQxMzU0ZDZhNjM3NzRlNDQ1MTMyNGU1NzU1MzI0ZTdhNTk3YTU5NTc0ZDMxNGY0NDQ1MzQ1YTU0NjczMTRlNDc1MTM0NTk1NzUyNmQ0ZTU0NDE3YTU5NmE2MzM1NGQ2YTZjNmI0ZjU0NTI2YzRlNmQ0OTc5NGU2YTQ5Nzc1YTY3M2QzZDIyMmMyMjYzNjg2MTY5NmU0OTQ0MjIzYTIyNGQ1MTNkM2QyMjJjMjI3NjY1NzI3MzY5NmY2ZTIyM2EzMTJjMjI3MzY5Njc2ZTYxNzQ3NTcyNjUyMjNhMjI1MjM5NDYyYjM0NTQ2MzUyNDE1YTM4NmQ3NzcxMzI0NTU5MzAzMTYzNTk2YzMzNzY2MjcxNmM0NjY1NzE3NjM4N2E3NjQ3NGE3NzVhNjgzMzU5NGQ0ZjU1NmI0MjM0NjQzNDUxNTc0ZTY2Mzc2NzQ0NjI2YzQ4NDgzMjU3NmI3MTYxNGE3NjYxNDg0NTc0NDM1NjYxNzA0OTcxMzM2NTM1NjU2MjM4NGU0MTc3M2QzZDIyN2Q=","signature":"","timestamp":5040,"status":"success","searchOrder":0,"hasScResults":true}`
	expectedRelayedTxAfterRefund = `{"miniBlockHash":"fed7c174a849c30b88c36a26453407f1b95970941d0872e603e641c5c804104a","nonce":1196667,"round":50,"value":"0","receiver":"6572643134657961796672766c72687a66727767357a776c65756132356d6b7a676e6367676e33356e766336786876357978776d6c326573306633646874","sender":"657264316b376a3665776a736c61347a73677638763666366665336476726b677633643064396a6572637a773435687a6564687965643873683275333475","receiverShard":0,"senderShard":0,"gasPrice":1000000000,"gasLimit":16610000,"gasUsed":7982817,"fee":"1673728170000000","data":"cmVsYXllZFR4QDdiMjI2ZTZmNmU2MzY1MjIzYTMyMmMyMjc2NjE2Yzc1NjUyMjNhMzAyYzIyNzI2NTYzNjU2OTc2NjU3MjIyM2EyMjQxNDE0MTQxNDE0MTQxNDE0MTQxNDE0NjQxNDk3NDY3MzczODM1MmY3MzZjNzM1NTQxNDg2ODZiNTczMzQ1Njk2MjRjNmU0NzUyNGI3NjQ5NmY0ZTRkM2QyMjJjMjI3MzY1NmU2NDY1NzIyMjNhMjI3MjZiNmU1MzRhNDc3YTM0Mzc2OTUzNGU3OTRiNDM2NDJmNTA0ZjcxNzA3NTc3NmI1NDc3Njg0NTM0MzA2ZDdhNDc2YTU4NWE1MTY4NmU2MjJiNzI0ZDNkMjIyYzIyNjc2MTczNTA3MjY5NjM2NTIyM2EzMTMwMzAzMDMwMzAzMDMwMzAzMDJjMjI2NzYxNzM0YzY5NmQ2OTc0MjIzYTMxMzUzMDMwMzAzMDMwMzAyYzIyNjQ2MTc0NjEyMjNhMjI2MzMyNDYzMjVhNTU0NjMwNjQ0NzU2N2E2NDQ3NDYzMDYxNTczOTc1NTE0NDQ2Njg1OTdhNDkzMTRkNmE1OTM1NTk2ZDUxMzM1YTQ0NDk3NzU5MzI0YTY5NTk1NDRkMzE1OTZkNTY2YzRmNDQ1OTMxNGQ0NDY0Njg0ZjU3NGU2YTRlN2E2NzdhNWE0NzU1Nzc0ZjQ0NWE2OTRlNDQ0NTMzNGU1NDZiMzQ1YTU0NTE3YTU5NTQ0ZTZiNWE2YTU2NmE1OTMyNDU3OTVhNTQ2ODY4NGQ2YTZjNDE0ZDZhNTEzNDRlNTQ2NzdhNGQ1NzRlNmQ0ZDU0NDUzMDRkNTQ1NjZkNTk2YTQxMzU0ZDZhNjM3NzRlNDQ1MTMyNGU1NzU1MzI0ZTdhNTk3YTU5NTc0ZDMxNGY0NDQ1MzQ1YTU0NjczMTRlNDc1MTM0NTk1NzUyNmQ0ZTU0NDE3YTU5NmE2MzM1NGQ2YTZjNmI0ZjU0NTI2YzRlNmQ0OTc5NGU2YTQ5Nzc1YTY3M2QzZDIyMmMyMjYzNjg2MTY5NmU0OTQ0MjIzYTIyNGQ1MTNkM2QyMjJjMjI3NjY1NzI3MzY5NmY2ZTIyM2EzMTJjMjI3MzY5Njc2ZTYxNzQ3NTcyNjUyMjNhMjI1MjM5NDYyYjM0NTQ2MzUyNDE1YTM4NmQ3NzcxMzI0NTU5MzAzMTYzNTk2YzMzNzY2MjcxNmM0NjY1NzE3NjM4N2E3NjQ3NGE3NzVhNjgzMzU5NGQ0ZjU1NmI0MjM0NjQzNDUxNTc0ZTY2Mzc2NzQ0NjI2YzQ4NDgzMjU3NmI3MTYxNGE3NjYxNDg0NTc0NDM1NjYxNzA0OTcxMzM2NTM1NjU2MjM4NGU0MTc3M2QzZDIyN2Q=","signature":"","timestamp":5040,"status":"success","searchOrder":0,"hasScResults":true}`
)

func TestRelayedTransactionGasUsed(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	txHash := []byte("relayedTx")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}

	scrHash1 := []byte("scrHashRefund")
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
				TxHashes:        [][]byte{scrHash1},
			},
		},
	}

	initialTx := &transaction.Transaction{
		Nonce:    1196667,
		SndAddr:  []byte("erd1k7j6ewjsla4zsgv8v6f6fe3dvrkgv3d0d9jerczw45hzedhyed8sh2u34u"),
		RcvAddr:  []byte("erd14eyayfrvlrhzfrwg5zwleua25mkzgncggn35nvc6xhv5yxwml2es0f3dht"),
		GasLimit: 16610000,
		GasPrice: 1000000000,
		Data:     []byte("relayedTx@7b226e6f6e6365223a322c2276616c7565223a302c227265636569766572223a22414141414141414141414146414974673738352f736c73554148686b57334569624c6e47524b76496f4e4d3d222c2273656e646572223a22726b6e534a477a343769534e794b43642f504f717075776b5477684534306d7a476a585a51686e622b724d3d222c226761735072696365223a313030303030303030302c226761734c696d6974223a31353030303030302c2264617461223a22633246325a5546306447567a644746306157397551444668597a49314d6a5935596d51335a44497759324a6959544d31596d566c4f4459314d4464684f574e6a4e7a677a5a4755774f445a694e4445334e546b345a54517a59544e6b5a6a566a593245795a5468684d6a6c414d6a51344e54677a4d574e6d4d5445304d54566d596a41354d6a63774e4451324e5755324e7a597a59574d314f4445345a5467314e4751345957526d4e54417a596a63354d6a6c6b4f54526c4e6d49794e6a49775a673d3d222c22636861696e4944223a224d513d3d222c2276657273696f6e223a312c227369676e6174757265223a225239462b34546352415a386d7771324559303163596c337662716c46657176387a76474a775a6833594d4f556b4234643451574e66376744626c484832576b71614a76614845744356617049713365356562384e41773d3d227d"),
		Value:    big.NewInt(0),
	}

	scr1 := &smartContractResult.SmartContractResult{
		Nonce:          2,
		GasPrice:       1000000000,
		GasLimit:       14732500,
		SndAddr:        []byte("erd14eyayfrvlrhzfrwg5zwleua25mkzgncggn35nvc6xhv5yxwml2es0f3dht"),
		RcvAddr:        []byte("erd1qqqqqqqqqqqqqpgq3dswlnnlkfd3gqrcv3dhzgnvh8ryf27g5rfsecnn2s"),
		Data:           []byte("aveAttestation@1ac25269bd7d20cbba35bee86507a9cc783de086b417598e43a3df5cca2e8a29@2485831cf11415fb092704465e6763ac5818e854d8adf503b7929d94e6b2620f"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
	}

	pool := &indexer.Pool{
		Txs: map[string]coreData.TransactionHandler{
			string(txHash): initialTx,
		},
		Scrs: map[string]coreData.TransactionHandler{
			string(scrHash1): scr1,
		},
	}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, []byte(expectedRelayedTxSource), genericResponse.Docs[0].Source)

	// EXECUTE transfer on the destination shard
	bodyDstShard := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   0,
				ReceiverShardID: 1,
				TxHashes:        [][]byte{scrHash1},
			},
		},
	}
	scrWithRefund := []byte("scrWithRefund")
	refundValueBig, _ := big.NewInt(0).SetString("86271830000000", 10)
	poolDstShard := &indexer.Pool{
		Scrs: map[string]coreData.TransactionHandler{
			string(scrHash1): scr1,
			string(scrWithRefund): &smartContractResult.SmartContractResult{
				Nonce:          3,
				SndAddr:        []byte("erd1qqqqqqqqqqqqqpgq3dswlnnlkfd3gqrcv3dhzgnvh8ryf27g5rfsecnn2s"),
				RcvAddr:        []byte("erd1k7j6ewjsla4zsgv8v6f6fe3dvrkgv3d0d9jerczw45hzedhyed8sh2u34u"),
				PrevTxHash:     []byte("f639cb7a0231191e04ec19dcb1359bd93a03fe8dc4a28a80d00835c5d1c988f8"),
				OriginalTxHash: txHash,
				Value:          refundValueBig,
				Data:           []byte(""),
				ReturnMessage:  []byte("gas refund for relayer"),
			},
		},
	}

	err = esProc.SaveTransactions(bodyDstShard, header, poolDstShard)
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, []byte(expectedRelayedTxAfterRefund), genericResponse.Docs[0].Source)
}
