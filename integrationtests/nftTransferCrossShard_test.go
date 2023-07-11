//go:build integrationtests

package integrationtests

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	indexerdata "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestNFTTransferCrossShardWithSCCall(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	txHash := []byte("nftTransferWithSCCall")
	header := &dataBlock.Header{
		Round:     50,
		ShardID:   0,
		TimeStamp: 5040,
	}

	scrHash1 := []byte("scrHash2")
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

	address1 := "erd1ef9xx3k3m89azf4c4xc98wpcdnx5h0cnxy6em47r6dc4alud0uwqx24f50"
	address2 := "erd1qqqqqqqqqqqqqpgq78y09lw93f3udvsplshdv2vk957l5vl70n4splrad2"
	scr1 := &smartContractResult.SmartContractResult{
		Nonce:          0,
		GasPrice:       1000000000,
		GasLimit:       148957500,
		SndAddr:        decodeAddress(address1),
		RcvAddr:        decodeAddress(address2),
		Data:           []byte("ESDTNFTTransfer@4c4b4641524d2d336634663962@016534@6f1e6f01bc7627f5ae@0801120a006f1e6f01bc7627f5ae227608b4ca051a2000000000000000000500f1c8f2fdc58a63c6b201fc2ed629962d3dfa33fe7ceb32003a4c0000000e4d45584641524d2d6239336536300000000000016ab5000000096f1e6f01bc7627f5ae0000000c4c4b4d45582d643163346162000000000000733b000000096e9018a1f0cc9a9aef@636f6d706f756e645265776172647350726f7879@000000000000000005004f79ec44bb13372b5ac9d996d749120f476427627ceb"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
	}

	tx := &transaction.Transaction{
		Nonce:    79,
		SndAddr:  decodeAddress(address1),
		RcvAddr:  decodeAddress(address1),
		GasLimit: 150000000,
		GasPrice: 1000000000,
		Data:     []byte("ESDTNFTTransfer@4c4b4641524d2d336634663962@016534@6f1e6f01bc7627f5ae@00000000000000000500f1c8f2fdc58a63c6b201fc2ed629962d3dfa33fe7ceb@636f6d706f756e645265776172647350726f7879@000000000000000005004f79ec44bb13372b5ac9d996d749120f476427627ceb"),
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
	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/nftTransferCrossShard/tx-nft-transfer-sc-call-source.json"),
		string(genericResponse.Docs[0].Source),
	)

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
	refundValueBig, _ := big.NewInt(0).SetString("101676480000000", 10)
	poolDstShard := &outport.TransactionPool{
		SmartContractResults: map[string]*outport.SCRInfo{
			hex.EncodeToString(scrHash1): {SmartContractResult: scr1, FeeInfo: &outport.FeeInfo{}},
			hex.EncodeToString(scrWithRefund): {SmartContractResult: &smartContractResult.SmartContractResult{
				SndAddr:        decodeAddress(address2),
				RcvAddr:        decodeAddress(address1),
				PrevTxHash:     []byte("f639cb7a0231191e04ec19dcb1359bd93a03fe8dc4a28a80d00835c5d1c988f8"),
				OriginalTxHash: txHash,
				Value:          refundValueBig,
				Data:           []byte("@6f6b@017d15@0000000e4d45584641524d2d6239336536300000000000017d15000000097045173cc97554b65d@0178af"),
			}, FeeInfo: &outport.FeeInfo{GasUsed: 139832352, Fee: big.NewInt(1802738520000000)}},
		},
	}

	err = esProc.SaveTransactions(createOutportBlockWithHeader(bodyDstShard, header, poolDstShard, nil, testNumOfShards))
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/nftTransferCrossShard/tx-nft-transfer-sc-call-after-refund.json"),
		string(genericResponse.Docs[0].Source),
	)

	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.OperationsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t,
		readExpectedResult("./testdata/nftTransferCrossShard/op-nft-transfer-sc-call-after-refund.json"),
		string(genericResponse.Docs[0].Source),
	)
}

func TestNFTTransferCrossShard(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
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

	address1 := "erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"
	address2 := "erd1qqqqqqqqqqqqqpgq57szwud2quysucrlq2e97ntdysdl7v4ejz3qn3njq4"
	scr2 := &smartContractResult.SmartContractResult{
		Nonce:          0,
		GasPrice:       1000000000,
		SndAddr:        decodeAddress(address1),
		RcvAddr:        decodeAddress(address2),
		Data:           []byte("ESDTNFTTransfer@434f4c45435449452d323663313838@01@01@08011202000122e50108011204434f4f4c1a20e0f3ecf555f63f2d101241dfc98b4614aff9284edd50b46a1c6e36b83558744d20c4132a2e516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a324368747470733a2f2f697066732e696f2f697066732f516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a3a41746167733a436f6f6c3b6d657461646174613a516d5869417850396e535948515954546143357358717a4d32645856334142516145355241725932777a4e686179"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
	}

	tx := &transaction.Transaction{
		Nonce:    79,
		SndAddr:  decodeAddress(address1),
		RcvAddr:  decodeAddress(address1),
		GasLimit: 5000000,
		GasPrice: 1000000000,
		Data:     []byte("ESDTNFTTransfer@536f6d657468696e672d616263646566@01@01@00000000000000000500a7a02771aa07090e607f02b25f4d6d241bff32b990a2"),
		Value:    big.NewInt(0),
	}

	txInfo := &outport.TxInfo{
		Transaction: tx,
		FeeInfo: &outport.FeeInfo{
			GasUsed:        963500,
			Fee:            big.NewInt(235850000000000),
			InitialPaidFee: big.NewInt(276215000000000),
		},
		ExecutionOrder: 0,
	}

	refundValueBig, _ := big.NewInt(0).SetString("40365000000000", 10)
	pool := &outport.TransactionPool{
		Transactions: map[string]*outport.TxInfo{
			hex.EncodeToString(txHash): txInfo,
		},
		SmartContractResults: map[string]*outport.SCRInfo{
			hex.EncodeToString(scrHash1): {SmartContractResult: &smartContractResult.SmartContractResult{
				Nonce:          80,
				Value:          refundValueBig,
				GasPrice:       1000000000,
				SndAddr:        decodeAddress(address1),
				RcvAddr:        decodeAddress(address1),
				Data:           []byte("@6f6b"),
				PrevTxHash:     txHash,
				OriginalTxHash: txHash,
			}, FeeInfo: &outport.FeeInfo{}},
			hex.EncodeToString(scrHash2): {SmartContractResult: scr2, FeeInfo: &outport.FeeInfo{}},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/nftTransferCrossShard/tx-nft-transfer.json"),
		string(genericResponse.Docs[0].Source),
	)

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
	poolDstShard := &outport.TransactionPool{
		SmartContractResults: map[string]*outport.SCRInfo{
			hex.EncodeToString(scrHash2): {SmartContractResult: scr2, FeeInfo: &outport.FeeInfo{}},
			hex.EncodeToString(scr3WithErrHash): {SmartContractResult: &smartContractResult.SmartContractResult{
				SndAddr:        decodeAddress(address2),
				RcvAddr:        decodeAddress(address1),
				PrevTxHash:     []byte("1546eb9970a6dc1710b6528274e75d5095c1349706f4ff70f52a1f58e1156316"),
				OriginalTxHash: txHash,
				Data:           []byte("ESDTNFTTransfer@434f4c45435449452d323663313838@01@01@08011202000122e50108011204434f4f4c1a20e0f3ecf555f63f2d101241dfc98b4614aff9284edd50b46a1c6e36b83558744d20c4132a2e516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a324368747470733a2f2f697066732e696f2f697066732f516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a3a41746167733a436f6f6c3b6d657461646174613a516d5869417850396e535948515954546143357358717a4d32645856334142516145355241725932777a4e686179@75736572206572726f72"),
			}, FeeInfo: &outport.FeeInfo{}},
		},
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString(scr3WithErrHash),
				Log: &transaction.Log{
					Address: decodeAddress(address1),
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address1),
							Identifier: []byte(core.InternalVMErrorsOperation),
						},
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(createOutportBlockWithHeader(bodyDstShard, header, poolDstShard, nil, testNumOfShards))
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/nftTransferCrossShard/tx-nft-transfer-failed-on-dst.json"),
		string(genericResponse.Docs[0].Source),
	)
}

func TestNFTTransferCrossShardImportDBScenarioFirstIndexDestinationAfterSource(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	txHash := []byte("nftTransferCross")
	scrHash1 := []byte("scrHashCross1")
	scrHash2 := []byte("scrHashCross2")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}

	address1 := "erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"
	address2 := "erd1qqqqqqqqqqqqqpgq57szwud2quysucrlq2e97ntdysdl7v4ejz3qn3njq4"
	scr2 := &smartContractResult.SmartContractResult{
		Nonce:          0,
		GasPrice:       1000000000,
		SndAddr:        decodeAddress(address1),
		RcvAddr:        decodeAddress(address2),
		Data:           []byte("ESDTNFTTransfer@434f4c4c454354494f4e2d323663313838@01@01@08011202000122e50108011204434f4f4c1a20e0f3ecf555f63f2d101241dfc98b4614aff9284edd50b46a1c6e36b83558744d20c4132a2e516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a324368747470733a2f2f697066732e696f2f697066732f516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a3a41746167733a436f6f6c3b6d657461646174613a516d5869417850396e535948515954546143357358717a4d32645856334142516145355241725932777a4e686179"),
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
	poolDstShard := &outport.TransactionPool{
		SmartContractResults: map[string]*outport.SCRInfo{
			hex.EncodeToString(scrHash2): {SmartContractResult: scr2, FeeInfo: &outport.FeeInfo{}},
			hex.EncodeToString(scr3WithErrHash): {SmartContractResult: &smartContractResult.SmartContractResult{
				SndAddr:        decodeAddress(address2),
				RcvAddr:        decodeAddress(address1),
				PrevTxHash:     []byte("1546eb9970a6dc1710b6528274e75d5095c1349706f4ff70f52a1f58e1156316"),
				OriginalTxHash: txHash,
				Data:           []byte("ESDTNFTTransfer@434f4c4c454354494f4e2d323663313838@01@01@08011202000122e50108011204434f4f4c1a20e0f3ecf555f63f2d101241dfc98b4614aff9284edd50b46a1c6e36b83558744d20c4132a2e516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a324368747470733a2f2f697066732e696f2f697066732f516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a3a41746167733a436f6f6c3b6d657461646174613a516d5869417850396e535948515954546143357358717a4d32645856334142516145355241725932777a4e686179@75736572206572726f72"),
			}, FeeInfo: &outport.FeeInfo{}},
		},
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString(scr3WithErrHash),
				Log: &transaction.Log{
					Address: decodeAddress(address1),
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address1),
							Identifier: []byte(core.SignalErrorOperation),
						},
					},
				},
			},
		},
	}

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(bodyDstShard, header, poolDstShard, nil, testNumOfShards))
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/nftTransferCrossShard/tx-with-status-only.json"),
		string(genericResponse.Docs[0].Source),
	)

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

	tx := &transaction.Transaction{
		Nonce:    79,
		SndAddr:  decodeAddress(address1),
		RcvAddr:  decodeAddress(address1),
		GasLimit: 5000000,
		GasPrice: 1000000000,
		Data:     []byte("ESDTNFTTransfer@434f4c4c454354494f4e2d323663313838@01@01@00000000000000000500a7a02771aa07090e607f02b25f4d6d241bff32b990a2"),
		Value:    big.NewInt(0),
	}

	txInfo := &outport.TxInfo{
		Transaction: tx,
		FeeInfo: &outport.FeeInfo{
			GasUsed:        963500,
			Fee:            big.NewInt(238820000000000),
			InitialPaidFee: big.NewInt(595490000000000),
		},
		ExecutionOrder: 0,
	}

	refundValueBig, _ := big.NewInt(0).SetString("40365000000000", 10)
	pool := &outport.TransactionPool{
		Transactions: map[string]*outport.TxInfo{
			hex.EncodeToString(txHash): txInfo,
		},
		SmartContractResults: map[string]*outport.SCRInfo{
			hex.EncodeToString(scrHash1): {SmartContractResult: &smartContractResult.SmartContractResult{
				Nonce:          80,
				Value:          refundValueBig,
				GasPrice:       1000000000,
				SndAddr:        decodeAddress(address1),
				RcvAddr:        decodeAddress(address1),
				Data:           []byte("@6f6b"),
				PrevTxHash:     txHash,
				OriginalTxHash: txHash,
			}, FeeInfo: &outport.FeeInfo{}},
			hex.EncodeToString(scrHash2): {SmartContractResult: scr2, FeeInfo: &outport.FeeInfo{}},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/nftTransferCrossShard/tx-complete-with-status.json"),
		string(genericResponse.Docs[0].Source),
	)
}
