package block

import (
	"encoding/hex"
	"math/big"
	"testing"

	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/stretchr/testify/require"
)

func TestGetMbsDetailsFromIntraShardMB_NilAndEmpty(t *testing.T) {
	t.Parallel()

	res := getMbsDetailsFromIntraShardMB(nil, &outport.TransactionPool{}, 0)
	require.NotNil(t, res)
	require.Empty(t, res)

	res = getMbsDetailsFromIntraShardMB([]*dataBlock.MiniBlock{}, &outport.TransactionPool{}, 5)
	require.NotNil(t, res)
	require.Empty(t, res)
}

func TestGetMbsDetailsFromIntraShardMB_SkipsPeerAndReceipt(t *testing.T) {
	t.Parallel()

	intraMbs := []*dataBlock.MiniBlock{
		{
			Type:            dataBlock.PeerBlock,
			SenderShardID:   0,
			ReceiverShardID: 0,
			TxHashes:        [][]byte{[]byte("peer-tx")},
		},
		{
			Type:            dataBlock.ReceiptBlock,
			SenderShardID:   0,
			ReceiverShardID: 1,
			TxHashes:        [][]byte{[]byte("receipt-tx")},
		},
		{
			Type:            dataBlock.TxBlock,
			SenderShardID:   0,
			ReceiverShardID: 0,
			TxHashes:        [][]byte{[]byte("tx1")},
		},
	}

	pool := &outport.TransactionPool{
		Transactions: map[string]*outport.TxInfo{
			hex.EncodeToString([]byte("tx1")): {
				Transaction:    &transaction.Transaction{},
				ExecutionOrder: 7,
			},
		},
	}

	res := getMbsDetailsFromIntraShardMB(intraMbs, pool, 2)
	require.Len(t, res, 1)
	require.Equal(t, 2+2, res[0].MBIndex)
	require.Equal(t, dataBlock.TxBlock.String(), res[0].Type)
	require.Equal(t, []int{7}, res[0].ExecutionOrderTxsIndices)
}

func TestGetMbsDetailsFromIntraShardMB_OffsetAndExecutionOrder(t *testing.T) {
	t.Parallel()

	txHash1 := []byte("intra-tx-1")
	txHash2 := []byte("intra-tx-2")
	missingTxHash := []byte("missing-tx")

	intraMbs := []*dataBlock.MiniBlock{
		{
			Type:            dataBlock.SmartContractResultBlock,
			SenderShardID:   1,
			ReceiverShardID: 1,
			TxHashes:        [][]byte{txHash1, txHash2, missingTxHash},
		},
	}

	pool := &outport.TransactionPool{
		SmartContractResults: map[string]*outport.SCRInfo{
			hex.EncodeToString(txHash1): {
				SmartContractResult: &smartContractResult.SmartContractResult{},
				ExecutionOrder:      4,
			},
			hex.EncodeToString(txHash2): {
				SmartContractResult: &smartContractResult.SmartContractResult{},
				ExecutionOrder:      5,
			},
		},
	}

	const offset = 3
	res := getMbsDetailsFromIntraShardMB(intraMbs, pool, offset)
	require.Len(t, res, 1)

	details := res[0]
	require.Equal(t, int32(0), details.IndexFirstProcessedTx)
	require.Equal(t, int32(2), details.IndexLastProcessedTx)
	require.Equal(t, uint32(1), details.SenderShardID)
	require.Equal(t, uint32(1), details.ReceiverShardID)
	require.Equal(t, offset, details.MBIndex)
	require.Equal(t, dataBlock.SmartContractResultBlock.String(), details.Type)
	require.Equal(t, dataBlock.Normal.String(), details.ProcessingType)
	require.Equal(t, []string{hex.EncodeToString(txHash1), hex.EncodeToString(txHash2), hex.EncodeToString(missingTxHash)}, details.TxsHashes)
	require.Equal(t, []int{4, 5, notFound}, details.ExecutionOrderTxsIndices)
}

func TestPrepareBlockForDB_IntraShardMBsFilteringAndOffset(t *testing.T) {
	t.Parallel()

	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.PubkeyConverterMock{})

	intraTxHash := []byte("intra-scr")
	header := &dataBlock.Header{
		MiniBlockHeaders: []dataBlock.MiniBlockHeader{
			{
				Hash:            []byte("mb0"),
				SenderShardID:   0,
				ReceiverShardID: 0,
				Type:            dataBlock.TxBlock,
			},
		},
	}
	headerBytes, _ := bp.marshalizer.Marshal(header)

	txHash := []byte("tx0")
	obh := &outport.OutportBlockWithHeader{
		Header: header,
		OutportBlock: &outport.OutportBlock{
			BlockData: &outport.BlockData{
				HeaderBytes: headerBytes,
				HeaderHash:  []byte("hash"),
				Body: &dataBlock.Body{
					MiniBlocks: []*dataBlock.MiniBlock{
						{
							Type:     dataBlock.TxBlock,
							TxHashes: [][]byte{txHash},
						},
					},
				},
				IntraShardMiniBlocks: []*dataBlock.MiniBlock{
					{
						Type:            dataBlock.SmartContractResultBlock,
						SenderShardID:   0,
						ReceiverShardID: 0,
						TxHashes:        [][]byte{intraTxHash},
					},
					{
						Type:     dataBlock.PeerBlock,
						TxHashes: [][]byte{[]byte("peer")},
					},
					{
						Type:     dataBlock.ReceiptBlock,
						TxHashes: [][]byte{[]byte("receipt")},
					},
				},
			},
			TransactionPool: &outport.TransactionPool{
				Transactions: map[string]*outport.TxInfo{
					hex.EncodeToString(txHash): {
						Transaction:    &transaction.Transaction{},
						ExecutionOrder: 1,
					},
				},
				SmartContractResults: map[string]*outport.SCRInfo{
					hex.EncodeToString(intraTxHash): {
						SmartContractResult: &smartContractResult.SmartContractResult{},
						ExecutionOrder:      2,
					},
				},
			},
			HeaderGasConsumption: &outport.HeaderGasConsumption{},
		},
	}

	blockResults, err := bp.PrepareBlockForDB(obh)
	require.NoError(t, err)
	require.Len(t, blockResults.Block.MiniBlocksDetails, 2)

	// first entry comes from the header miniblock, second from intra-shard with offset == 1
	require.Equal(t, 0, blockResults.Block.MiniBlocksDetails[0].MBIndex)
	require.Equal(t, 1, blockResults.Block.MiniBlocksDetails[1].MBIndex)
	require.Equal(t, dataBlock.SmartContractResultBlock.String(), blockResults.Block.MiniBlocksDetails[1].Type)
	require.Equal(t, []int{2}, blockResults.Block.MiniBlocksDetails[1].ExecutionOrderTxsIndices)
	require.Equal(t, int32(0), blockResults.Block.MiniBlocksDetails[1].IndexLastProcessedTx)

	// peer and receipt intra-shard miniblocks must not be indexed
	for _, details := range blockResults.Block.MiniBlocksDetails {
		require.NotEqual(t, dataBlock.PeerBlock.String(), details.Type)
		require.NotEqual(t, dataBlock.ReceiptBlock.String(), details.Type)
	}
}

func TestPrepareExecutionResult_WithIntraShardMBs(t *testing.T) {
	t.Parallel()

	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.PubkeyConverterMock{})

	executionResultHeaderHash := []byte("er-h1")
	mbTxHash := []byte("er-tx")
	intraTxHash := []byte("er-intra-scr")

	obh := &outport.OutportBlockWithHeader{
		Header: &dataBlock.HeaderV3{
			ExecutionResults: []*dataBlock.ExecutionResult{
				{
					BaseExecutionResult: &dataBlock.BaseExecutionResult{
						HeaderHash:  executionResultHeaderHash,
						HeaderNonce: 10,
						HeaderRound: 11,
						HeaderEpoch: 12,
					},
					MiniBlockHeaders: []dataBlock.MiniBlockHeader{
						{
							Hash:            []byte("mb-er"),
							SenderShardID:   2,
							ReceiverShardID: 2,
							Type:            dataBlock.TxBlock,
							TxCount:         1,
						},
					},
					AccumulatedFees: big.NewInt(1),
					DeveloperFees:   big.NewInt(2),
				},
			},
		},
		OutportBlock: &outport.OutportBlock{
			ShardID: 2,
			BlockData: &outport.BlockData{
				HeaderHash: []byte("notarized-hash"),
				Body:       &dataBlock.Body{},
				Results: map[string]*outport.ExecutionResultData{
					hex.EncodeToString(executionResultHeaderHash): {
						TimestampMs: 999,
						Body: &dataBlock.Body{
							MiniBlocks: []*dataBlock.MiniBlock{
								{
									SenderShardID:   2,
									ReceiverShardID: 2,
									Type:            dataBlock.TxBlock,
									TxHashes:        [][]byte{mbTxHash},
								},
							},
						},
						IntraShardMiniBlocks: []*dataBlock.MiniBlock{
							{
								Type:            dataBlock.SmartContractResultBlock,
								SenderShardID:   2,
								ReceiverShardID: 2,
								TxHashes:        [][]byte{intraTxHash},
							},
							{
								Type:     dataBlock.PeerBlock,
								TxHashes: [][]byte{[]byte("peer")},
							},
						},
						TransactionPool: &outport.TransactionPool{
							Transactions: map[string]*outport.TxInfo{
								hex.EncodeToString(mbTxHash): {
									Transaction:    &transaction.Transaction{},
									ExecutionOrder: 3,
								},
							},
							SmartContractResults: map[string]*outport.SCRInfo{
								hex.EncodeToString(intraTxHash): {
									SmartContractResult: &smartContractResult.SmartContractResult{},
									ExecutionOrder:      9,
								},
							},
						},
					},
				},
			},
		},
	}

	results, err := bp.PrepareBlockForDB(obh)
	require.NoError(t, err)
	require.Len(t, results.ExecutionResults, 1)

	executionResult := results.ExecutionResults[0]
	require.Len(t, executionResult.MiniBlocksDetails, 2)

	require.Equal(t, 0, executionResult.MiniBlocksDetails[0].MBIndex)
	require.Equal(t, dataBlock.TxBlock.String(), executionResult.MiniBlocksDetails[0].Type)
	require.Equal(t, []int{3}, executionResult.MiniBlocksDetails[0].ExecutionOrderTxsIndices)

	// intra-shard miniblock is appended with offset == number of header miniblocks (1)
	require.Equal(t, 1, executionResult.MiniBlocksDetails[1].MBIndex)
	require.Equal(t, dataBlock.SmartContractResultBlock.String(), executionResult.MiniBlocksDetails[1].Type)
	require.Equal(t, []string{hex.EncodeToString(intraTxHash)}, executionResult.MiniBlocksDetails[1].TxsHashes)
	require.Equal(t, []int{9}, executionResult.MiniBlocksDetails[1].ExecutionOrderTxsIndices)

	// peer intra-shard miniblock is skipped both in details and hashes
	require.Len(t, executionResult.MiniBlocksHashes, 2)
}

func TestPrepareExecutionResult_MetaExecutionResultWithIntraShardMBs(t *testing.T) {
	t.Parallel()

	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.PubkeyConverterMock{})

	executionResultHeaderHash := []byte("meta-er-h1")
	intraTxHash := []byte("meta-er-intra-tx")

	obh := &outport.OutportBlockWithHeader{
		Header: &dataBlock.MetaBlockV3{
			ExecutionResults: []*dataBlock.MetaExecutionResult{
				{
					ExecutionResult: &dataBlock.BaseMetaExecutionResult{
						BaseExecutionResult: &dataBlock.BaseExecutionResult{
							HeaderHash:  executionResultHeaderHash,
							HeaderNonce: 20,
							HeaderRound: 21,
							HeaderEpoch: 22,
						},
					},
					MiniBlockHeaders: []dataBlock.MiniBlockHeader{},
					AccumulatedFees:  big.NewInt(5),
					DeveloperFees:    big.NewInt(6),
				},
			},
		},
		OutportBlock: &outport.OutportBlock{
			ShardID: 4294967295,
			BlockData: &outport.BlockData{
				HeaderHash: []byte("meta-notarized-hash"),
				Body:       &dataBlock.Body{},
				Results: map[string]*outport.ExecutionResultData{
					hex.EncodeToString(executionResultHeaderHash): {
						TimestampMs: 111,
						Body: &dataBlock.Body{
							MiniBlocks: []*dataBlock.MiniBlock{},
						},
						IntraShardMiniBlocks: []*dataBlock.MiniBlock{
							{
								Type:            dataBlock.TxBlock,
								SenderShardID:   4294967295,
								ReceiverShardID: 4294967295,
								TxHashes:        [][]byte{intraTxHash},
							},
						},
						TransactionPool: &outport.TransactionPool{
							Transactions: map[string]*outport.TxInfo{
								hex.EncodeToString(intraTxHash): {
									Transaction:    &transaction.Transaction{},
									ExecutionOrder: 11,
								},
							},
						},
					},
				},
			},
		},
	}

	results, err := bp.PrepareBlockForDB(obh)
	require.NoError(t, err)
	require.Len(t, results.ExecutionResults, 1)

	executionResult := results.ExecutionResults[0]
	require.Len(t, executionResult.MiniBlocksDetails, 1)
	require.Equal(t, 0, executionResult.MiniBlocksDetails[0].MBIndex)
	require.Equal(t, dataBlock.TxBlock.String(), executionResult.MiniBlocksDetails[0].Type)
	require.Equal(t, []int{11}, executionResult.MiniBlocksDetails[0].ExecutionOrderTxsIndices)
	require.Equal(t, "5", executionResult.AccumulatedFees)
	require.Equal(t, "6", executionResult.DeveloperFees)
	require.Len(t, executionResult.MiniBlocksHashes, 1)
}

func TestPrepareExecutionResult_IntraShardMBTxNotFound(t *testing.T) {
	t.Parallel()

	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{}, &mock.PubkeyConverterMock{})

	executionResultHeaderHash := []byte("er-missing")
	unknownTxHash := []byte("unknown-tx")

	obh := &outport.OutportBlockWithHeader{
		Header: &dataBlock.HeaderV3{
			ExecutionResults: []*dataBlock.ExecutionResult{
				{
					BaseExecutionResult: &dataBlock.BaseExecutionResult{
						HeaderHash: executionResultHeaderHash,
					},
					MiniBlockHeaders: []dataBlock.MiniBlockHeader{},
					AccumulatedFees:  big.NewInt(0),
					DeveloperFees:    big.NewInt(0),
				},
			},
		},
		OutportBlock: &outport.OutportBlock{
			BlockData: &outport.BlockData{
				Body: &dataBlock.Body{},
				Results: map[string]*outport.ExecutionResultData{
					hex.EncodeToString(executionResultHeaderHash): {
						Body: &dataBlock.Body{},
						IntraShardMiniBlocks: []*dataBlock.MiniBlock{
							{
								Type:     dataBlock.TxBlock,
								TxHashes: [][]byte{unknownTxHash},
							},
						},
						TransactionPool: &outport.TransactionPool{},
					},
				},
			},
		},
	}

	results, err := bp.PrepareBlockForDB(obh)
	require.NoError(t, err)
	require.Len(t, results.ExecutionResults, 1)
	require.Len(t, results.ExecutionResults[0].MiniBlocksDetails, 1)
	require.Equal(t, []int{notFound}, results.ExecutionResults[0].MiniBlocksDetails[0].ExecutionOrderTxsIndices)
}
