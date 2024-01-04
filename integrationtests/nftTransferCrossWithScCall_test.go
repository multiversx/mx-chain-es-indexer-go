//go:build integrationtests

package integrationtests

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"

	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	indexerdata "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestNFTTransferCrossShardWithScCall(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
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

	address1 := "erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"
	address2 := "erd1qqqqqqqqqqqqqpgq57szwud2quysucrlq2e97ntdysdl7v4ejz3qn3njq4"
	scr2 := &smartContractResult.SmartContractResult{
		Nonce:          0,
		GasPrice:       1000000000,
		SndAddr:        decodeAddress(address1),
		RcvAddr:        decodeAddress(address2),
		Data:           []byte("ESDTNFTTransfer@4d45584641524d2d636362323532@078b@0347543e5b59c9be8670@000000000000000005005754e4f6ba0b94efd71a0e4dd4814ee24e5f75297ceb@636c61696d52657761726473"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
	}

	tx := &transaction.Transaction{
		Nonce:    79,
		SndAddr:  decodeAddress(address1),
		RcvAddr:  decodeAddress(address1),
		GasLimit: 5000000,
		GasPrice: 1000000000,
		Data:     []byte("ESDTNFTTransfer@4d45584641524d2d636362323532@078b@0347543e5b59c9be8670@00000000000000000500a7a02771aa07090e607f02b25f4d6d241bff32b990a2@636c61696d52657761726473"),
		Value:    big.NewInt(0),
	}

	txInfo := &outport.TxInfo{
		Transaction: tx,
		FeeInfo: &outport.FeeInfo{
			GasUsed:        5000000,
			Fee:            big.NewInt(595490000000000),
			InitialPaidFee: big.NewInt(595490000000000),
		},
		ExecutionOrder: 0,
	}

	pool := &outport.TransactionPool{
		Transactions: map[string]*outport.TxInfo{
			hex.EncodeToString(txHash): txInfo,
		},
		SmartContractResults: map[string]*outport.SCRInfo{
			hex.EncodeToString(scrHash2): {SmartContractResult: scr2, FeeInfo: &outport.FeeInfo{}},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t,
		readExpectedResult("./testdata/nftTransferCrossShardWithScCall/cross-shard-transfer-with-sc-call.json"),
		string(genericResponse.Docs[0].Source),
	)
}
