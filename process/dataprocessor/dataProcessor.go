package dataprocessor

import (
	"errors"
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/process/outputs"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("indexer/dataprocessor")

// ArgDataProcessor holds all dependencies required by the dataProcessor in order to create new instances
type ArgDataProcessor struct {
	TransactionsProc  TransactionsHandler
	LogsAndEventsProc LogsAndEventsHandler
	OperationsProc    OperationsHandler
	// Outputs are called in the provided order
	Outputs []outputs.OutputWithPolicy
}

// dataProcessor prepares the data received from the node only once and hands it to every output
type dataProcessor struct {
	transactionsProc  TransactionsHandler
	logsAndEventsProc LogsAndEventsHandler
	operationsProc    OperationsHandler
	outputs           []outputs.OutputWithPolicy

	importDB    bool
	mutImportDB sync.RWMutex

	// savedBlocks holds, for every output and every shard, the hash of the last block whose transactions were saved.
	// If the node sends the same block again (because another output failed), the outputs that already saved it are skipped
	savedBlocks    []map[uint32]string
	mutSavedBlocks sync.Mutex
}

// NewDataProcessor creates a new instance of dataProcessor
func NewDataProcessor(args ArgDataProcessor) (*dataProcessor, error) {
	err := checkArguments(args)
	if err != nil {
		return nil, err
	}

	savedBlocks := make([]map[uint32]string, len(args.Outputs))
	for idx := range savedBlocks {
		savedBlocks[idx] = make(map[uint32]string)
	}

	return &dataProcessor{
		transactionsProc:  args.TransactionsProc,
		logsAndEventsProc: args.LogsAndEventsProc,
		operationsProc:    args.OperationsProc,
		outputs:           args.Outputs,
		savedBlocks:       savedBlocks,
	}, nil
}

func checkArguments(args ArgDataProcessor) error {
	if check.IfNilReflect(args.TransactionsProc) {
		return ErrNilTransactionsHandler
	}
	if check.IfNilReflect(args.LogsAndEventsProc) {
		return ErrNilLogsAndEventsHandler
	}
	if check.IfNilReflect(args.OperationsProc) {
		return ErrNilOperationsHandler
	}
	if len(args.Outputs) == 0 {
		return ErrNoOutputs
	}
	for _, output := range args.Outputs {
		if check.IfNil(output.Output) {
			return fmt.Errorf("%w, output name: %s", ErrNilOutput, output.Name)
		}
	}

	return nil
}

// SaveHeader hands the header to every output
func (dp *dataProcessor) SaveHeader(obh *outport.OutportBlockWithHeader) error {
	return dp.callOutputs(func(output outputs.Output) error {
		return output.SaveHeader(obh)
	})
}

// SaveMiniblocks hands the miniblocks to every output
func (dp *dataProcessor) SaveMiniblocks(obh *outport.OutportBlockWithHeader) error {
	return dp.callOutputs(func(output outputs.Output) error {
		return output.SaveMiniblocks(obh)
	})
}

// SaveTransactions prepares the transactions, smart contract results, logs, events and operations of the block
// and hands them to every output
func (dp *dataProcessor) SaveTransactions(obh *outport.OutportBlockWithHeader) error {
	block := dp.PrepareBlock(obh)

	shardID := obh.Header.GetShardID()
	blockHash := string(obh.BlockData.HeaderHash)
	for idx, output := range dp.outputs {
		if dp.wasBlockSaved(idx, shardID, blockHash) {
			log.Debug("dataProcessor.SaveTransactions: the block was already saved, skipping the output",
				"output", output.Name, "shardID", shardID, "hash", obh.BlockData.HeaderHash)
			continue
		}

		err := handleOutputError(output, output.Output.SaveTransactions(block))
		if err != nil {
			return err
		}

		dp.setBlockSaved(idx, shardID, blockHash)
	}

	return nil
}

// RevertBlock prepares the data of the block that should be reverted and hands it to every output
func (dp *dataProcessor) RevertBlock(header coreData.HeaderHandler, blockData *outport.BlockData) error {
	// after a revert the same block can be received and it has to be saved again by all the outputs
	dp.resetSavedBlocks(header.GetShardID())

	revertedBlock := dp.PrepareRevertedBlock(header, blockData)

	return dp.callOutputs(func(output outputs.Output) error {
		return output.RevertBlock(revertedBlock)
	})
}

// SaveValidatorsRating hands the validators rating to every output
func (dp *dataProcessor) SaveValidatorsRating(ratingData *outport.ValidatorsRating) error {
	return dp.callOutputs(func(output outputs.Output) error {
		return output.SaveValidatorsRating(ratingData)
	})
}

// SaveRoundsInfo hands the rounds info to every output
func (dp *dataProcessor) SaveRoundsInfo(rounds *outport.RoundsInfo) error {
	return dp.callOutputs(func(output outputs.Output) error {
		return output.SaveRoundsInfo(rounds)
	})
}

// SaveShardValidatorsPubKeys hands the validators public keys to every output
func (dp *dataProcessor) SaveShardValidatorsPubKeys(validatorsPubKeys *outport.ValidatorsPubKeys) error {
	return dp.callOutputs(func(output outputs.Output) error {
		return output.SaveShardValidatorsPubKeys(validatorsPubKeys)
	})
}

// SaveAccounts hands the accounts to every output
func (dp *dataProcessor) SaveAccounts(accounts *outport.Accounts) error {
	return dp.callOutputs(func(output outputs.Output) error {
		return output.SaveAccounts(accounts)
	})
}

// SetOutportConfig will set the outport config
func (dp *dataProcessor) SetOutportConfig(cfg outport.OutportConfig) error {
	dp.mutImportDB.Lock()
	defer dp.mutImportDB.Unlock()

	dp.importDB = cfg.IsInImportDBMode

	return nil
}

func (dp *dataProcessor) isImportDB() bool {
	dp.mutImportDB.RLock()
	defer dp.mutImportDB.RUnlock()

	return dp.importDB
}

// Close closes all the outputs, in the reverse order
func (dp *dataProcessor) Close() error {
	errs := make([]error, 0)
	for idx := len(dp.outputs) - 1; idx >= 0; idx-- {
		output := dp.outputs[idx]
		err := output.Output.Close()
		if err != nil {
			errs = append(errs, fmt.Errorf("%w when closing the output %s", err, output.Name))
		}
	}

	return errors.Join(errs...)
}

// IsInterfaceNil returns true if there is no value under the interface
func (dp *dataProcessor) IsInterfaceNil() bool {
	return dp == nil
}

func (dp *dataProcessor) callOutputs(call func(output outputs.Output) error) error {
	for _, output := range dp.outputs {
		err := handleOutputError(output, call(output.Output))
		if err != nil {
			return err
		}
	}

	return nil
}

func handleOutputError(output outputs.OutputWithPolicy, err error) error {
	if err == nil {
		return nil
	}

	if output.ContinueOnError {
		log.Warn("dataProcessor: the output returned an error, will continue", "output", output.Name, "error", err)
		return nil
	}

	return fmt.Errorf("%w, output: %s", err, output.Name)
}

func (dp *dataProcessor) wasBlockSaved(outputIdx int, shardID uint32, blockHash string) bool {
	// the blocks without hash cannot be identified
	if len(blockHash) == 0 {
		return false
	}

	dp.mutSavedBlocks.Lock()
	defer dp.mutSavedBlocks.Unlock()

	savedHash, found := dp.savedBlocks[outputIdx][shardID]

	return found && savedHash == blockHash
}

func (dp *dataProcessor) setBlockSaved(outputIdx int, shardID uint32, blockHash string) {
	if len(blockHash) == 0 {
		return
	}

	dp.mutSavedBlocks.Lock()
	defer dp.mutSavedBlocks.Unlock()

	dp.savedBlocks[outputIdx][shardID] = blockHash
}

func (dp *dataProcessor) resetSavedBlocks(shardID uint32) {
	dp.mutSavedBlocks.Lock()
	defer dp.mutSavedBlocks.Unlock()

	for idx := range dp.savedBlocks {
		delete(dp.savedBlocks[idx], shardID)
	}
}
