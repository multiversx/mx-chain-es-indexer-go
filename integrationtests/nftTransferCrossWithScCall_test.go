//go:build integrationtests

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
	expectedCrossShardTransferWithSCCall = `{"miniBlockHash":"99a07aab4f6722a1473b33bd7bb35e339c69339c400737b14a94ad8bceaa1734","nonce":79,"round":50,"value":"0","receiver":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","sender":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","receiverShard":0,"senderShard":0,"gasPrice":1000000000,"gasLimit":5000000,"gasUsed":5000000,"fee":"595490000000000","data":"RVNEVE5GVFRyYW5zZmVyQDRkNDU1ODQ2NDE1MjRkMmQ2MzYzNjIzMjM1MzJAMDc4YkAwMzQ3NTQzZTViNTljOWJlODY3MEAwODAxMTIwYjAwMDM0NzU0M2U1YjU5YzliZTg2NzAyMjY2MDg4YjBmMWEyMDAwMDAwMDAwMDAwMDAwMDAwNTAwNTc1NGU0ZjZiYTBiOTRlZmQ3MWEwZTRkZDQ4MTRlZTI0ZTVmNzUyOTdjZWIzMjAwM2EzZDAwMDAwMDA3MDFiNjQwODYzNjU4N2MwMDAwMDAwMDAwMDAwNDEwMDAwMDAwMDAwMDAwMDQxMDAxMDAwMDAwMDAwYTAzNDc1NDNlNWI1OWM5YmU4NjcwMDAwMDAwMDAwMDAwMDAwYTAzNDc1NDNlNWI1OWM5YmU4NjcwQDYzNmM2MTY5NmQ1MjY1Nzc2MTcyNjQ3Mw==","signature":"","timestamp":5040,"status":"success","searchOrder":0,"hasScResults":true,"tokens":["MEXFARM-ccb252-078b"],"esdtValues":["15482888667631250736752"],"receivers":["0801120b000347543e5b59c9be86702266088b0f1a20000000000000000005005754e4f6ba0b94efd71a0e4dd4814ee24e5f75297ceb32003a3d0000000701b6408636587c0000000000000410000000000000041001000000000a0347543e5b59c9be8670000000000000000a0347543e5b59c9be8670"],"receiversShardIDs":[0],"operation":"ESDTNFTTransfer"}`
)

func TestNFTTransferCrossShardWithScCall(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, shardCoordinator, feeComputer)
	require.Nil(t, err)

	txHash := []byte("nftTransferWithScCall")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}
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
		Data:           []byte("ESDTNFTTransfer@4d45584641524d2d636362323532@078b@0347543e5b59c9be8670@000000000000000005005754e4f6ba0b94efd71a0e4dd4814ee24e5f75297ceb@636c61696d52657761726473"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
	}

	// refundValueBig, _ := big.NewInt(0).SetString("40365000000000", 10)
	pool := &indexer.Pool{
		Txs: map[string]coreData.TransactionHandler{
			string(txHash): &transaction.Transaction{
				Nonce:    79,
				SndAddr:  []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				RcvAddr:  []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				GasLimit: 5000000,
				GasPrice: 1000000000,
				Data:     []byte("ESDTNFTTransfer@4d45584641524d2d636362323532@078b@0347543e5b59c9be8670@0801120b000347543e5b59c9be86702266088b0f1a20000000000000000005005754e4f6ba0b94efd71a0e4dd4814ee24e5f75297ceb32003a3d0000000701b6408636587c0000000000000410000000000000041001000000000a0347543e5b59c9be8670000000000000000a0347543e5b59c9be8670@636c61696d52657761726473"),
				Value:    big.NewInt(0),
			},
		},
		Scrs: map[string]coreData.TransactionHandler{
			string(scrHash2): scr2,
		},
	}
	err = esProc.SaveTransactions(body, header, pool, nil)
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, []byte(expectedCrossShardTransferWithSCCall), genericResponse.Docs[0].Source)
}
