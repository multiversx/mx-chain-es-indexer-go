package transactions

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strings"

	elasticIndexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-logger/check"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/indexer"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

const (
	okHexEncoded = "6f6b"
)

func checkTxsProcessorArg(args *ArgsTransactionProcessor) error {
	if args == nil {
		return elasticIndexer.ErrNilTransactionsProcessorArguments
	}
	if check.IfNil(args.Marshalizer) {
		return elasticIndexer.ErrNilMarshalizer
	}
	if check.IfNil(args.ShardCoordinator) {
		return elasticIndexer.ErrNilShardCoordinator
	}
	if check.IfNil(args.Hasher) {
		return elasticIndexer.ErrNilHasher
	}
	if check.IfNil(args.AddressPubkeyConverter) {
		return elasticIndexer.ErrNilPubkeyConverter
	}
	if check.IfNil(args.TxFeeCalculator) {
		return elasticIndexer.ErrNilTransactionFeeCalculator
	}

	return nil
}

func checkPrepareTransactionForDatabaseArguments(
	body *block.Body,
	header nodeData.HeaderHandler,
	pool *indexer.Pool,
) error {
	if body == nil {
		return elasticIndexer.ErrNilBlockBody
	}
	if check.IfNil(header) {
		return elasticIndexer.ErrNilHeaderHandler
	}
	if pool == nil {
		return elasticIndexer.ErrNilPool
	}

	return nil
}

func isScResultSuccessful(scResultData []byte) bool {
	okReturnDataNewVersion := []byte("@" + hex.EncodeToString([]byte(vmcommon.Ok.String())))
	okReturnDataOldVersion := []byte("@" + vmcommon.Ok.String()) // backwards compatible

	return bytes.Contains(scResultData, okReturnDataNewVersion) || bytes.Contains(scResultData, okReturnDataOldVersion)
}

func isSCRForSenderWithRefund(dbScResult *data.ScResult, tx *data.Transaction) bool {
	isForSender := dbScResult.Receiver == tx.Sender
	isRightNonce := dbScResult.Nonce == tx.Nonce+1
	isFromCurrentTx := dbScResult.PrevTxHash == tx.Hash
	isScrDataOk := isDataOk(dbScResult.Data)

	return isFromCurrentTx && isForSender && isRightNonce && isScrDataOk
}

func isDataOk(data []byte) bool {
	dataFieldStr := "@" + okHexEncoded

	return strings.HasPrefix(string(data), dataFieldStr)
}

func stringValueToBigInt(strValue string) *big.Int {
	value, ok := big.NewInt(0).SetString(strValue, 10)
	if !ok {
		return big.NewInt(0)
	}

	return value
}

func isRelayedTx(tx *data.Transaction) bool {
	return strings.HasPrefix(string(tx.Data), core.RelayedTransaction) && len(tx.SmartContractResults) > 0
}

func isESDTNFTTransfer(tx *data.Transaction) bool {
	return strings.HasPrefix(string(tx.Data), core.BuiltInFunctionESDTNFTTransfer) && len(tx.SmartContractResults) > 0
}

func isCrossShardDstMe(tx *data.Transaction, selfShardID uint32) bool {
	return tx.SenderShard != tx.ReceiverShard && tx.ReceiverShard == selfShardID
}

func isIntraShardOrInvalid(tx *data.Transaction, selfShardID uint32) bool {
	return (tx.SenderShard == tx.ReceiverShard && tx.ReceiverShard == selfShardID) || tx.Status == transaction.TxStatusInvalid.String()
}
