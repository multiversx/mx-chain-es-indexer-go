package transactions

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
)

// SerializeScResults will serialize the provided smart contract results in a way that ElasticSearch expects a bulk request
func (tdp *txsDatabaseProcessor) SerializeScResults(scResults []*data.ScResult, buffSlice *data.BufferSlice, index string) error {
	for _, sc := range scResults {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_index": "%s", "_id" : "%s" } }%s`, index, converters.JsonEscape(sc.Hash), "\n"))
		serializedData, errPrepareSc := json.Marshal(sc)
		if errPrepareSc != nil {
			return errPrepareSc
		}

		err := buffSlice.PutData(meta, serializedData)
		if err != nil {
			return err
		}
	}

	return nil
}

// SerializeReceipts will serialize the receipts in a way that ElasticSearch expects a bulk request
func (tdp *txsDatabaseProcessor) SerializeReceipts(receipts []*data.Receipt, buffSlice *data.BufferSlice, index string) error {
	for _, rec := range receipts {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_index": "%s", "_id" : "%s" } }%s`, index, converters.JsonEscape(rec.Hash), "\n"))
		serializedData, errPrepareReceipt := json.Marshal(rec)
		if errPrepareReceipt != nil {
			return errPrepareReceipt
		}

		err := buffSlice.PutData(meta, serializedData)
		if err != nil {
			return err
		}
	}

	return nil
}

// SerializeTransactionsFeeData will serialize transactions fee data
func (tdp *txsDatabaseProcessor) SerializeTransactionsFeeData(txHashRefund map[string]*data.FeeData, buffSlice *data.BufferSlice, index string) error {
	for txHash, feeData := range txHashRefund {
		meta := []byte(fmt.Sprintf(`{"update":{ "_index":"%s","_id":"%s"}}%s`, index, converters.JsonEscape(txHash), "\n"))

		var codeToExecute string
		if feeData.GasRefunded != 0 {
			codeToExecute = `
 			if ('create' == ctx.op) {
 				ctx.op = 'noop'
 			} else {
				boolean ok1 = ((ctx._source.containsKey('initialPaidFee')) && (ctx._source.initialPaidFee != null) && (!ctx._source.initialPaidFee.isEmpty()));
				boolean ok2 = ((ctx._source.containsKey('fee')) && (ctx._source.fee != null) && (!ctx._source.fee.isEmpty()));
				if (!ok1 || !ok2) {
					return
				}
 				BigInteger feeFromSource;
				if ((ctx._source.containsKey('hadRefund')) && (ctx._source.hadRefund)) {
					feeFromSource = new BigInteger(ctx._source.fee);
				} else {
					feeFromSource = new BigInteger(ctx._source.initialPaidFee);
					ctx._source.hadRefund = true;
				}

 				BigInteger fee = new BigInteger(params.fee);
 				if (feeFromSource.compareTo(fee) > 0) {
 					ctx._source.fee = feeFromSource.subtract(fee).toString();	
 				}
 				if (ctx._source.feeNum > params.feeNum) {
 					ctx._source.feeNum -= params.feeNum;	
 				}
 				if (ctx._source.gasUsed > params.gasRefunded) {
 					ctx._source.gasUsed -= params.gasRefunded;	
 				}
 			}
 `
		} else {
			codeToExecute = `
			if ('create' == ctx.op) {
				ctx.op = 'noop'
			} else {
				ctx._source.fee = params.fee;
				ctx._source.feeNum = params.feeNum;
				ctx._source.gasUsed = params.gasUsed;
			}
`
		}

		serializedDataStr := fmt.Sprintf(`{"scripted_upsert": true, "script": {`+
			`"source": "%s",`+
			`"lang": "painless",`+
			`"params": {"fee": "%s", "gasUsed": %d, "feeNum": %g, "gasRefunded": %d}},`+
			`"upsert": {}}`,
			converters.FormatPainlessSource(codeToExecute), feeData.Fee, feeData.GasUsed, feeData.FeeNum, feeData.GasRefunded,
		)

		err := buffSlice.PutData(meta, []byte(serializedDataStr))
		if err != nil {
			return err
		}
	}

	return nil
}

// SerializeTransactions will serialize the transactions in a way that Elasticsearch expects a bulk request
func (tdp *txsDatabaseProcessor) SerializeTransactions(
	transactions []*data.Transaction,
	txHashStatusInfo map[string]*outport.StatusInfo,
	selfShardID uint32,
	buffSlice *data.BufferSlice,
	index string,
) error {
	for _, tx := range transactions {
		meta, serializedData, err := prepareSerializedDataForATransaction(tx, selfShardID, index)
		if err != nil {
			return err
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			return err
		}
	}

	err := serializeTxHashStatus(buffSlice, txHashStatusInfo, index)
	if err != nil {
		return err
	}

	return nil
}

func serializeTxHashStatus(buffSlice *data.BufferSlice, txHashStatusInfo map[string]*outport.StatusInfo, index string) error {
	for txHash, statusInfo := range txHashStatusInfo {
		metaData := []byte(fmt.Sprintf(`{"update":{ "_index":"%s","_id":"%s"}}%s`, index, txHash, "\n"))

		newTx := &data.Transaction{
			Status:         statusInfo.Status,
			ErrorEvent:     statusInfo.ErrorEvent,
			CompletedEvent: statusInfo.CompletedEvent,
		}
		marshaledTx, err := json.Marshal(newTx)
		if err != nil {
			return err
		}
		marshaledStatusInfo, err := json.Marshal(statusInfo)
		if err != nil {
			return err
		}

		codeToExecute := `
			if (!params.statusInfo.status.isEmpty()) {
				ctx._source.status = params.statusInfo.status;
			}

			if (params.statusInfo.completedEvent) {
				ctx._source.completedEvent = params.statusInfo.completedEvent;
			}
			
			if (params.statusInfo.errorEvent) {
				ctx._source.errorEvent = params.statusInfo.errorEvent;
			}
`
		serializedData := []byte(fmt.Sprintf(`{"script": {"source": "%s","lang": "painless","params": {"statusInfo": %s}}, "upsert": %s }`, converters.FormatPainlessSource(codeToExecute), string(marshaledStatusInfo), string(marshaledTx)))
		err = buffSlice.PutData(metaData, serializedData)
		if err != nil {
			return err
		}
	}

	return nil
}

func prepareSerializedDataForATransaction(
	tx *data.Transaction,
	selfShardID uint32,
	index string,
) ([]byte, []byte, error) {
	metaData := []byte(fmt.Sprintf(`{"update":{ "_index":"%s", "_id":"%s"}}%s`, index, converters.JsonEscape(tx.Hash), "\n"))
	marshaledTx, err := json.Marshal(tx)
	if err != nil {
		return nil, nil, err
	}

	if isCrossShardOnSourceShard(tx, selfShardID) {
		if isSimpleESDTTransfer(tx) {
			codeToExecute := `
				if ('create' == ctx.op) {
					ctx._source = params.tx;
				} else {
					ctx._source.gasUsed = params.tx.gasUsed;
					ctx._source.fee = params.tx.fee;
					ctx._source.feeNum = params.tx.feeNum;
				}
			`
			serializedData := []byte(fmt.Sprintf(`{"scripted_upsert": true, "script":{"source":"%s","lang": "painless","params":{"tx": %s}},"upsert":{}}`,
				converters.FormatPainlessSource(codeToExecute), string(marshaledTx)))

			return metaData, serializedData, nil
		}

		serializedData := []byte(fmt.Sprintf(`{"script":{"source":"return"},"upsert":%s}`, string(marshaledTx)))

		return metaData, serializedData, nil
	}

	if isNFTTransferOrMultiTransfer(tx) {
		serializedData, errPrep := prepareNFTESDTTransferOrMultiESDTTransfer(marshaledTx)
		if errPrep != nil {
			return nil, nil, err
		}

		return metaData, serializedData, nil
	}

	if isSimpleESDTTransferCrossShardOnDestination(tx, selfShardID) {
		return metaData, prepareSerializedDataForESDTTransferOnDestination(marshaledTx), nil
	}

	// transaction is intra-shard, invalid or cross-shard destination me
	meta := []byte(fmt.Sprintf(`{ "index" : { "_index":"%s", "_id" : "%s" } }%s`, index, converters.JsonEscape(tx.Hash), "\n"))

	return meta, marshaledTx, nil
}

func prepareSerializedDataForESDTTransferOnDestination(marshaledTx []byte) []byte {
	codeToExecute := `
		if ('create' == ctx.op) {
			ctx._source = params.tx;
		} else {
			def gasUsed = ctx._source.gasUsed;
			def fee = ctx._source.fee;
			def feeNum = ctx._source.feeNum;
			ctx._source = params.tx;
			ctx._source.gasUsed = gasUsed;
			ctx._source.fee = fee;
			ctx._source.feeNum = feeNum;
		}
`
	return []byte(fmt.Sprintf(`{"scripted_upsert": true, "script":{"source":"%s","lang": "painless","params":{"tx": %s}},"upsert":{}}`,
		converters.FormatPainlessSource(codeToExecute), string(marshaledTx)))
}

func prepareNFTESDTTransferOrMultiESDTTransfer(marshaledTx []byte) ([]byte, error) {
	codeToExecute := `
		if ('create' == ctx.op) {
			ctx._source = params.tx;
		} else {
			def status = ctx._source.status;
			def errorEvent = ctx._source.errorEvent;
			def completedEvent = ctx._source.completedEvent;

			ctx._source = params.tx;
			if (!status.isEmpty()) {
				ctx._source.status = status;
			}
			if (errorEvent != null) {
				ctx._source.errorEvent = errorEvent;
			}
			if (completedEvent != null) {
				ctx._source.completedEvent = completedEvent;
			}
		}
`
	serializedData := []byte(fmt.Sprintf(`{"scripted_upsert": true, "script":{"source":"%s","lang": "painless","params":{"tx": %s}},"upsert":{}}`,
		converters.FormatPainlessSource(codeToExecute), string(marshaledTx)))

	return serializedData, nil
}

func isNFTTransferOrMultiTransfer(tx *data.Transaction) bool {
	if tx.SenderShard != tx.ReceiverShard {
		return false
	}

	splitData := strings.Split(string(tx.Data), data.AtSeparator)
	if len(splitData) < minNumOfArgumentsNFTTransferORMultiTransfer {
		return false
	}

	return splitData[0] == core.BuiltInFunctionESDTNFTTransfer || splitData[0] == core.BuiltInFunctionMultiESDTNFTTransfer
}

func isSimpleESDTTransferCrossShardOnDestination(tx *data.Transaction, selfShard uint32) bool {
	isCrossOnDestination := tx.SenderShard != tx.ReceiverShard && tx.ReceiverShard == selfShard

	return isSimpleESDTTransfer(tx) && isCrossOnDestination
}

func isSimpleESDTTransfer(tx *data.Transaction) bool {
	return tx.Operation == core.BuiltInFunctionESDTTransfer && tx.Function == ""
}
