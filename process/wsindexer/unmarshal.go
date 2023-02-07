package wsindexer

import (
	"encoding/hex"
	"errors"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
)

func (i *indexer) getArgsSaveBlock(marshaledData []byte) (*outport.ArgsSaveBlockData, error) {
	defer func(start time.Time) {
		log.Debug("indexer.getArgsSaveBlock", "duration", time.Since(start))
	}(time.Now())

	header, err := i.getHeader(marshaledData)
	if err != nil {
		return nil, err
	}

	body, err := i.getBody(marshaledData)
	if err != nil {
		return nil, err
	}

	txsPool, err := i.getTxsPool(marshaledData)
	if err != nil {
		return nil, err
	}

	argsBlockS := struct {
		HeaderHash             []byte
		SignersIndexes         []uint64
		NotarizedHeadersHashes []string
		HeaderGasConsumption   outport.HeaderGasConsumption
		AlteredAccounts        map[string]*outport.AlteredAccount
		NumberOfShards         uint32
		IsImportDB             bool
	}{}
	err = i.marshaller.Unmarshal(&argsBlockS, marshaledData)
	if err != nil {
		return nil, err
	}

	return &outport.ArgsSaveBlockData{
		HeaderHash:             argsBlockS.HeaderHash,
		Body:                   body,
		Header:                 header,
		SignersIndexes:         argsBlockS.SignersIndexes,
		NotarizedHeadersHashes: argsBlockS.NotarizedHeadersHashes,
		HeaderGasConsumption:   argsBlockS.HeaderGasConsumption,
		TransactionsPool:       txsPool,
		AlteredAccounts:        argsBlockS.AlteredAccounts,
		NumberOfShards:         argsBlockS.NumberOfShards,
		IsImportDB:             argsBlockS.IsImportDB,
	}, nil
}

func (i *indexer) getTxsPool(marshaledData []byte) (*outport.Pool, error) {
	argSaveBlock := argsSaveBlock{}

	err := i.marshaller.Unmarshal(&argSaveBlock, marshaledData)
	if err != nil {
		return nil, err
	}

	normalTxs := make(map[string]data.TransactionHandlerWithGasUsedAndFee, len(argSaveBlock.TransactionsPool.Txs))
	for txHash, tx := range argSaveBlock.TransactionsPool.Txs {
		decoded := getDecodedHash(txHash)

		normalTxs[decoded] = outport.NewTransactionHandlerWithGasAndFee(tx.TransactionHandler, tx.GasUsed, tx.Fee)
		normalTxs[decoded].SetInitialPaidFee(tx.InitialPaidFee)
	}

	invalidTxs := make(map[string]data.TransactionHandlerWithGasUsedAndFee, len(argSaveBlock.TransactionsPool.Invalid))
	for txHash, tx := range argSaveBlock.TransactionsPool.Invalid {
		decoded := getDecodedHash(txHash)

		invalidTxs[decoded] = outport.NewTransactionHandlerWithGasAndFee(tx.TransactionHandler, tx.GasUsed, tx.Fee)
		invalidTxs[decoded].SetInitialPaidFee(tx.InitialPaidFee)
	}

	scrs := make(map[string]data.TransactionHandlerWithGasUsedAndFee, len(argSaveBlock.TransactionsPool.Scrs))
	for txHash, tx := range argSaveBlock.TransactionsPool.Scrs {
		decoded := getDecodedHash(txHash)

		scrs[decoded] = outport.NewTransactionHandlerWithGasAndFee(tx.TransactionHandler, tx.GasUsed, tx.Fee)
		scrs[decoded].SetInitialPaidFee(tx.InitialPaidFee)
	}

	receipts := make(map[string]data.TransactionHandlerWithGasUsedAndFee, len(argSaveBlock.TransactionsPool.Receipts))
	for txHash, tx := range argSaveBlock.TransactionsPool.Receipts {
		decoded := getDecodedHash(txHash)

		receipts[decoded] = outport.NewTransactionHandlerWithGasAndFee(tx.TransactionHandler, tx.GasUsed, tx.Fee)
		receipts[decoded].SetInitialPaidFee(tx.InitialPaidFee)
	}

	rewards := make(map[string]data.TransactionHandlerWithGasUsedAndFee, len(argSaveBlock.TransactionsPool.Rewards))
	for txHash, tx := range argSaveBlock.TransactionsPool.Rewards {
		decoded := getDecodedHash(txHash)

		rewards[decoded] = outport.NewTransactionHandlerWithGasAndFee(tx.TransactionHandler, tx.GasUsed, tx.Fee)
		rewards[decoded].SetInitialPaidFee(tx.InitialPaidFee)
	}

	logs := make([]*data.LogData, 0, len(argSaveBlock.TransactionsPool.Logs))
	for _, txLog := range argSaveBlock.TransactionsPool.Logs {
		logs = append(logs, &data.LogData{
			LogHandler: txLog.LogHandler,
			TxHash:     getDecodedHash(txLog.TxHash),
		})
	}

	return &outport.Pool{
		Txs:      normalTxs,
		Scrs:     scrs,
		Rewards:  rewards,
		Invalid:  invalidTxs,
		Receipts: receipts,
		Logs:     logs,
	}, nil
}

func getDecodedHash(hash string) string {
	decoded, err := hex.DecodeString(hash)
	if err != nil {
		log.Warn("getDecodedHash.cannot decode hash", "error", err, "hash", hash)
		return hash
	}
	return string(decoded)
}

func (i *indexer) getHeaderAndBody(marshaledData []byte) (data.HeaderHandler, data.BodyHandler, error) {
	body, err := i.getBody(marshaledData)
	if err != nil {
		return nil, nil, err
	}

	header, err := i.getHeader(marshaledData)
	if err != nil {
		return nil, nil, err
	}

	return header, body, nil
}

func (i *indexer) getBody(marshaledData []byte) (data.BodyHandler, error) {
	bodyStruct := struct {
		Body *block.Body
	}{}

	err := i.marshaller.Unmarshal(&bodyStruct, marshaledData)
	return bodyStruct.Body, err
}

func (i *indexer) getHeader(marshaledData []byte) (data.HeaderHandler, error) {
	headerTypeStruct := struct {
		HeaderType core.HeaderType
	}{}

	err := i.marshaller.Unmarshal(&headerTypeStruct, marshaledData)
	if err != nil {
		return nil, err
	}

	switch headerTypeStruct.HeaderType {
	case core.MetaHeader:
		hStruct := struct {
			H1 *block.MetaBlock `json:"Header"`
		}{}
		err = i.marshaller.Unmarshal(&hStruct, marshaledData)
		return hStruct.H1, err
	case core.ShardHeaderV1:
		hStruct := struct {
			H1 *block.Header `json:"Header"`
		}{}
		err = i.marshaller.Unmarshal(&hStruct, marshaledData)
		return hStruct.H1, err
	case core.ShardHeaderV2:
		hStruct := struct {
			H1 *block.HeaderV2 `json:"Header"`
		}{}
		err = i.marshaller.Unmarshal(&hStruct, marshaledData)
		return hStruct.H1, err
	default:
		return nil, errors.New("invalid header type")
	}
}
