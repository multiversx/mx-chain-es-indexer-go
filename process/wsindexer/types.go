package wsindexer

import (
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/receipt"
	"github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
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
