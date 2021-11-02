package data

import (
	"math/big"
	"time"
)

// Transaction is a structure containing all the fields that need
//  to be saved for a transaction. It has all the default fields
//  plus some extra information for ease of search and filter
type Transaction struct {
	MBHash               string               `json:"miniBlockHash"`
	Nonce                uint64               `json:"nonce"`
	Round                uint64               `json:"round"`
	Value                string               `json:"value"`
	Receiver             string               `json:"receiver"`
	Sender               string               `json:"sender"`
	ReceiverShard        uint32               `json:"receiverShard"`
	SenderShard          uint32               `json:"senderShard"`
	GasPrice             uint64               `json:"gasPrice"`
	GasLimit             uint64               `json:"gasLimit"`
	GasUsed              uint64               `json:"gasUsed"`
	Fee                  string               `json:"fee"`
	Data                 []byte               `json:"data"`
	Signature            string               `json:"signature"`
	Timestamp            time.Duration        `json:"timestamp"`
	Status               string               `json:"status"`
	SearchOrder          uint32               `json:"searchOrder"`
	SenderUserName       []byte               `json:"senderUserName,omitempty"`
	ReceiverUserName     []byte               `json:"receiverUserName,omitempty"`
	HasSCR               bool                 `json:"hasScResults,omitempty"`
	IsScCall             bool                 `json:"isScCall,omitempty"`
	HasOperations        bool                 `json:"hasOperations,omitempty"`
	Tokens               []string             `json:"tokens,omitempty"`
	ESDTValues           []string             `json:"esdtValues,omitempty"`
	SmartContractResults map[string]*ScResult `json:"scresults,omitempty"`
	Logs                 []*Logs              `json:"logs,omitempty"`
	ReceiverAddressBytes []byte               `json:"-"`
	Hash                 string               `json:"-"`
	BlockHash            string               `json:"-"`
	Receivers            []string             `json:"receivers,omitempty"`
	ReceiversShardIDs    []uint32             `json:"receiversShardIDs,omitempty"`
	Type                 string               `json:"type,omitempty"`
}

// GetGasLimit will return transaction gas limit
func (t *Transaction) GetGasLimit() uint64 {
	return t.GasLimit
}

// GetGasPrice will return transaction gas price
func (t *Transaction) GetGasPrice() uint64 {
	return t.GasPrice
}

// GetData will return transaction data field
func (t *Transaction) GetData() []byte {
	return t.Data
}

// GetRcvAddr will return transaction receiver address
func (t *Transaction) GetRcvAddr() []byte {
	return t.ReceiverAddressBytes
}

// GetValue wil return transaction value
func (t *Transaction) GetValue() *big.Int {
	bigIntValue, ok := big.NewInt(0).SetString(t.Value, 10)
	if !ok {
		return big.NewInt(0)
	}

	return bigIntValue
}

// Receipt is a structure containing all the fields that need to be save for a Receipt
type Receipt struct {
	Hash      string        `json:"-"`
	Value     string        `json:"value"`
	Sender    string        `json:"sender"`
	Data      string        `json:"data,omitempty"`
	TxHash    string        `json:"txHash"`
	Timestamp time.Duration `json:"timestamp"`
}

// ScResult is a structure containing all the fields that need to be saved for a smart contract result
type ScResult struct {
	Hash              string        `json:"-"`
	MBHash            string        `json:"miniBlockHash,omitempty"`
	Nonce             uint64        `json:"nonce"`
	GasLimit          uint64        `json:"gasLimit"`
	GasPrice          uint64        `json:"gasPrice"`
	Value             string        `json:"value"`
	Sender            string        `json:"sender"`
	Receiver          string        `json:"receiver"`
	SenderShard       uint32        `json:"senderShard"`
	ReceiverShard     uint32        `json:"receiverShard"`
	RelayerAddr       string        `json:"relayerAddr,omitempty"`
	RelayedValue      string        `json:"relayedValue,omitempty"`
	Code              string        `json:"code,omitempty"`
	Data              []byte        `json:"data,omitempty"`
	PrevTxHash        string        `json:"prevTxHash"`
	OriginalTxHash    string        `json:"originalTxHash"`
	CallType          string        `json:"callType"`
	CodeMetadata      []byte        `json:"codeMetaData,omitempty"`
	ReturnMessage     string        `json:"returnMessage,omitempty"`
	Timestamp         time.Duration `json:"timestamp"`
	HasOperations     bool          `json:"hasOperations,omitempty"`
	Tokens            []string      `json:"tokens,omitempty"`
	ESDTValues        []string      `json:"esdtValues,omitempty"`
	Logs              []*Logs       `json:"logs,omitempty"`
	Receivers         []string      `json:"receivers,omitempty"`
	ReceiversShardIDs []uint32      `json:"receiversShardIDs,omitempty"`
	Status            string        `json:"status,omitempty"`
	Type              string        `json:"type,omitempty"`
}

// PreparedResults is the DTO that holds all the results after processing
type PreparedResults struct {
	Transactions []*Transaction
	ScResults    []*ScResult
	Receipts     []*Receipt
	AlteredAccts AlteredAccountsHandler
	TxHashStatus map[string]string
	TxHashRefund map[string]*RefundData
	SCRSNoTx     []*ScResult
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

// RefundData is the structure that contains data about a refund
type RefundData struct {
	Value    string
	Receiver string
}
