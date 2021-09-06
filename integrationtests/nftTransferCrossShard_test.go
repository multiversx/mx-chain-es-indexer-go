package integrationtests

import (
	"encoding/hex"
	"math/big"
	"testing"

	indexer2 "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/client"
	"github.com/ElrondNetwork/elastic-indexer-go/client/logging"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/stretchr/testify/require"
)

const expectedTxNFTTransfer = `{"miniBlockHash":"83c60064098aa89220b5adc9d71f22b489bfc78cb3dcb516381102d7fec959e8","nonce":79,"round":50,"value":"0","receiver":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","sender":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","receiverShard":0,"senderShard":0,"gasPrice":1000000000,"gasLimit":5000000,"gasUsed":963500,"fee":"232880000000000","data":"RVNEVE5GVFRyYW5zZmVyQDQzNGY0YzQ1NDM1NDQ5NDUyZDMyMzY2MzMxMzgzOEAwMUAwMUAwMDAwMDAwMDAwMDAwMDAwMDUwMGE3YTAyNzcxYWEwNzA5MGU2MDdmMDJiMjVmNGQ2ZDI0MWJmZjMyYjk5MGEy","signature":"","timestamp":5040,"status":"success","searchOrder":0,"hasScResults":true}`
const expectedTxNFTTransferFailOnDestination = `{"miniBlockHash":"83c60064098aa89220b5adc9d71f22b489bfc78cb3dcb516381102d7fec959e8","nonce":79,"round":50,"value":"0","receiver":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","sender":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","receiverShard":0,"senderShard":0,"gasPrice":1000000000,"gasLimit":5000000,"gasUsed":963500,"fee":"232880000000000","data":"RVNEVE5GVFRyYW5zZmVyQDQzNGY0YzQ1NDM1NDQ5NDUyZDMyMzY2MzMxMzgzOEAwMUAwMUAwMDAwMDAwMDAwMDAwMDAwMDUwMGE3YTAyNzcxYWEwNzA5MGU2MDdmMDJiMjVmNGQ2ZDI0MWJmZjMyYjk5MGEy","signature":"","timestamp":5040,"status":"fail","searchOrder":0,"hasScResults":true}`

func TestNFTTransferCrossShard(t *testing.T) {
	setLogLevelDebug()

	esClient, err := client.NewElasticClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
		Logger:    &logging.CustomLogger{},
	})
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	txHash := []byte("nftTransfer")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}
	scrHash1 := []byte("scrHash1")
	scrHash2 := []byte("scrHash2")
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

	scr2 := &smartContractResult.SmartContractResult{
		Nonce:          0,
		GasPrice:       1000000000,
		SndAddr:        []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
		RcvAddr:        []byte("erd1qqqqqqqqqqqqqpgq57szwud2quysucrlq2e97ntdysdl7v4ejz3qn3njq4"),
		Data:           []byte("ESDTNFTTransfer@434f4c45435449452d323663313838@01@01@08011202000122e50108011204434f4f4c1a20e0f3ecf555f63f2d101241dfc98b4614aff9284edd50b46a1c6e36b83558744d20c4132a2e516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a324368747470733a2f2f697066732e696f2f697066732f516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a3a41746167733a436f6f6c3b6d657461646174613a516d5869417850396e535948515954546143357358717a4d32645856334142516145355241725932777a4e686179"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
	}

	refundValueBig, _ := big.NewInt(0).SetString("40365000000000", 10)
	pool := &indexer.Pool{
		Txs: map[string]coreData.TransactionHandler{
			string(txHash): &transaction.Transaction{
				Nonce:    79,
				SndAddr:  []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				RcvAddr:  []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				GasLimit: 5000000,
				GasPrice: 1000000000,
				Data:     []byte("ESDTNFTTransfer@434f4c45435449452d323663313838@01@01@00000000000000000500a7a02771aa07090e607f02b25f4d6d241bff32b990a2"),
				Value:    big.NewInt(0),
			},
		},
		Scrs: map[string]coreData.TransactionHandler{
			string(scrHash1): &smartContractResult.SmartContractResult{
				Nonce:          80,
				Value:          refundValueBig,
				GasPrice:       1000000000,
				SndAddr:        []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				RcvAddr:        []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				Data:           []byte("@6f6b"),
				PrevTxHash:     txHash,
				OriginalTxHash: txHash,
			},
			string(scrHash2): scr2,
		},
	}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexer2.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, expectedTxNFTTransfer, string(genericResponse.Docs[0].Source))

	// EXECUTE transfer on the destination shard
	bodyDstShard := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   0,
				ReceiverShardID: 1,
				TxHashes:        [][]byte{scrHash2},
			},
		},
	}
	scr3WithErrHash := []byte("scrWithError")
	poolDstShard := &indexer.Pool{
		Scrs: map[string]coreData.TransactionHandler{
			string(scrHash2): scr2,
			string(scr3WithErrHash): &smartContractResult.SmartContractResult{
				SndAddr:        []byte("erd1qqqqqqqqqqqqqpgq57szwud2quysucrlq2e97ntdysdl7v4ejz3qn3njq4"),
				RcvAddr:        []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				PrevTxHash:     []byte("1546eb9970a6dc1710b6528274e75d5095c1349706f4ff70f52a1f58e1156316"),
				OriginalTxHash: txHash,
				Data:           []byte("ESDTNFTTransfer@434f4c45435449452d323663313838@01@01@08011202000122e50108011204434f4f4c1a20e0f3ecf555f63f2d101241dfc98b4614aff9284edd50b46a1c6e36b83558744d20c4132a2e516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a324368747470733a2f2f697066732e696f2f697066732f516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a3a41746167733a436f6f6c3b6d657461646174613a516d5869417850396e535948515954546143357358717a4d32645856334142516145355241725932777a4e686179@75736572206572726f72"),
			},
		},
	}

	err = esProc.SaveTransactions(bodyDstShard, header, poolDstShard)
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexer2.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, expectedTxNFTTransferFailOnDestination, string(genericResponse.Docs[0].Source))
}

const (
	txWithOnlyStatus     = `{"miniBlockHash":"","nonce":0,"round":0,"value":"","receiver":"","sender":"","receiverShard":0,"senderShard":0,"gasPrice":0,"gasLimit":0,"gasUsed":0,"fee":"","data":null,"signature":"","timestamp":0,"status":"fail","searchOrder":0}`
	txCompleteWithStatus = `{"receiver":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","data":"RVNEVE5GVFRyYW5zZmVyQDQzNGY0YzQ1NDM1NDQ5NDUyZDMyMzY2MzMxMzgzOEAwMUAwMUAwMDAwMDAwMDAwMDAwMDAwMDUwMGE3YTAyNzcxYWEwNzA5MGU2MDdmMDJiMjVmNGQ2ZDI0MWJmZjMyYjk5MGEy","signature":"","fee":"232880000000000","nonce":79,"gasLimit":5000000,"gasUsed":963500,"miniBlockHash":"db7161a83f08489cba131e55f042536ee49116b622e33e70335a13e51a6c268c","round":50,"hasScResults":true,"sender":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","receiverShard":0,"senderShard":0,"value":"0","gasPrice":1000000000,"timestamp":5040,"status":"fail","searchOrder":0}`
)

func TestNFTTransferCrossShardImportDBScenarioFirstIndexDestinationAfterSource(t *testing.T) {
	setLogLevelDebug()

	esClient, err := client.NewElasticClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
		Logger:    &logging.CustomLogger{},
	})
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	txHash := []byte("nftTransferCross")
	scrHash1 := []byte("scrHashCross1")
	scrHash2 := []byte("scrHashCross2")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}

	scr2 := &smartContractResult.SmartContractResult{
		Nonce:          0,
		GasPrice:       1000000000,
		SndAddr:        []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
		RcvAddr:        []byte("erd1qqqqqqqqqqqqqpgq57szwud2quysucrlq2e97ntdysdl7v4ejz3qn3njq4"),
		Data:           []byte("ESDTNFTTransfer@434f4c45435449452d323663313838@01@01@08011202000122e50108011204434f4f4c1a20e0f3ecf555f63f2d101241dfc98b4614aff9284edd50b46a1c6e36b83558744d20c4132a2e516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a324368747470733a2f2f697066732e696f2f697066732f516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a3a41746167733a436f6f6c3b6d657461646174613a516d5869417850396e535948515954546143357358717a4d32645856334142516145355241725932777a4e686179"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
	}

	// EXECUTE transfer on the destination shard
	bodyDstShard := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   0,
				ReceiverShardID: 1,
				TxHashes:        [][]byte{scrHash2},
			},
		},
	}
	scr3WithErrHash := []byte("scrWithError")
	poolDstShard := &indexer.Pool{
		Scrs: map[string]coreData.TransactionHandler{
			string(scrHash2): scr2,
			string(scr3WithErrHash): &smartContractResult.SmartContractResult{
				SndAddr:        []byte("erd1qqqqqqqqqqqqqpgq57szwud2quysucrlq2e97ntdysdl7v4ejz3qn3njq4"),
				RcvAddr:        []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				PrevTxHash:     []byte("1546eb9970a6dc1710b6528274e75d5095c1349706f4ff70f52a1f58e1156316"),
				OriginalTxHash: txHash,
				Data:           []byte("ESDTNFTTransfer@434f4c45435449452d323663313838@01@01@08011202000122e50108011204434f4f4c1a20e0f3ecf555f63f2d101241dfc98b4614aff9284edd50b46a1c6e36b83558744d20c4132a2e516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a324368747470733a2f2f697066732e696f2f697066732f516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a3a41746167733a436f6f6c3b6d657461646174613a516d5869417850396e535948515954546143357358717a4d32645856334142516145355241725932777a4e686179@75736572206572726f72"),
			},
		},
	}

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esProc.SaveTransactions(bodyDstShard, header, poolDstShard)
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexer2.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, txWithOnlyStatus, string(genericResponse.Docs[0].Source))

	// execute on source

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

	refundValueBig, _ := big.NewInt(0).SetString("40365000000000", 10)
	pool := &indexer.Pool{
		Txs: map[string]coreData.TransactionHandler{
			string(txHash): &transaction.Transaction{
				Nonce:    79,
				SndAddr:  []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				RcvAddr:  []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				GasLimit: 5000000,
				GasPrice: 1000000000,
				Data:     []byte("ESDTNFTTransfer@434f4c45435449452d323663313838@01@01@00000000000000000500a7a02771aa07090e607f02b25f4d6d241bff32b990a2"),
				Value:    big.NewInt(0),
			},
		},
		Scrs: map[string]coreData.TransactionHandler{
			string(scrHash1): &smartContractResult.SmartContractResult{
				Nonce:          80,
				Value:          refundValueBig,
				GasPrice:       1000000000,
				SndAddr:        []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				RcvAddr:        []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				Data:           []byte("@6f6b"),
				PrevTxHash:     txHash,
				OriginalTxHash: txHash,
			},
			string(scrHash2): scr2,
		},
	}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexer2.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, txCompleteWithStatus, string(genericResponse.Docs[0].Source))
}
