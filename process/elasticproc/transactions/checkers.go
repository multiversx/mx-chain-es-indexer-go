package transactions

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-es-indexer-go/data"
	elasticIndexer "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
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
	if check.IfNil(args.Hasher) {
		return elasticIndexer.ErrNilHasher
	}
	if check.IfNil(args.AddressPubkeyConverter) {
		return elasticIndexer.ErrNilPubkeyConverter
	}
	if check.IfNil(args.BalanceConverter) {
		return elasticIndexer.ErrNilBalanceConverter
	}
	if check.IfNil(args.TxHashExtractor) {
		return ErrNilTxHashExtractor
	}
	if check.IfNil(args.RewardTxData) {
		return ErrNilRewardTxDataHandler
	}

	return nil
}

func areESDTValuesOK(values []string) bool {
	for _, value := range values {
		if len(value) > data.MaxESDTValueLength {
			return false
		}
	}

	return true
}

func checkPrepareTransactionForDatabaseArguments(
	header coreData.HeaderHandler,
	pool *outport.TransactionPool,
) error {

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
	isRelayed := strings.HasPrefix(string(tx.Data), core.RelayedTransaction) || strings.HasPrefix(string(tx.Data), core.RelayedTransactionV2)
	return isRelayed && len(tx.SmartContractResults) > 0
}

func isCrossShardOnSourceShard(tx *data.Transaction, selfShardID uint32) bool {
	return tx.SenderShard != tx.ReceiverShard && tx.SenderShard == selfShardID
}
