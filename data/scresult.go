package data

import "time"

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
	Type              string        `json:"type,omitempty"`
	Status            string        `json:"status,omitempty"`
	Tokens            []string      `json:"tokens,omitempty"`
	ESDTValues        []string      `json:"esdtValues,omitempty"`
	Receivers         []string      `json:"receivers,omitempty"`
	ReceiversShardIDs []uint32      `json:"receiversShardIDs,omitempty"`
	Operation         string        `json:"operation"`
	Function          string        `json:"function,omitempty"`
}
