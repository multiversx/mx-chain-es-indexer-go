package data

import (
	"math/big"
	"time"
)

// Transaction is a structure containing all the fields that need
//  to be saved for a transaction. It has all the default fields
//  plus some extra information for ease of search and filter
type Transaction struct {
	MBHash               string        `json:"miniBlockHash"`
	Nonce                uint64        `json:"nonce"`
	Round                uint64        `json:"round"`
	Value                string        `json:"value"`
	Receiver             string        `json:"receiver"`
	Sender               string        `json:"sender"`
	ReceiverShard        uint32        `json:"receiverShard"`
	SenderShard          uint32        `json:"senderShard"`
	GasPrice             uint64        `json:"gasPrice"`
	GasLimit             uint64        `json:"gasLimit"`
	GasUsed              uint64        `json:"gasUsed"`
	Fee                  string        `json:"fee" gorm:"serializer:base64"`
	Data                 []byte        `json:"data" gorm:"serializer:base64"`
	Signature            string        `json:"signature" gorm:"serializer:base64"`
	Timestamp            time.Duration `json:"timestamp"`
	Status               string        `json:"status" gorm:"serializer:base64"`
	SearchOrder          uint32        `json:"searchOrder"`
	SenderUserName       []byte        `json:"senderUserName,omitempty" gorm:"serializer:base64"`
	ReceiverUserName     []byte        `json:"receiverUserName,omitempty" gorm:"serializer:base64"`
	HasSCR               bool          `json:"hasScResults,omitempty"`
	IsScCall             bool          `json:"isScCall,omitempty"`
	HasOperations        bool          `json:"hasOperations,omitempty"`
	Tokens               []string      `json:"tokens,omitempty" gorm:"type:text[]"`
	ESDTValues           []string      `json:"esdtValues,omitempty" gorm:"type:text[]"`
	Receivers            []string      `json:"receivers,omitempty" gorm:"type:text[]"`
	ReceiversShardIDs    []uint32      `json:"receiversShardIDs,omitempty" gorm:"serializer:json"`
	Type                 string        `json:"type,omitempty" gorm:"serializer:base64"`
	Operation            string        `json:"operation,omitempty" gorm:"serializer:base64"`
	Function             string        `json:"function,omitempty" gorm:"serializer:base64"`
	IsRelayed            bool          `json:"isRelayed,omitempty"`
	Version              uint32        `json:"version,omitempty"`
	SmartContractResults []*ScResult   `json:"-" gorm:"-"`
	ReceiverAddressBytes []byte        `json:"-" gorm:"serializer:base64"`
	Hash                 string        `json:"-" gorm:"primaryKey;unique"`
	BlockHash            string        `json:"-" gorm:"serializer:base64"`
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
	Hash      string        `json:"-" gorm:"primaryKey;unique"`
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
	AlteredAccts AlteredAccountsHandler
	TxHashStatus map[string]string
	TxHashRefund map[string]*RefundData
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
