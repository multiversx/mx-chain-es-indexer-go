//go:build integrationtests

package integrationtests

import (
	"encoding/hex"
	"math/big"
	"testing"

	indexerdata "github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

const (
	expectedRelayedTxSource      = `{"miniBlockHash":"fed7c174a849c30b88c36a26453407f1b95970941d0872e603e641c5c804104a","nonce":1196667,"round":50,"value":"0","receiver":"6572643134657961796672766c72687a66727767357a776c65756132356d6b7a676e6367676e33356e766336786876357978776d6c326573306633646874","sender":"657264316b376a3665776a736c61347a73677638763666366665336476726b677633643064396a6572637a773435687a6564687965643873683275333475","receiverShard":0,"senderShard":0,"gasPrice":1000000000,"gasLimit":16610000,"gasUsed":16610000,"fee":"1760000000000000","data":"cmVsYXllZFR4QDdiMjI2ZTZmNmU2MzY1MjIzYTMyMmMyMjc2NjE2Yzc1NjUyMjNhMzAyYzIyNzI2NTYzNjU2OTc2NjU3MjIyM2EyMjQxNDE0MTQxNDE0MTQxNDE0MTQxNDE0NjQxNDk3NDY3MzczODM1MmY3MzZjNzM1NTQxNDg2ODZiNTczMzQ1Njk2MjRjNmU0NzUyNGI3NjQ5NmY0ZTRkM2QyMjJjMjI3MzY1NmU2NDY1NzIyMjNhMjI3MjZiNmU1MzRhNDc3YTM0Mzc2OTUzNGU3OTRiNDM2NDJmNTA0ZjcxNzA3NTc3NmI1NDc3Njg0NTM0MzA2ZDdhNDc2YTU4NWE1MTY4NmU2MjJiNzI0ZDNkMjIyYzIyNjc2MTczNTA3MjY5NjM2NTIyM2EzMTMwMzAzMDMwMzAzMDMwMzAzMDJjMjI2NzYxNzM0YzY5NmQ2OTc0MjIzYTMxMzUzMDMwMzAzMDMwMzAyYzIyNjQ2MTc0NjEyMjNhMjI2MzMyNDYzMjVhNTU0NjMwNjQ0NzU2N2E2NDQ3NDYzMDYxNTczOTc1NTE0NDQ2Njg1OTdhNDkzMTRkNmE1OTM1NTk2ZDUxMzM1YTQ0NDk3NzU5MzI0YTY5NTk1NDRkMzE1OTZkNTY2YzRmNDQ1OTMxNGQ0NDY0Njg0ZjU3NGU2YTRlN2E2NzdhNWE0NzU1Nzc0ZjQ0NWE2OTRlNDQ0NTMzNGU1NDZiMzQ1YTU0NTE3YTU5NTQ0ZTZiNWE2YTU2NmE1OTMyNDU3OTVhNTQ2ODY4NGQ2YTZjNDE0ZDZhNTEzNDRlNTQ2NzdhNGQ1NzRlNmQ0ZDU0NDUzMDRkNTQ1NjZkNTk2YTQxMzU0ZDZhNjM3NzRlNDQ1MTMyNGU1NzU1MzI0ZTdhNTk3YTU5NTc0ZDMxNGY0NDQ1MzQ1YTU0NjczMTRlNDc1MTM0NTk1NzUyNmQ0ZTU0NDE3YTU5NmE2MzM1NGQ2YTZjNmI0ZjU0NTI2YzRlNmQ0OTc5NGU2YTQ5Nzc1YTY3M2QzZDIyMmMyMjYzNjg2MTY5NmU0OTQ0MjIzYTIyNGQ1MTNkM2QyMjJjMjI3NjY1NzI3MzY5NmY2ZTIyM2EzMTJjMjI3MzY5Njc2ZTYxNzQ3NTcyNjUyMjNhMjI1MjM5NDYyYjM0NTQ2MzUyNDE1YTM4NmQ3NzcxMzI0NTU5MzAzMTYzNTk2YzMzNzY2MjcxNmM0NjY1NzE3NjM4N2E3NjQ3NGE3NzVhNjgzMzU5NGQ0ZjU1NmI0MjM0NjQzNDUxNTc0ZTY2Mzc2NzQ0NjI2YzQ4NDgzMjU3NmI3MTYxNGE3NjYxNDg0NTc0NDM1NjYxNzA0OTcxMzM2NTM1NjU2MjM4NGU0MTc3M2QzZDIyN2Q=","signature":"","timestamp":5040,"status":"success","searchOrder":0,"hasScResults":true,"receivers":["000000000000000005008b60efce7fb25b140078645b71226cb9c644abc8a0d3"],"receiversShardIDs":[0],"operation":"transfer","function":"saveAttestation","isRelayed":true}`
	expectedRelayedTxAfterRefund = `{"miniBlockHash":"fed7c174a849c30b88c36a26453407f1b95970941d0872e603e641c5c804104a","nonce":1196667,"round":50,"value":"0","receiver":"6572643134657961796672766c72687a66727767357a776c65756132356d6b7a676e6367676e33356e766336786876357978776d6c326573306633646874","sender":"657264316b376a3665776a736c61347a73677638763666366665336476726b677633643064396a6572637a773435687a6564687965643873683275333475","receiverShard":0,"senderShard":0,"gasPrice":1000000000,"gasLimit":16610000,"gasUsed":7982817,"fee":"1673728170000000","data":"cmVsYXllZFR4QDdiMjI2ZTZmNmU2MzY1MjIzYTMyMmMyMjc2NjE2Yzc1NjUyMjNhMzAyYzIyNzI2NTYzNjU2OTc2NjU3MjIyM2EyMjQxNDE0MTQxNDE0MTQxNDE0MTQxNDE0NjQxNDk3NDY3MzczODM1MmY3MzZjNzM1NTQxNDg2ODZiNTczMzQ1Njk2MjRjNmU0NzUyNGI3NjQ5NmY0ZTRkM2QyMjJjMjI3MzY1NmU2NDY1NzIyMjNhMjI3MjZiNmU1MzRhNDc3YTM0Mzc2OTUzNGU3OTRiNDM2NDJmNTA0ZjcxNzA3NTc3NmI1NDc3Njg0NTM0MzA2ZDdhNDc2YTU4NWE1MTY4NmU2MjJiNzI0ZDNkMjIyYzIyNjc2MTczNTA3MjY5NjM2NTIyM2EzMTMwMzAzMDMwMzAzMDMwMzAzMDJjMjI2NzYxNzM0YzY5NmQ2OTc0MjIzYTMxMzUzMDMwMzAzMDMwMzAyYzIyNjQ2MTc0NjEyMjNhMjI2MzMyNDYzMjVhNTU0NjMwNjQ0NzU2N2E2NDQ3NDYzMDYxNTczOTc1NTE0NDQ2Njg1OTdhNDkzMTRkNmE1OTM1NTk2ZDUxMzM1YTQ0NDk3NzU5MzI0YTY5NTk1NDRkMzE1OTZkNTY2YzRmNDQ1OTMxNGQ0NDY0Njg0ZjU3NGU2YTRlN2E2NzdhNWE0NzU1Nzc0ZjQ0NWE2OTRlNDQ0NTMzNGU1NDZiMzQ1YTU0NTE3YTU5NTQ0ZTZiNWE2YTU2NmE1OTMyNDU3OTVhNTQ2ODY4NGQ2YTZjNDE0ZDZhNTEzNDRlNTQ2NzdhNGQ1NzRlNmQ0ZDU0NDUzMDRkNTQ1NjZkNTk2YTQxMzU0ZDZhNjM3NzRlNDQ1MTMyNGU1NzU1MzI0ZTdhNTk3YTU5NTc0ZDMxNGY0NDQ1MzQ1YTU0NjczMTRlNDc1MTM0NTk1NzUyNmQ0ZTU0NDE3YTU5NmE2MzM1NGQ2YTZjNmI0ZjU0NTI2YzRlNmQ0OTc5NGU2YTQ5Nzc1YTY3M2QzZDIyMmMyMjYzNjg2MTY5NmU0OTQ0MjIzYTIyNGQ1MTNkM2QyMjJjMjI3NjY1NzI3MzY5NmY2ZTIyM2EzMTJjMjI3MzY5Njc2ZTYxNzQ3NTcyNjUyMjNhMjI1MjM5NDYyYjM0NTQ2MzUyNDE1YTM4NmQ3NzcxMzI0NTU5MzAzMTYzNTk2YzMzNzY2MjcxNmM0NjY1NzE3NjM4N2E3NjQ3NGE3NzVhNjgzMzU5NGQ0ZjU1NmI0MjM0NjQzNDUxNTc0ZTY2Mzc2NzQ0NjI2YzQ4NDgzMjU3NmI3MTYxNGE3NjYxNDg0NTc0NDM1NjYxNzA0OTcxMzM2NTM1NjU2MjM4NGU0MTc3M2QzZDIyN2Q=","signature":"","timestamp":5040,"status":"success","searchOrder":0,"hasScResults":true,"receivers":["000000000000000005008b60efce7fb25b140078645b71226cb9c644abc8a0d3"],"receiversShardIDs":[0],"operation":"transfer","function":"saveAttestation","isRelayed":true}`

	expectedRelayedTxIntra = `{"miniBlockHash":"2709174224d13e49fd76a70b48bd3db7838ca715bcfe09be59cef043241d7ef3","nonce":1196665,"round":50,"value":"0","receiver":"6572643134657961796672766c72687a66727767357a776c65756132356d6b7a676e6367676e33356e766336786876357978776d6c326573306633646874","sender":"657264316b376a3665776a736c61347a73677638763666366665336476726b677633643064396a6572637a773435687a6564687965643873683275333475","receiverShard":0,"senderShard":0,"gasPrice":1000000000,"gasLimit":15406000,"gasUsed":10556000,"fee":"2257820000000000","data":"cmVsYXllZFR4QDdiMjI2ZTZmNmU2MzY1MjIzYTMwMmMyMjc2NjE2Yzc1NjUyMjNhMzAyYzIyNzI2NTYzNjU2OTc2NjU3MjIyM2EyMjcyNmI2ZTUzNGE0NzdhMzQzNzY5NTM0ZTc5NGI0MzY0MmY1MDRmNzE3MDc1Nzc2YjU0Nzc2ODQ1MzQzMDZkN2E0NzZhNTg1YTUxNjg2ZTYyMmI3MjRkM2QyMjJjMjI3MzY1NmU2NDY1NzIyMjNhMjI3MjZiNmU1MzRhNDc3YTM0Mzc2OTUzNGU3OTRiNDM2NDJmNTA0ZjcxNzA3NTc3NmI1NDc3Njg0NTM0MzA2ZDdhNDc2YTU4NWE1MTY4NmU2MjJiNzI0ZDNkMjIyYzIyNjc2MTczNTA3MjY5NjM2NTIyM2EzMTMwMzAzMDMwMzAzMDMwMzAzMDJjMjI2NzYxNzM0YzY5NmQ2OTc0MjIzYTMxMzMzMjMzMzIzMDMwMzAyYzIyNjQ2MTc0NjEyMjNhMjI1NTMyNDYzMjVhNTU3NDZjNjU1NjVhNjg2MjQ4NTY2YzUxNDQ1OTc5NGU2YjU1MzI0ZDZiNDEzMjRkNmE1YTQ2NGU2YTQ5N2E0ZDU0NGQzNTRlNmE1NTMyNGQ1NDRkMzI0ZTZiNGQzMjUxNTQ2Mzc4NGQ3YTZiMzI1MTdhNjMzMDRlNmE1NTMzNGU0NDYzMzA0ZTdhNjczMjRlNTQ0ZDMzNGU3YTUxN2E0ZTdhNTkzMzRlNmI1NTdhNGQ0NDYzNzc0ZDdhNDk3YTRmNTQ2NDQyNGU3YTU5MzM0ZTU0NWE0NDRlNmE0NTMzNGU1NDU5MzI0ZTZhNjMzMzRlNTQ0ZDMwNGU3YTQ1N2E0ZTdhNTkzMzRlN2E0YTQxNGU2YTU1MzM0ZTQ0NTkzNDUxNDQ0ZDc3NGU3YTY3N2E0ZDU0NGQzMDRlNDQ1MTMyNGQ2YTRkNzg0ZTZhNTU3YTRkNDQ0ZDMxNGU2YTU5N2E0ZTU0NTE3YTRlNmE0NTdhNGU1NDRkMzE0ZTZhNDk3YTRkNTQ0ZDdhNGU0NDQ1MzA0ZDdhNTk3ODRkN2E2MzMyNGQ3YTRkMzA0ZTZhNGQ3YTRkNDQ0ZDc4NGU2YTU5N2E0ZTU0NTkzMDRlNDQ1NTdhNGU1NDRkMzE0ZTQ0NDk3YTRmNTQ1MTdhNGU0NDQ1MzI0ZTZhNTk3ODRkN2E2MzdhNGY1NTQxMzI0ZDZhNjMzMDRlNmE0ZTQxNGU2YTQ5MzI0ZDdhNGQ3ODRlN2E0NTMyNGQ1NDU5MzE0ZDdhNTU3YTRmNDQ0ZDdhNGQ3YTRkMzM0ZDZhNjM3OTRlN2E1OTdhNGU0NDRkNzc0ZDdhNDE3YTRlNTQ1YTQ0NGU2YTU5MzI1MjQ0NGQzMTRlN2E1OTMyNTI0NDRkMzI0ZTdhNTUzMjUxNTQ2MzMyNGU2YTUxN2E0ZTQ0NTk3YTRlN2E1MTdhNGY1NDVhNDI0ZTZhNGQ3YTRlNTQ1YTQyNGU3YTYzMzM1MTU0NGQzNDRlNmE1NTMzNGY0NDYzNzcyMjJjMjI2MzY4NjE2OTZlNDk0NDIyM2EyMjRkNTEzZDNkMjIyYzIyNzY2NTcyNzM2OTZmNmUyMjNhMzEyYzIyNzM2OTY3NmU2MTc0NzU3MjY1MjIzYTIyNzE2NjcwNGE0Nzc2NzM0NDQ0NDI1NTUxNGUyZjUyNTU0NzRmNTA1Mzc1NTIzMjQ4NGY0YTYxNGI3MDM4NDUzNjYzNGU1NDc3MzAzMzQzMzc2OTM0NTU3Nzc2MmY0YzU0NzM2ZDJiNmE3MDQyMzk3NTZjNDgzOTY2NTMyYjQ0NzE2MTcyNzE0ZjYyNDg0MTcwMzg2NjZkNzIzMDZhNDE1NTMxNzM2ZTM1NDE2NzNkM2QyMjdk","signature":"","timestamp":5040,"status":"success","searchOrder":0,"hasScResults":true,"receivers":["ae49d2246cf8ee248dc8a09dfcf3aaa6ec244f0844e349b31a35d94219dbfab3"],"receiversShardIDs":[0],"operation":"transfer","isRelayed":true}`
)

func TestRelayedTransactionGasUsedCrossShard(t *testing.T) {
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

func TestRelayedTransactionIntraShard(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	txHash := []byte("relayedTxIntra")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}

	scrHash1 := []byte("scrAtt")
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
		Nonce:    1196665,
		SndAddr:  []byte("erd1k7j6ewjsla4zsgv8v6f6fe3dvrkgv3d0d9jerczw45hzedhyed8sh2u34u"),
		RcvAddr:  []byte("erd14eyayfrvlrhzfrwg5zwleua25mkzgncggn35nvc6xhv5yxwml2es0f3dht"),
		GasLimit: 15406000,
		GasPrice: 1000000000,
		Data:     []byte("relayedTx@7b226e6f6e6365223a302c2276616c7565223a302c227265636569766572223a22726b6e534a477a343769534e794b43642f504f717075776b5477684534306d7a476a585a51686e622b724d3d222c2273656e646572223a22726b6e534a477a343769534e794b43642f504f717075776b5477684534306d7a476a585a51686e622b724d3d222c226761735072696365223a313030303030303030302c226761734c696d6974223a31333233323030302c2264617461223a22553246325a55746c65565a686248566c514459794e6b55324d6b41324d6a5a464e6a497a4d544d354e6a55324d544d324e6b4d32515463784d7a6b32517a63304e6a55334e4463304e7a67324e544d334e7a517a4e7a59334e6b557a4d4463774d7a497a4f5464424e7a59334e545a444e6a45334e5459324e6a63334e544d304e7a457a4e7a59334e7a4a414e6a55334e44593451444d774e7a677a4d544d304e4451324d6a4d784e6a557a4d444d314e6a597a4e54517a4e6a457a4e544d314e6a497a4d544d7a4e4445304d7a59784d7a63324d7a4d304e6a4d7a4d444d784e6a597a4e5459304e44557a4e544d314e44497a4f54517a4e4445324e6a59784d7a637a4f5541324d6a63304e6a4e414e6a49324d7a4d784e7a45324d5459314d7a557a4f444d7a4d7a4d334d6a63794e7a597a4e444d774d7a417a4e545a444e6a593252444d314e7a593252444d324e7a5532515463324e6a517a4e44597a4e7a517a4f545a424e6a4d7a4e545a424e7a633351544d344e6a55334f446377222c22636861696e4944223a224d513d3d222c2276657273696f6e223a312c227369676e6174757265223a227166704a47767344444255514e2f5255474f5053755232484f4a614b70384536634e54773033433769345577762f4c54736d2b6a704239756c483966532b44716172714f6248417038666d72306a415531736e3541673d3d227d"),
		Value:    big.NewInt(0),
	}

	scr1 := &smartContractResult.SmartContractResult{
		Nonce:          0,
		GasPrice:       1000000000,
		GasLimit:       12750000,
		SndAddr:        []byte("erd14eyayfrvlrhzfrwg5zwleua25mkzgncggn35nvc6xhv5yxwml2es0f3dht"),
		RcvAddr:        []byte("erd14eyayfrvlrhzfrwg5zwleua25mkzgncggn35nvc6xhv5yxwml2es0f3dht"),
		Data:           []byte("SaveKeyValue@626E62@626E6231396561366C6A71396C746574747865377437676E307032397A76756C61756667753471376772@657468@307831344462316530356635436135356231334143613763346330316635644535354239434166613739@627463@62633171616535383333727276343030356C666D35766D36756A7664346374396A63356A777A38657870"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
	}

	refundValueBig, _ := big.NewInt(0).SetString("48500000000000", 10)
	scrHash2 := []byte("scrWithRefund")
	scr2 := &smartContractResult.SmartContractResult{
		Nonce:          1,
		RcvAddr:        []byte("erd1k7j6ewjsla4zsgv8v6f6fe3dvrkgv3d0d9jerczw45hzedhyed8sh2u34u"),
		SndAddr:        []byte("erd14eyayfrvlrhzfrwg5zwleua25mkzgncggn35nvc6xhv5yxwml2es0f3dht"),
		PrevTxHash:     []byte("a98ee38f22153ae9fb497504b228077fb515502946b87c7d570852476ca3329b"),
		OriginalTxHash: txHash,
		ReturnMessage:  []byte("gas refund for relayer"),
		Value:          refundValueBig,
	}

	pool := &indexer.Pool{
		Txs: map[string]coreData.TransactionHandler{
			string(txHash): initialTx,
		},
		Scrs: map[string]coreData.TransactionHandler{
			string(scrHash1): scr1,
			string(scrHash2): scr2,
		},
	}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, []byte(expectedRelayedTxIntra), genericResponse.Docs[0].Source)
}
