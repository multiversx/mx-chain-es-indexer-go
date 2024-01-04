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

func TestMultiTransferCrossShardAndScCallErrorSignaledBySC(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
		ShardID:   0,
	}

	txHash, scrHash1, scrHash2 := []byte("multiTransferWithScCall"), []byte("scrMultiTransfer"), []byte("scrMultiTransferReverse")
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

	address1 := "erd1ju8pkvg57cwdmjsjx58jlmnuf4l9yspstrhr9tgsrt98n9edpm2qtlgy99"
	address2 := "erd1qqqqqqqqqqqqqpgqa0fsfshnff4n76jhcye6k7uvd7qacsq42jpsp6shh2"

	// process transaction on shard 0
	scr1 := &smartContractResult.SmartContractResult{
		Nonce:          0,
		GasPrice:       1000000000,
		GasLimit:       148957500,
		SndAddr:        decodeAddress(address1),
		RcvAddr:        decodeAddress(address2),
		Data:           []byte("MultiESDTNFTTransfer@02@5745474c442d626434643739@00@38e62046fb1a0000@584d45582d666461333535@07@0801120c00048907e58284c28e898e2922520807120a4d45582d3435356335371a20000000000000000005007afb2c871d1647372fd53a9eb3e53e5a8ec9251cb05532003a1e0000000a4d45582d343535633537000000000000000000000000000008e8@6164644c697175696469747950726f7879@00000000000000000500ebd304c2f34a6b3f6a57c133ab7b8c6f81dc40155483@38d78f595785c000@0487deac313c6f6b111906"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
	}

	tx := &transaction.Transaction{
		Nonce:    79,
		SndAddr:  decodeAddress(address1),
		RcvAddr:  decodeAddress(address1),
		GasLimit: 150000000,
		GasPrice: 1000000000,
		Data:     []byte("MultiESDTNFTTransfer@000000000000000005005ebeb3515cb42056a81d42adaf756a3f63a360bfb055@02@5745474c442d626434643739@@38e62046fb1a0000@584d45582d666461333535@07@048907e58284c28e898e29@6164644c697175696469747950726f7879@00000000000000000500ebd304c2f34a6b3f6a57c133ab7b8c6f81dc40155483@38d78f595785c000@0487deac313c6f6b111906"),
		Value:    big.NewInt(0),
	}

	txInfo := &outport.TxInfo{
		Transaction: tx,
		FeeInfo: &outport.FeeInfo{
			GasUsed:        150000000,
			Fee:            big.NewInt(1904415000000000),
			InitialPaidFee: big.NewInt(1904415000000000),
		},
		ExecutionOrder: 0,
	}

	pool := &outport.TransactionPool{
		Transactions: map[string]*outport.TxInfo{
			hex.EncodeToString(txHash): txInfo,
		},
		SmartContractResults: map[string]*outport.SCRInfo{
			hex.EncodeToString(scrHash1): {SmartContractResult: scr1, FeeInfo: &outport.FeeInfo{}},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)
	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/multiTransferWithScCallAndErrorSignaledBySC/transaction-executed-on-source.json"),
		string(genericResponse.Docs[0].Source),
	)

	// process SCR on shard 1
	header = &dataBlock.Header{
		Round:     52,
		TimeStamp: 5050,
		ShardID:   1,
	}
	body = &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   0,
				ReceiverShardID: 1,
				TxHashes:        [][]byte{scrHash1},
			},
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   1,
				ReceiverShardID: 0,
				TxHashes:        [][]byte{scrHash2},
			},
		},
	}

	scr2 := &smartContractResult.SmartContractResult{
		OriginalTxHash: txHash,
		SndAddr:        decodeAddress(address2),
		RcvAddr:        decodeAddress(address1),
		Data:           []byte("MultiESDTNFTTransfer@02@5745474c442d626434643739@00@38e62046fb1a0000@584d45582d666461333535@07@0801120c00048907e58284c28e898e2922520807120a4d45582d3435356335371a20000000000000000005007afb2c871d1647372fd53a9eb3e53e5a8ec9251cb05532003a1e0000000a4d45582d343535633537000000000000000000000000000008e8@657865637574696f6e206661696c6564"),
		ReturnMessage:  []byte("error signalled by smartcontract"),
	}

	pool = &outport.TransactionPool{
		SmartContractResults: map[string]*outport.SCRInfo{
			hex.EncodeToString(scrHash1): {SmartContractResult: scr1, FeeInfo: &outport.FeeInfo{}},
			hex.EncodeToString(scrHash2): {SmartContractResult: scr2, FeeInfo: &outport.FeeInfo{}},
		},
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString(scrHash1),
				Log: &transaction.Log{
					Address: decodeAddress(address2),
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address2),
							Identifier: []byte("signalError"),
						},
						{
							Address:    decodeAddress(address1),
							Identifier: []byte("internalVMErrors"),
						},
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids = []string{hex.EncodeToString(txHash)}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/multiTransferWithScCallAndErrorSignaledBySC/transaction-after-execution-of-scr-dst-shard.json"),
		string(genericResponse.Docs[0].Source),
	)
}
