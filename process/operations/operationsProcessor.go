package operations

import (
	"encoding/hex"
	"strings"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type operationsProcessor struct {
	importDBMode     bool
	shardCoordinator indexer.ShardCoordinator
}

// NewOperationsProcessor will create a new instance of operationsProcessor
func NewOperationsProcessor(importDBMode bool, shardCoordinator indexer.ShardCoordinator) (*operationsProcessor, error) {
	if check.IfNil(shardCoordinator) {
		return nil, indexer.ErrNilShardCoordinator
	}

	return &operationsProcessor{
		importDBMode:     importDBMode,
		shardCoordinator: shardCoordinator,
	}, nil
}

// ProcessTransactionsAndSCRs will prepare transactions and smart contract results to be indexed
func (op *operationsProcessor) ProcessTransactionsAndSCRs(
	txs []*data.Transaction,
	scrs []*data.ScResult,
) ([]*data.Transaction, []*data.ScResult) {
	for idx, tx := range txs {
		if !op.shouldIndex(txs[idx].ReceiverShard) {
			// remove tx from slice
			txs = append(txs[:idx], txs[idx+1:]...)
			continue
		}

		tx.SmartContractResults = nil
		tx.Type = string(transaction.TxTypeNormal)
	}

	for idx := 0; idx < len(scrs); idx++ {
		if !op.shouldIndex(scrs[idx].ReceiverShard) {
			// remove scr from slice
			scrs = append(scrs[:idx], scrs[idx+1:]...)
			continue
		}

		scr := scrs[idx]
		scr.Type = string(transaction.TxTypeUnsigned)

		setCanBeIgnoredField(scr)

		selfShard := op.shardCoordinator.SelfId()
		if selfShard == scr.ReceiverShard {
			scr.Status = transaction.TxStatusSuccess.String()
		} else {
			scr.Status = transaction.TxStatusPending.String()
		}
	}

	return txs, scrs
}

func (op *operationsProcessor) shouldIndex(destinationShardID uint32) bool {
	if !op.importDBMode {
		return true
	}

	return op.shardCoordinator.SelfId() == destinationShardID
}

func setCanBeIgnoredField(scr *data.ScResult) {
	dataFieldStr := string(scr.Data)
	hasOkPrefix := strings.HasPrefix(dataFieldStr, hex.EncodeToString([]byte(vmcommon.Ok.String())))
	if hasOkPrefix {
		scr.CanBeIgnored = true
		return
	}

	isNFTTransferOrMultiTransfer := strings.HasPrefix(dataFieldStr, core.BuiltInFunctionESDTNFTTransfer) || strings.HasPrefix(dataFieldStr, core.BuiltInFunctionMultiESDTNFTTransfer)
	isSCAddr := core.IsSmartContractAddress(scr.SenderAddressBytes)
	if isNFTTransferOrMultiTransfer && !isSCAddr {
		scr.CanBeIgnored = true
		return
	}
}
