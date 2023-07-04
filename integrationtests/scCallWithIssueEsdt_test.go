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
	"github.com/multiversx/mx-chain-core-go/data/vm"
	indexerData "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestScCallIntraShardWithIssueESDT(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	txHash := []byte("txWithScCall")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}
	scrWithIssueHash := []byte("scrWithIssue")
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
				ReceiverShardID: core.MetachainShardId,
				TxHashes:        [][]byte{scrWithIssueHash},
			},
		},
	}

	sndAddress := "erd148m2sx48mfm8322c2kpfmgj78g5j0x7r6z6y4z8j28qk45a74nwq5pq2ts"
	contractAddress := "erd1qqqqqqqqqqqqqpgqahumqen35dr9k4rmcnd70mqt5t4mt7ey4nwqwjme9g"
	tx := &transaction.Transaction{
		Nonce:    46,
		SndAddr:  decodeAddress(sndAddress),
		RcvAddr:  decodeAddress(contractAddress),
		GasLimit: 75_000_000,
		GasPrice: 1000000000,
		Data:     []byte("issueToken@4D79546573744E667464@544553544E4654"),
		Value:    big.NewInt(50000000000000000),
	}

	esdtSystemSC := "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"
	scrWithIssueESDT := &smartContractResult.SmartContractResult{
		Nonce:          0,
		SndAddr:        decodeAddress(contractAddress),
		RcvAddr:        decodeAddress(esdtSystemSC),
		OriginalTxHash: txHash,
		PrevTxHash:     txHash,
		Data:           []byte("issueNonFungible@4d79546573744e667464@544553544e4654@63616e467265657a65@74727565@63616e57697065@74727565@63616e5061757365@74727565@63616e4368616e67654f776e6572@66616c7365@63616e55706772616465@66616c7365@63616e4164645370656369616c526f6c6573@74727565@58f638"),
		Value:          big.NewInt(50000000000000000),
		CallType:       vm.AsynchronousCall,
	}
	txInfo := &outport.TxInfo{
		Transaction: tx,
		FeeInfo: &outport.FeeInfo{
			GasUsed:        75_000_000,
			Fee:            big.NewInt(867810000000000),
			InitialPaidFee: big.NewInt(867810000000000),
		},
	}

	scrInfoWithIssue := &outport.SCRInfo{SmartContractResult: scrWithIssueESDT, FeeInfo: &outport.FeeInfo{}}
	pool := &outport.TransactionPool{
		Transactions: map[string]*outport.TxInfo{
			hex.EncodeToString(txHash): txInfo,
		},
		SmartContractResults: map[string]*outport.SCRInfo{
			hex.EncodeToString(scrWithIssueHash): scrInfoWithIssue,
		},
	}

	// ############################
	// execute on the source shard
	// ############################

	header.ShardID = 0
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerData.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/scCallWithIssueEsdt/tx-after-execution-on-source-shard.json"),
		string(genericResponse.Docs[0].Source),
	)

	ids = []string{hex.EncodeToString(scrWithIssueHash)}
	err = esClient.DoMultiGet(ids, indexerData.OperationsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/scCallWithIssueEsdt/scr-with-issue-executed-on-source-shard.json"),
		string(genericResponse.Docs[0].Source),
	)

	// ############################
	// execute scr on the destination shard (metachain)
	// ############################

	scrWithCallBackHash := []byte("scrWithCallback")
	header.ShardID = core.MetachainShardId
	body = &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   0,
				ReceiverShardID: core.MetachainShardId,
				TxHashes:        [][]byte{scrWithIssueHash},
			},
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   core.MetachainShardId,
				ReceiverShardID: 0,
				TxHashes:        [][]byte{scrWithCallBackHash},
			},
		},
	}

	scrWithCallBack := &smartContractResult.SmartContractResult{
		Nonce:          0,
		Value:          big.NewInt(0),
		SndAddr:        decodeAddress(esdtSystemSC),
		RcvAddr:        decodeAddress(contractAddress),
		Data:           []byte("@00@544553544e46542d643964353463"),
		OriginalTxHash: txHash,
		PrevTxHash:     scrWithIssueHash,
		CallType:       vm.AsynchronousCallBack,
	}
	scrInfoWithCallBack := &outport.SCRInfo{SmartContractResult: scrWithCallBack, FeeInfo: &outport.FeeInfo{}}
	pool = &outport.TransactionPool{
		SmartContractResults: map[string]*outport.SCRInfo{
			hex.EncodeToString(scrWithIssueHash):    scrInfoWithIssue,
			hex.EncodeToString(scrWithCallBackHash): scrInfoWithCallBack,
		},
	}

	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids = []string{hex.EncodeToString(scrWithIssueHash), hex.EncodeToString(scrWithCallBackHash)}
	err = esClient.DoMultiGet(ids, indexerData.OperationsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/scCallWithIssueEsdt/scr-with-issue-executed-on-destination-shard.json"),
		string(genericResponse.Docs[0].Source),
	)
	require.JSONEq(t,
		readExpectedResult("./testdata/scCallWithIssueEsdt/scr-with-callback-executed-on-source.json"),
		string(genericResponse.Docs[1].Source),
	)

	// ############################
	// execute scr with callback on the destination shard (0)
	// ############################
	header.ShardID = 0
	body = &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   core.MetachainShardId,
				ReceiverShardID: 0,
				TxHashes:        [][]byte{scrWithCallBackHash},
			},
		},
	}
	pool = &outport.TransactionPool{
		SmartContractResults: map[string]*outport.SCRInfo{
			hex.EncodeToString(scrWithCallBackHash): scrInfoWithCallBack,
		},
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString(scrWithCallBackHash),
				Log: &transaction.Log{
					Address: decodeAddress(contractAddress),
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(contractAddress),
							Identifier: []byte(core.SignalErrorOperation),
						},
						{
							Address:    decodeAddress(contractAddress),
							Identifier: []byte(core.InternalVMErrorsOperation),
						},
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids = []string{hex.EncodeToString(txHash)}
	err = esClient.DoMultiGet(ids, indexerData.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/scCallWithIssueEsdt/tx-after-execution-of-callback-on-destination-shard.json"),
		string(genericResponse.Docs[0].Source),
	)

	ids = []string{hex.EncodeToString(txHash), hex.EncodeToString(scrWithCallBackHash)}
	err = esClient.DoMultiGet(ids, indexerData.OperationsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/scCallWithIssueEsdt/tx-in-op-index-execution-of-callback-on-destination-shard.json"),
		string(genericResponse.Docs[0].Source),
	)
	require.JSONEq(t,
		readExpectedResult("./testdata/scCallWithIssueEsdt/scr-with-callback-executed-on-destination-shard.json"),
		string(genericResponse.Docs[1].Source),
	)
}
