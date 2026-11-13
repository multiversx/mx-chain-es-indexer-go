package dataprocessor

import (
	"encoding/hex"
	"errors"
	"testing"

	coreData "github.com/multiversx/mx-chain-core-go/data"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/outputs"
	"github.com/stretchr/testify/require"
)

type transactionsHandlerStub struct {
	PrepareTransactionsForDatabaseCalled func(miniBlocks []*dataBlock.MiniBlock, headerData *data.HeaderData, pool *outport.TransactionPool, isImportDB bool) *data.PreparedResults
	GetHexEncodedHashesForRemoveCalled   func(headerData *data.HeaderData, body *dataBlock.Body) ([]string, []string)
}

func (ths *transactionsHandlerStub) PrepareTransactionsForDatabase(miniBlocks []*dataBlock.MiniBlock, headerData *data.HeaderData, pool *outport.TransactionPool, isImportDB bool) *data.PreparedResults {
	if ths.PrepareTransactionsForDatabaseCalled != nil {
		return ths.PrepareTransactionsForDatabaseCalled(miniBlocks, headerData, pool, isImportDB)
	}

	return &data.PreparedResults{}
}

func (ths *transactionsHandlerStub) GetHexEncodedHashesForRemove(headerData *data.HeaderData, body *dataBlock.Body) ([]string, []string) {
	if ths.GetHexEncodedHashesForRemoveCalled != nil {
		return ths.GetHexEncodedHashesForRemoveCalled(headerData, body)
	}

	return nil, nil
}

type logsAndEventsHandlerStub struct {
	ExtractDataFromLogsCalled func(logsAndEvents []*transaction.LogData, preparedResults *data.PreparedResults, shardID uint32, numOfShards uint32, timestampMs uint64) *data.PreparedLogsResults
}

func (lhs *logsAndEventsHandlerStub) ExtractDataFromLogs(logsAndEvents []*transaction.LogData, preparedResults *data.PreparedResults, shardID uint32, numOfShards uint32, timestampMs uint64) *data.PreparedLogsResults {
	if lhs.ExtractDataFromLogsCalled != nil {
		return lhs.ExtractDataFromLogsCalled(logsAndEvents, preparedResults, shardID, numOfShards, timestampMs)
	}

	return &data.PreparedLogsResults{}
}

type operationsHandlerStub struct {
	ProcessTransactionsAndSCRsCalled func(txs []*data.Transaction, scrs []*data.ScResult, isImportDB bool, shardID uint32) ([]*data.Transaction, []*data.ScResult)
}

func (ohs *operationsHandlerStub) ProcessTransactionsAndSCRs(txs []*data.Transaction, scrs []*data.ScResult, isImportDB bool, shardID uint32) ([]*data.Transaction, []*data.ScResult) {
	if ohs.ProcessTransactionsAndSCRsCalled != nil {
		return ohs.ProcessTransactionsAndSCRsCalled(txs, scrs, isImportDB, shardID)
	}

	return txs, scrs
}

func createMockArgDataProcessor(providedOutputs ...outputs.OutputWithPolicy) ArgDataProcessor {
	if len(providedOutputs) == 0 {
		providedOutputs = []outputs.OutputWithPolicy{{Name: "first", Output: &mock.OutputStub{}}}
	}

	return ArgDataProcessor{
		TransactionsProc:  &transactionsHandlerStub{},
		LogsAndEventsProc: &logsAndEventsHandlerStub{},
		OperationsProc:    &operationsHandlerStub{},
		Outputs:           providedOutputs,
	}
}

func createOutportBlock(shardID uint32, headerHash []byte) *outport.OutportBlockWithHeader {
	return &outport.OutportBlockWithHeader{
		Header: &dataBlock.Header{ShardID: shardID, Round: 10},
		OutportBlock: &outport.OutportBlock{
			BlockData: &outport.BlockData{
				HeaderHash: headerHash,
				Body:       &dataBlock.Body{},
			},
			TransactionPool: &outport.TransactionPool{},
		},
	}
}

func TestNewDataProcessor(t *testing.T) {
	t.Parallel()

	t.Run("nil transactions handler", func(t *testing.T) {
		t.Parallel()

		args := createMockArgDataProcessor()
		args.TransactionsProc = nil
		dp, err := NewDataProcessor(args)
		require.Nil(t, dp)
		require.Equal(t, ErrNilTransactionsHandler, err)
	})
	t.Run("nil logs and events handler", func(t *testing.T) {
		t.Parallel()

		args := createMockArgDataProcessor()
		args.LogsAndEventsProc = nil
		dp, err := NewDataProcessor(args)
		require.Nil(t, dp)
		require.Equal(t, ErrNilLogsAndEventsHandler, err)
	})
	t.Run("nil operations handler", func(t *testing.T) {
		t.Parallel()

		args := createMockArgDataProcessor()
		args.OperationsProc = nil
		dp, err := NewDataProcessor(args)
		require.Nil(t, dp)
		require.Equal(t, ErrNilOperationsHandler, err)
	})
	t.Run("no outputs", func(t *testing.T) {
		t.Parallel()

		args := createMockArgDataProcessor()
		args.Outputs = nil
		dp, err := NewDataProcessor(args)
		require.Nil(t, dp)
		require.Equal(t, ErrNoOutputs, err)
	})
	t.Run("nil output", func(t *testing.T) {
		t.Parallel()

		args := createMockArgDataProcessor(outputs.OutputWithPolicy{Name: "missing"})
		dp, err := NewDataProcessor(args)
		require.Nil(t, dp)
		require.True(t, errors.Is(err, ErrNilOutput))
		require.Contains(t, err.Error(), "missing")
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dp, err := NewDataProcessor(createMockArgDataProcessor())
		require.Nil(t, err)
		require.False(t, dp.IsInterfaceNil())
	})
}

func TestDataProcessor_SaveTransactionsShouldPrepareOnceAndCallTheOutputsInOrder(t *testing.T) {
	t.Parallel()

	prepareCalls := 0
	calls := make([]string, 0)
	var blockFirst, blockSecond *data.PreparedBlock

	args := createMockArgDataProcessor(
		outputs.OutputWithPolicy{Name: "first", Output: &mock.OutputStub{
			SaveTransactionsCalled: func(block *data.PreparedBlock) error {
				calls = append(calls, "first")
				blockFirst = block
				return nil
			},
		}},
		outputs.OutputWithPolicy{Name: "second", Output: &mock.OutputStub{
			SaveTransactionsCalled: func(block *data.PreparedBlock) error {
				calls = append(calls, "second")
				blockSecond = block
				return nil
			},
		}},
	)
	args.TransactionsProc = &transactionsHandlerStub{
		PrepareTransactionsForDatabaseCalled: func(_ []*dataBlock.MiniBlock, _ *data.HeaderData, _ *outport.TransactionPool, _ bool) *data.PreparedResults {
			prepareCalls++
			return &data.PreparedResults{}
		},
	}
	dp, _ := NewDataProcessor(args)

	obh := createOutportBlock(1, []byte("hash"))
	err := dp.SaveTransactions(obh)
	require.Nil(t, err)

	require.Equal(t, 1, prepareCalls)
	require.Equal(t, []string{"first", "second"}, calls)
	// both outputs receive the same prepared data
	require.NotNil(t, blockFirst)
	require.True(t, blockFirst == blockSecond)
	require.True(t, blockFirst.OutportBlock == obh)
}

func TestDataProcessor_OutputError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("output error")

	t.Run("the error is returned and the next outputs are not called", func(t *testing.T) {
		t.Parallel()

		secondCalled := false
		dp, _ := NewDataProcessor(createMockArgDataProcessor(
			outputs.OutputWithPolicy{Name: "first", Output: &mock.OutputStub{
				SaveTransactionsCalled: func(_ *data.PreparedBlock) error {
					return expectedErr
				},
			}},
			outputs.OutputWithPolicy{Name: "second", Output: &mock.OutputStub{
				SaveTransactionsCalled: func(_ *data.PreparedBlock) error {
					secondCalled = true
					return nil
				},
			}},
		))

		err := dp.SaveTransactions(createOutportBlock(1, []byte("hash")))
		require.True(t, errors.Is(err, expectedErr))
		require.Contains(t, err.Error(), "first")
		require.False(t, secondCalled)
	})
	t.Run("continue on error: the error is only logged and the next outputs are called", func(t *testing.T) {
		t.Parallel()

		secondCalled := false
		dp, _ := NewDataProcessor(createMockArgDataProcessor(
			outputs.OutputWithPolicy{Name: "first", ContinueOnError: true, Output: &mock.OutputStub{
				SaveTransactionsCalled: func(_ *data.PreparedBlock) error {
					return expectedErr
				},
			}},
			outputs.OutputWithPolicy{Name: "second", Output: &mock.OutputStub{
				SaveTransactionsCalled: func(_ *data.PreparedBlock) error {
					secondCalled = true
					return nil
				},
			}},
		))

		err := dp.SaveTransactions(createOutportBlock(1, []byte("hash")))
		require.Nil(t, err)
		require.True(t, secondCalled)
	})
}

func TestDataProcessor_SaveTransactionsSameBlockAgainShouldSkipTheOutputsThatSavedIt(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("second output error")
	firstCalls, secondCalls := 0, 0
	secondShouldFail := true

	dp, _ := NewDataProcessor(createMockArgDataProcessor(
		outputs.OutputWithPolicy{Name: "first", Output: &mock.OutputStub{
			SaveTransactionsCalled: func(_ *data.PreparedBlock) error {
				firstCalls++
				return nil
			},
		}},
		outputs.OutputWithPolicy{Name: "second", Output: &mock.OutputStub{
			SaveTransactionsCalled: func(_ *data.PreparedBlock) error {
				secondCalls++
				if secondShouldFail {
					return expectedErr
				}
				return nil
			},
		}},
	))

	// the second output fails: the node will send the block again
	err := dp.SaveTransactions(createOutportBlock(1, []byte("hash1")))
	require.True(t, errors.Is(err, expectedErr))
	require.Equal(t, 1, firstCalls)
	require.Equal(t, 1, secondCalls)

	// the same block is received again: the first output already saved it, only the second output is called
	secondShouldFail = false
	err = dp.SaveTransactions(createOutportBlock(1, []byte("hash1")))
	require.Nil(t, err)
	require.Equal(t, 1, firstCalls)
	require.Equal(t, 2, secondCalls)

	// the same block one more time: it was saved by all the outputs
	err = dp.SaveTransactions(createOutportBlock(1, []byte("hash1")))
	require.Nil(t, err)
	require.Equal(t, 1, firstCalls)
	require.Equal(t, 2, secondCalls)

	// the same hash from another shard is another block
	err = dp.SaveTransactions(createOutportBlock(2, []byte("hash1")))
	require.Nil(t, err)
	require.Equal(t, 2, firstCalls)
	require.Equal(t, 3, secondCalls)

	// a new block from the first shard
	err = dp.SaveTransactions(createOutportBlock(1, []byte("hash2")))
	require.Nil(t, err)
	require.Equal(t, 3, firstCalls)
	require.Equal(t, 4, secondCalls)
}

func TestDataProcessor_SaveTransactionsBlockWithoutHashIsAlwaysSaved(t *testing.T) {
	t.Parallel()

	calls := 0
	dp, _ := NewDataProcessor(createMockArgDataProcessor(
		outputs.OutputWithPolicy{Name: "first", Output: &mock.OutputStub{
			SaveTransactionsCalled: func(_ *data.PreparedBlock) error {
				calls++
				return nil
			},
		}},
	))

	for i := 0; i < 3; i++ {
		err := dp.SaveTransactions(createOutportBlock(1, nil))
		require.Nil(t, err)
	}
	require.Equal(t, 3, calls)
}

func TestDataProcessor_RevertBlock(t *testing.T) {
	t.Parallel()

	saveCalls := 0
	var revertedBlock *data.RevertedBlock
	args := createMockArgDataProcessor(
		outputs.OutputWithPolicy{Name: "first", Output: &mock.OutputStub{
			SaveTransactionsCalled: func(_ *data.PreparedBlock) error {
				saveCalls++
				return nil
			},
			RevertBlockCalled: func(block *data.RevertedBlock) error {
				revertedBlock = block
				return nil
			},
		}},
	)
	args.TransactionsProc = &transactionsHandlerStub{
		GetHexEncodedHashesForRemoveCalled: func(headerData *data.HeaderData, body *dataBlock.Body) ([]string, []string) {
			require.Equal(t, uint32(1), headerData.ShardID)
			require.Equal(t, uint64(10), headerData.Round)
			require.Equal(t, uint64(5000), headerData.TimestampMs)
			require.NotNil(t, body)

			return []string{"tx1", "tx2"}, []string{"scr1"}
		},
	}
	dp, _ := NewDataProcessor(args)

	obh := createOutportBlock(1, []byte("hash"))
	require.Nil(t, dp.SaveTransactions(obh))
	require.Nil(t, dp.SaveTransactions(obh))
	require.Equal(t, 1, saveCalls)

	header := &dataBlock.Header{ShardID: 1, Round: 10}
	blockData := &outport.BlockData{HeaderHash: []byte("hash"), Body: &dataBlock.Body{}, TimestampMs: 5000}
	err := dp.RevertBlock(header, blockData)
	require.Nil(t, err)

	require.NotNil(t, revertedBlock)
	require.True(t, revertedBlock.Header == coreData.HeaderHandler(header))
	require.True(t, revertedBlock.BlockData == blockData)
	require.Equal(t, []string{"tx1", "tx2"}, revertedBlock.TxHashes)
	require.Equal(t, []string{"scr1"}, revertedBlock.ScrHashes)

	// after the revert the same block has to be saved again
	require.Nil(t, dp.SaveTransactions(obh))
	require.Equal(t, 2, saveCalls)
}

func TestDataProcessor_PrepareBlock(t *testing.T) {
	t.Parallel()

	type prepareCall struct {
		round      uint64
		epoch      uint32
		timestamp  uint64
		isImportDB bool
		pool       *outport.TransactionPool
	}
	prepareCalls := make([]prepareCall, 0)
	operationsImportDB := make([]bool, 0)

	headerTx := &data.Transaction{Hash: "headerTx"}
	args := createMockArgDataProcessor()
	args.TransactionsProc = &transactionsHandlerStub{
		PrepareTransactionsForDatabaseCalled: func(_ []*dataBlock.MiniBlock, headerData *data.HeaderData, pool *outport.TransactionPool, isImportDB bool) *data.PreparedResults {
			prepareCalls = append(prepareCalls, prepareCall{
				round:      headerData.Round,
				epoch:      headerData.Epoch,
				timestamp:  headerData.TimestampMs,
				isImportDB: isImportDB,
				pool:       pool,
			})

			return &data.PreparedResults{Transactions: []*data.Transaction{headerTx}}
		},
	}
	args.LogsAndEventsProc = &logsAndEventsHandlerStub{
		ExtractDataFromLogsCalled: func(_ []*transaction.LogData, preparedResults *data.PreparedResults, shardID uint32, numOfShards uint32, _ uint64) *data.PreparedLogsResults {
			require.Equal(t, uint32(2), shardID)
			require.Equal(t, uint32(3), numOfShards)
			// the logs processor can change the prepared transactions before the operations are computed
			preparedResults.Transactions[0].HasLogs = true

			return &data.PreparedLogsResults{DBEvents: []*data.LogEvent{{ID: "event"}}}
		},
	}
	args.OperationsProc = &operationsHandlerStub{
		ProcessTransactionsAndSCRsCalled: func(txs []*data.Transaction, scrs []*data.ScResult, isImportDB bool, shardID uint32) ([]*data.Transaction, []*data.ScResult) {
			require.Equal(t, uint32(2), shardID)
			require.True(t, txs[0].HasLogs)
			operationsImportDB = append(operationsImportDB, isImportDB)

			copiedTx := *txs[0]
			return []*data.Transaction{&copiedTx}, scrs
		},
	}
	dp, _ := NewDataProcessor(args)
	require.Nil(t, dp.SetOutportConfig(outport.OutportConfig{IsInImportDBMode: true}))

	firstPool := &outport.TransactionPool{}
	secondPool := &outport.TransactionPool{}
	headerPool := &outport.TransactionPool{}
	obh := &outport.OutportBlockWithHeader{
		Header: &dataBlock.HeaderV3{
			ShardID: 2,
			Round:   30,
			Epoch:   7,
			ExecutionResults: []*dataBlock.ExecutionResult{
				{BaseExecutionResult: &dataBlock.BaseExecutionResult{HeaderHash: []byte("exec1"), HeaderRound: 28, HeaderEpoch: 6}},
				{BaseExecutionResult: &dataBlock.BaseExecutionResult{HeaderHash: []byte("missing"), HeaderRound: 29, HeaderEpoch: 6}},
				{BaseExecutionResult: &dataBlock.BaseExecutionResult{HeaderHash: []byte("exec2"), HeaderRound: 29, HeaderEpoch: 7}},
			},
		},
		OutportBlock: &outport.OutportBlock{
			NumberOfShards: 3,
			BlockData: &outport.BlockData{
				HeaderHash:  []byte("header"),
				Body:        &dataBlock.Body{},
				TimestampMs: 30000,
				Results: map[string]*outport.ExecutionResultData{
					hex.EncodeToString([]byte("exec1")): {Body: &dataBlock.Body{}, TransactionPool: firstPool, TimestampMs: 28000},
					hex.EncodeToString([]byte("exec2")): {Body: &dataBlock.Body{}, TransactionPool: secondPool, TimestampMs: 29000},
				},
			},
			TransactionPool: headerPool,
		},
	}

	block := dp.PrepareBlock(obh)
	require.True(t, block.OutportBlock == obh)

	// one unit for the header and one unit for every execution result that has data
	require.Len(t, block.Units, 3)
	require.Equal(t, []byte("header"), block.Units[0].BlockHash)
	require.Equal(t, []byte("exec1"), block.Units[1].BlockHash)
	require.Equal(t, []byte("exec2"), block.Units[2].BlockHash)
	// the block of the header v3 is not executed yet, the blocks of the execution results are executed
	require.False(t, block.Units[0].Executed)
	require.True(t, block.Units[1].Executed)
	require.True(t, block.Units[2].Executed)

	require.Equal(t, []prepareCall{
		{round: 30, epoch: 7, timestamp: 30000, isImportDB: true, pool: headerPool},
		{round: 28, epoch: 6, timestamp: 28000, isImportDB: true, pool: firstPool},
		{round: 29, epoch: 7, timestamp: 29000, isImportDB: true, pool: secondPool},
	}, prepareCalls)
	require.Equal(t, []bool{true, true, true}, operationsImportDB)

	for _, unit := range block.Units {
		require.Equal(t, uint32(2), unit.HeaderData.ShardID)
		require.Len(t, unit.Results.Transactions, 1)
		require.Len(t, unit.Logs.DBEvents, 1)
		// the operations are copies of the prepared transactions
		require.Len(t, unit.OperationsTxs, 1)
		require.True(t, unit.OperationsTxs[0] != unit.Results.Transactions[0])
		require.True(t, unit.OperationsTxs[0].HasLogs)
	}

	// the block of a header that is not v3 is executed
	oldHeaderBlock := dp.PrepareBlock(&outport.OutportBlockWithHeader{
		Header: &dataBlock.HeaderV2{Header: &dataBlock.Header{ShardID: 2, Round: 31}},
		OutportBlock: &outport.OutportBlock{
			NumberOfShards:  3,
			BlockData:       &outport.BlockData{HeaderHash: []byte("oldHeader"), Body: &dataBlock.Body{}},
			TransactionPool: &outport.TransactionPool{},
		},
	})
	require.Len(t, oldHeaderBlock.Units, 1)
	require.True(t, oldHeaderBlock.Units[0].Executed)
}

func TestDataProcessor_TheOtherCallsAreHandedToAllTheOutputs(t *testing.T) {
	t.Parallel()

	calls := make(map[string]int)
	createOutput := func(name string) outputs.OutputWithPolicy {
		return outputs.OutputWithPolicy{Name: name, Output: &mock.OutputStub{
			SaveHeaderCalled: func(_ *outport.OutportBlockWithHeader) error {
				calls[name+"-header"]++
				return nil
			},
			SaveMiniblocksCalled: func(_ *outport.OutportBlockWithHeader) error {
				calls[name+"-miniblocks"]++
				return nil
			},
			SaveValidatorsRatingCalled: func(_ *outport.ValidatorsRating) error {
				calls[name+"-rating"]++
				return nil
			},
			SaveRoundsInfoCalled: func(_ *outport.RoundsInfo) error {
				calls[name+"-rounds"]++
				return nil
			},
			SaveShardValidatorsPubKeysCalled: func(_ *outport.ValidatorsPubKeys) error {
				calls[name+"-validators"]++
				return nil
			},
			SaveAccountsCalled: func(_ *outport.Accounts) error {
				calls[name+"-accounts"]++
				return nil
			},
		}}
	}
	dp, _ := NewDataProcessor(createMockArgDataProcessor(createOutput("first"), createOutput("second")))

	obh := createOutportBlock(1, []byte("hash"))
	require.Nil(t, dp.SaveHeader(obh))
	require.Nil(t, dp.SaveMiniblocks(obh))
	require.Nil(t, dp.SaveValidatorsRating(&outport.ValidatorsRating{}))
	require.Nil(t, dp.SaveRoundsInfo(&outport.RoundsInfo{}))
	require.Nil(t, dp.SaveShardValidatorsPubKeys(&outport.ValidatorsPubKeys{}))
	require.Nil(t, dp.SaveAccounts(&outport.Accounts{}))

	// the header and the miniblocks are saved again if the block is received again
	require.Nil(t, dp.SaveHeader(obh))
	require.Nil(t, dp.SaveMiniblocks(obh))

	for _, name := range []string{"first", "second"} {
		require.Equal(t, 2, calls[name+"-header"])
		require.Equal(t, 2, calls[name+"-miniblocks"])
		require.Equal(t, 1, calls[name+"-rating"])
		require.Equal(t, 1, calls[name+"-rounds"])
		require.Equal(t, 1, calls[name+"-validators"])
		require.Equal(t, 1, calls[name+"-accounts"])
	}
}

func TestDataProcessor_Close(t *testing.T) {
	t.Parallel()

	errFirst := errors.New("first close error")
	errSecond := errors.New("second close error")
	calls := make([]string, 0)
	dp, _ := NewDataProcessor(createMockArgDataProcessor(
		outputs.OutputWithPolicy{Name: "first", Output: &mock.OutputStub{
			CloseCalled: func() error {
				calls = append(calls, "first")
				return errFirst
			},
		}},
		outputs.OutputWithPolicy{Name: "second", Output: &mock.OutputStub{
			CloseCalled: func() error {
				calls = append(calls, "second")
				return errSecond
			},
		}},
	))

	err := dp.Close()
	// the outputs are closed in the reverse order and all the errors are returned
	require.Equal(t, []string{"second", "first"}, calls)
	require.True(t, errors.Is(err, errFirst))
	require.True(t, errors.Is(err, errSecond))
}
