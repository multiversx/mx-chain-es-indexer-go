package transactions

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

// SerializeTokens will serialize the provided tokens data in a way that Elastic Search expects a bulk request
func (tdp *txsDatabaseProcessor) SerializeTokens(tokens []*data.TokenInfo) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()
	for _, tokenData := range tokens {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, tokenData.Identifier, "\n"))
		serializedData, errPrepareD := json.Marshal(tokenData)
		if errPrepareD != nil {
			return nil, errPrepareD
		}

		err := buffSlice.PutData(meta, serializedData)
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}

// SerializeDeploysData will serialize the provided deploys data in a way that Elastic Search expects a bulk request
func (tdp *txsDatabaseProcessor) SerializeDeploysData(deploys []*data.ScDeployInfo) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()
	for _, deployInfo := range deploys {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, deployInfo.ScAddress, "\n"))
		serializedData, errPrepareD := json.Marshal(deployInfo)
		if errPrepareD != nil {
			return nil, errPrepareD
		}

		err := buffSlice.PutData(meta, serializedData)
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}

// SerializeScResults will serialize the provided smart contract results in a way that Elastic Search expects a bulk request
func (tdp *txsDatabaseProcessor) SerializeScResults(scResults []*data.ScResult) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()
	for _, sc := range scResults {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, sc.Hash, "\n"))
		serializedData, errPrepareSc := json.Marshal(sc)
		if errPrepareSc != nil {
			return nil, errPrepareSc
		}

		err := buffSlice.PutData(meta, serializedData)
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}

// SerializeReceipts will serialize the receipts in a way that Elastic Search expects a bulk request
func (tdp *txsDatabaseProcessor) SerializeReceipts(receipts []*data.Receipt) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()
	for _, rec := range receipts {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, rec.Hash, "\n"))
		serializedData, errPrepareReceipt := json.Marshal(rec)
		if errPrepareReceipt != nil {
			return nil, errPrepareReceipt
		}

		err := buffSlice.PutData(meta, serializedData)
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}

// SerializeTransactions will serialize the transactions in a way that Elastic Search expects a bulk request
func (tdp *txsDatabaseProcessor) SerializeTransactions(
	transactions []*data.Transaction,
	selfShardID uint32,
	mbsHashInDB map[string]bool,
) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()
	for _, tx := range transactions {
		isMBOfTxInDB := mbsHashInDB[tx.MBHash]
		meta, serializedData, err := prepareSerializedDataForATransaction(tx, selfShardID, isMBOfTxInDB)
		if err != nil {
			return nil, err
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}

func prepareSerializedDataForATransaction(
	tx *data.Transaction,
	selfShardID uint32,
	_ bool,
) ([]byte, []byte, error) {
	metaData := []byte(fmt.Sprintf(`{"update":{"_id":"%s", "_type": "_doc"}}%s`, tx.Hash, "\n"))

	marshaledTx, err := json.Marshal(tx)
	if err != nil {
		return nil, nil, err
	}

	if isIntraShardOrInvalid(tx, selfShardID) {
		// if transaction is intra-shard, use basic insert as data can be re-written at forks
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s", "_type" : "%s" } }%s`, tx.Hash, "_doc", "\n"))
		log.Trace("indexer tx is intra shard or invalid tx", "meta", string(meta), "marshaledTx", string(marshaledTx))

		return meta, marshaledTx, nil
	}

	if !isCrossShardDstMe(tx, selfShardID) {
		// if transaction is cross-shard and current shard ID is source, use upsert without updating anything
		serializedData :=
			[]byte(fmt.Sprintf(`{"script":{"source":"return"},"upsert":%s}`,
				string(marshaledTx)))
		log.Trace("indexer tx is on sender shard", "metaData", string(metaData), "serializedData", string(serializedData))

		return metaData, serializedData, nil
	}

	serializedData, err := prepareCrossShardTxForDestinationSerialized(tx, marshaledTx)
	if err != nil {
		return nil, nil, err
	}

	log.Trace("indexer tx is on destination shard", "metaData", string(metaData), "serializedData", string(serializedData))

	return metaData, serializedData, nil
}

func prepareCrossShardTxForDestinationSerialized(tx *data.Transaction, marshaledTx []byte) ([]byte, error) {
	// if transaction is cross-shard and current shard ID is destination, use upsert with updating fields

	marshaledTimestamp, err := json.Marshal(tx.Timestamp)
	if err != nil {
		return nil, err
	}

	serializedData := []byte(fmt.Sprintf(`{"script":{"source":"`+
		`ctx._source.status = params.status;`+
		`ctx._source.miniBlockHash = params.miniBlockHash;`+
		`ctx._source.timestamp = params.timestamp;`+
		`ctx._source.gasUsed = params.gasUsed;`+
		`ctx._source.fee = params.fee;`+
		`ctx._source.hasScResults = params.hasScResults;`+
		`","lang": "painless","params":`+
		`{"status": "%s", "miniBlockHash": "%s", "timestamp": %s, "gasUsed": %d, "fee": "%s", "hasScResults": %t}},"upsert":%s}`,
		tx.Status, tx.MBHash, string(marshaledTimestamp), tx.GasUsed, tx.Fee, tx.HasSCR, string(marshaledTx)))

	return serializedData, nil
}
