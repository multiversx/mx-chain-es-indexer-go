//go:build integrationtests

package integrationtests

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	indexerdata "github.com/ElrondNetwork/elastic-indexer-go/process/dataindexer"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestNFTTransferCrossShardWithScCall(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, shardCoordinator)
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

	tx := outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
		Nonce:    79,
		SndAddr:  []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
		RcvAddr:  []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
		GasLimit: 5000000,
		GasPrice: 1000000000,
		Data:     []byte("ESDTNFTTransfer@4d45584641524d2d636362323532@078b@0347543e5b59c9be8670@0801120b000347543e5b59c9be86702266088b0f1a20000000000000000005005754e4f6ba0b94efd71a0e4dd4814ee24e5f75297ceb32003a3d0000000701b6408636587c0000000000000410000000000000041001000000000a0347543e5b59c9be8670000000000000000a0347543e5b59c9be8670@636c61696d52657761726473"),
		Value:    big.NewInt(0),
	}, 5000000, big.NewInt(595490000000000))
	tx.SetInitialPaidFee(big.NewInt(595490000000000))

	// refundValueBig, _ := big.NewInt(0).SetString("40365000000000", 10)
	pool := &outport.Pool{
		Txs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(txHash): tx,
		},
		Scrs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(scrHash2): outport.NewTransactionHandlerWithGasAndFee(scr2, 0, big.NewInt(0)),
		},
	}
	err = esProc.SaveTransactions(body, header, pool, nil, false)
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t,
		readExpectedResult("./testdata/nftTransferCrossShardWithScCall/cross-shard-transfer-with-sc-call.json"),
		string(genericResponse.Docs[0].Source),
	)
}
