package wsindexer

import (
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/receipt"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

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
	TxHash     string
	LogHandler *transaction.Log
}

type poolStruct struct {
	Txs      map[string]*normalTxWrapped
	Invalid  map[string]*normalTxWrapped
	Scrs     map[string]*scrWrapped
	Rewards  map[string]*rewardsTxsWrapped
	Receipts map[string]*receiptWrapped
	Logs     []*logWrapped
}

type argsSaveBlock struct {
	TransactionsPool *poolStruct
}
