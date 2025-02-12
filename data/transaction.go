package data

import (
	"time"
)

// Transaction is a structure containing all the fields that need
// to be saved for a transaction. It has all the default fields
// plus some extra information for ease of search and filter
type Transaction struct {
	MBHash               string        `json:"miniBlockHash"`
	Nonce                uint64        `json:"nonce"`
	Round                uint64        `json:"round"`
	Value                string        `json:"value"`
	ValueNum             float64       `json:"valueNum"`
	Receiver             string        `json:"receiver"`
	Sender               string        `json:"sender"`
	ReceiverShard        uint32        `json:"receiverShard"`
	SenderShard          uint32        `json:"senderShard"`
	GasPrice             uint64        `json:"gasPrice"`
	GasLimit             uint64        `json:"gasLimit"`
	GasUsed              uint64        `json:"gasUsed"`
	Fee                  string        `json:"fee"`
	FeeNum               float64       `json:"feeNum"`
	InitialPaidFee       string        `json:"initialPaidFee,omitempty"`
	Data                 []byte        `json:"data"`
	Signature            string        `json:"signature"`
	Timestamp            time.Duration `json:"timestamp"`
	Status               string        `json:"status"`
	SearchOrder          uint32        `json:"searchOrder"`
	SenderUserName       []byte        `json:"senderUserName,omitempty"`
	ReceiverUserName     []byte        `json:"receiverUserName,omitempty"`
	HasSCR               bool          `json:"hasScResults,omitempty"`
	IsScCall             bool          `json:"isScCall,omitempty"`
	HasOperations        bool          `json:"hasOperations,omitempty"`
	HasLogs              bool          `json:"hasLogs,omitempty"`
	Tokens               []string      `json:"tokens,omitempty"`
	ESDTValues           []string      `json:"esdtValues,omitempty"`
	ESDTValuesNum        []float64     `json:"esdtValuesNum,omitempty"`
	Receivers            []string      `json:"receivers,omitempty"`
	ReceiversShardIDs    []uint32      `json:"receiversShardIDs,omitempty"`
	Type                 string        `json:"type,omitempty"`
	Operation            string        `json:"operation,omitempty"`
	Function             string        `json:"function,omitempty"`
	IsRelayed            bool          `json:"isRelayed,omitempty"`
	Version              uint32        `json:"version,omitempty"`
	GuardianAddress      string        `json:"guardian,omitempty"`
	GuardianSignature    string        `json:"guardianSignature,omitempty"`
	ErrorEvent           bool          `json:"errorEvent,omitempty"`
	CompletedEvent       bool          `json:"completedEvent,omitempty"`
	RelayedAddr          string        `json:"relayer,omitempty"`
	RelayedSignature     string        `json:"relayerSignature,omitempty"`
	HadRefund            bool          `json:"hadRefund,omitempty"`
	ExecutionOrder       int           `json:"-"`
	SmartContractResults []*ScResult   `json:"-"`
	Hash                 string        `json:"-"`
	BlockHash            string        `json:"-"`
}

// Receipt is a structure containing all the fields that need to be safe for a Receipt
type Receipt struct {
	Hash      string        `json:"-"`
	Value     string        `json:"value"`
	Sender    string        `json:"sender"`
	Data      string        `json:"data,omitempty"`
	TxHash    string        `json:"txHash"`
	Timestamp time.Duration `json:"timestamp"`
}

// PreparedResults is the DTO that holds all the results after processing
type PreparedResults struct {
	Transactions []*Transaction
	ScResults    []*ScResult
	Receipts     []*Receipt
	TxHashFee    map[string]*FeeData
}

// ResponseTransactions is the structure for the transactions response
type ResponseTransactions struct {
	Docs []*ResponseTransactionDB `json:"docs"`
}

// ResponseTransactionDB is the structure for the transaction response
type ResponseTransactionDB struct {
	Found  bool        `json:"found"`
	ID     string      `json:"_id"`
	Source Transaction `json:"_source"`
}

// FeeData is the structure that contains data about transaction fee and gas used
type FeeData struct {
	FeeNum      float64
	Fee         string
	GasUsed     uint64
	Receiver    string
	GasRefunded uint64
}
