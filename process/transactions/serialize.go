package transactions

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
)

// SerializeScResults will serialize the provided smart contract results in a way that Elastic Search expects a bulk request
func (tdp *txsDatabaseProcessor) SerializeScResults(scResults []*data.ScResult, buffSlice *data.BufferSlice, index string) error {
	for _, sc := range scResults {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_index": "%s", "_id" : "%s" } }%s`, index, sc.Hash, "\n"))
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

// SerializeReceipts will serialize the receipts in a way that Elastic Search expects a bulk request
func (tdp *txsDatabaseProcessor) SerializeReceipts(receipts []*data.Receipt, buffSlice *data.BufferSlice, index string) error {
	for _, rec := range receipts {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_index": "%s", "_id" : "%s" } }%s`, index, rec.Hash, "\n"))
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

// SerializeTransactionWithRefund will serialize transaction based on refund
func (tdp *txsDatabaseProcessor) SerializeTransactionWithRefund(
	txs map[string]*data.Transaction,
	txHashRefund map[string]*data.RefundData,
	buffSlice *data.BufferSlice,
	index string,
) error {
	for txHash, tx := range txs {
		refundForTx, ok := txHashRefund[txHash]
		if !ok {
			continue
		}

		if refundForTx.Receiver != tx.Sender {
			continue
		}

		refundValueBig, ok := big.NewInt(0).SetString(refundForTx.Value, 10)
		if !ok {
			continue
		}
		gasUsed, fee := tdp.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(tx, refundValueBig)
		tx.GasUsed = gasUsed
		tx.Fee = fee.String()

		meta := []byte(fmt.Sprintf(`{ "index" : { "_index": "%s", "_id" : "%s" } }%s`, index, txHash, "\n"))
		serializedData, errPrepare := json.Marshal(tx)
		if errPrepare != nil {
			return errPrepare
		}

		err := buffSlice.PutData(meta, serializedData)
		if err != nil {
			return err
		}
	}

	return nil
}

// SerializeTransactions will serialize the transactions in a way that Elastic Search expects a bulk request
func (tdp *txsDatabaseProcessor) SerializeTransactions(
	transactions []*data.Transaction,
	txHashStatus map[string]string,
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

	err := serializeTxHashStatus(buffSlice, txHashStatus, index)
	if err != nil {
		return err
	}

	return nil
}

func serializeTxHashStatus(buffSlice *data.BufferSlice, txHashStatus map[string]string, index string) error {
	for txHash, status := range txHashStatus {
		metaData := []byte(fmt.Sprintf(`{"update":{ "_index":"%s","_id":"%s", "_type": "_doc"}}%s`, index, txHash, "\n"))

		newTx := &data.Transaction{
			Status: status,
		}
		marshaledTx, err := json.Marshal(newTx)
		if err != nil {
			return err
		}

		serializedData := []byte(fmt.Sprintf(`{"script": {"source": "ctx._source.status = params.status","lang": "painless","params": {"status": "%s"}},"upsert": %s }`, status, string(marshaledTx)))
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
	metaData := []byte(fmt.Sprintf(`{"update":{ "_index":"%s", "_id":"%s", "_type": "_doc"}}%s`, index, tx.Hash, "\n"))
	marshaledTx, err := json.Marshal(tx)
	if err != nil {
		return nil, nil, err
	}

	if isCrossShardOnSourceShard(tx, selfShardID) {
		// if transaction is cross-shard and current shard ID is source, use upsert without updating anything
		serializedData :=
			[]byte(fmt.Sprintf(`{"script":{"source":"return"},"upsert":%s}`,
				string(marshaledTx)))
		log.Trace("indexer tx is on sender shard", "metaData", string(metaData), "serializedData", string(serializedData))

		return metaData, serializedData, nil
	}

	if isNFTTransferOrMultiTransfer(tx) {
		serializedData, errPrep := prepareNFTESDTTransferOrMultiESDTTransfer(marshaledTx)
		if errPrep != nil {
			return nil, nil, err
		}

		return metaData, serializedData, nil
	}

	// transaction is intra-shard, invalid or cross-shard destination me
	meta := []byte(fmt.Sprintf(`{ "index" : { "_index":"%s", "_id" : "%s", "_type" : "%s" } }%s`, index, tx.Hash, "_doc", "\n"))
	log.Trace("indexer tx is intra shard or invalid tx", "meta", string(meta), "marshaledTx", string(marshaledTx))

	return meta, marshaledTx, nil
}

func prepareNFTESDTTransferOrMultiESDTTransfer(marshaledTx []byte) ([]byte, error) {
	serializedData := []byte(fmt.Sprintf(`{"script":{"source":"`+
		`def status = ctx._source.status;`+
		`ctx._source = params.tx;`+
		`ctx._source.status = status;`+
		`","lang": "painless","params":`+
		`{"tx": %s}},"upsert":%s}`,
		string(marshaledTx), string(marshaledTx)))

	return serializedData, nil
}

func isNFTTransferOrMultiTransfer(tx *data.Transaction) bool {
	if len(tx.SmartContractResults) < 0 || tx.SenderShard != tx.ReceiverShard {
		return false
	}

	splitData := strings.Split(string(tx.Data), data.AtSeparator)
	if len(splitData) < minNumOfArgumentsNFTTransferORMultiTransfer {
		return false
	}

	return splitData[0] == core.BuiltInFunctionESDTNFTTransfer || splitData[0] == core.BuiltInFunctionMultiESDTNFTTransfer
}
