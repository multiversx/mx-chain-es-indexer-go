package wsindexer

import (
	"errors"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/receipt"
	"github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	outportData "github.com/ElrondNetwork/elrond-go-core/websocketOutportDriver/data"
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
	type normalTxWrapped struct {
		TransactionHandler *transaction.Transaction
		outport.FeeInfo
	}
	type rewardsTxsWrapped struct {
		TransactionHandler *rewardTx.RewardTx
		outport.FeeInfo
	}
	type scrWrapped struct {
		TransactionHandler *smartContractResult.SmartContractResult
		outport.FeeInfo
	}
	type receiptWrapped struct {
		TransactionHandler *receipt.Receipt
		outport.FeeInfo
	}
	type logWrapped struct {
		TxHash string
		transaction.Log
	}

	type poolStruct struct {
		Txs      map[string]*normalTxWrapped
		Invalid  map[string]*normalTxWrapped
		Scrs     map[string]*scrWrapped
		Rewards  map[string]*rewardsTxsWrapped
		Receipts map[string]*receiptWrapped
		Logs     []*logWrapped
	}

	argSaveBlock := struct {
		TransactionsPool *poolStruct
	}{}

	err := i.marshaller.Unmarshal(&argSaveBlock, marshaledData)
	if err != nil {
		return nil, err
	}

	normalTxs := make(map[string]data.TransactionHandlerWithGasUsedAndFee, len(argSaveBlock.TransactionsPool.Txs))
	for txHash, tx := range argSaveBlock.TransactionsPool.Txs {
		normalTxs[txHash] = outport.NewTransactionHandlerWithGasAndFee(tx.TransactionHandler, tx.GasUsed, tx.Fee)
		normalTxs[txHash].SetInitialPaidFee(tx.InitialPaidFee)
	}

	invalidTxs := make(map[string]data.TransactionHandlerWithGasUsedAndFee, len(argSaveBlock.TransactionsPool.Invalid))
	for txHash, tx := range argSaveBlock.TransactionsPool.Invalid {
		invalidTxs[txHash] = outport.NewTransactionHandlerWithGasAndFee(tx.TransactionHandler, tx.GasUsed, tx.Fee)
		invalidTxs[txHash].SetInitialPaidFee(tx.InitialPaidFee)
	}

	scrs := make(map[string]data.TransactionHandlerWithGasUsedAndFee, len(argSaveBlock.TransactionsPool.Scrs))
	for txHash, tx := range argSaveBlock.TransactionsPool.Scrs {
		scrs[txHash] = outport.NewTransactionHandlerWithGasAndFee(tx.TransactionHandler, tx.GasUsed, tx.Fee)
		scrs[txHash].SetInitialPaidFee(tx.InitialPaidFee)
	}

	receipts := make(map[string]data.TransactionHandlerWithGasUsedAndFee, len(argSaveBlock.TransactionsPool.Receipts))
	for txHash, tx := range argSaveBlock.TransactionsPool.Receipts {
		receipts[txHash] = outport.NewTransactionHandlerWithGasAndFee(tx.TransactionHandler, tx.GasUsed, tx.Fee)
		receipts[txHash].SetInitialPaidFee(tx.InitialPaidFee)
	}

	rewards := make(map[string]data.TransactionHandlerWithGasUsedAndFee, len(argSaveBlock.TransactionsPool.Rewards))
	for txHash, tx := range argSaveBlock.TransactionsPool.Rewards {
		rewards[txHash] = outport.NewTransactionHandlerWithGasAndFee(tx.TransactionHandler, tx.GasUsed, tx.Fee)
		rewards[txHash].SetInitialPaidFee(tx.InitialPaidFee)
	}

	logs := make([]*data.LogData, 0, len(argSaveBlock.TransactionsPool.Logs))
	for _, txLog := range argSaveBlock.TransactionsPool.Logs {
		logs = append(logs, &data.LogData{
			LogHandler: txLog,
			TxHash:     txLog.TxHash,
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
		HeaderType outportData.HeaderType
	}{}

	err := i.marshaller.Unmarshal(&headerTypeStruct, marshaledData)
	if err != nil {
		return nil, err
	}

	switch headerTypeStruct.HeaderType {
	case outportData.MetaHeader:
		hStruct := struct {
			H1 *block.Header `json:"Header"`
		}{}
		err = i.marshaller.Unmarshal(&hStruct, marshaledData)
		return hStruct.H1, err
	case outportData.ShardHeaderV1:
		hStruct := struct {
			H1 *block.MetaBlock `json:"Header"`
		}{}
		err = i.marshaller.Unmarshal(&hStruct, marshaledData)
		return hStruct.H1, err
	case outportData.ShardHeaderV2:
		hStruct := struct {
			H1 *block.HeaderV2 `json:"Header"`
		}{}
		err = i.marshaller.Unmarshal(&hStruct, marshaledData)
		return hStruct.H1, err
	default:
		return nil, errors.New("invalid header type")
	}
}
