package data

import "time"

// ScResult is a structure containing all the fields that need to be saved for a smart contract result
type ScResult struct {
	Hash               string        `json:"-" gorm:"primaryKey;unique"`
	MBHash             string        `json:"miniBlockHash,omitempty"`
	Nonce              uint64        `json:"nonce"`
	GasLimit           uint64        `json:"gasLimit"`
	GasPrice           uint64        `json:"gasPrice"`
	Value              string        `json:"value"`
	Sender             string        `json:"sender"`
	Receiver           string        `json:"receiver"`
	SenderShard        uint32        `json:"senderShard"`
	ReceiverShard      uint32        `json:"receiverShard"`
	RelayerAddr        string        `json:"relayerAddr,omitempty"`
	RelayedValue       string        `json:"relayedValue,omitempty"`
	Code               string        `json:"code,omitempty" gorm:"serializer:base64"`
	Data               []byte        `json:"data,omitempty" gorm:"serializer:base64"`
	PrevTxHash         string        `json:"prevTxHash"`
	OriginalTxHash     string        `json:"originalTxHash"`
	CallType           string        `json:"callType"`
	CodeMetadata       []byte        `json:"codeMetaData,omitempty" gorm:"serializer:base64"`
	ReturnMessage      string        `json:"returnMessage,omitempty" gorm:"serializer:base64"`
	Timestamp          time.Duration `json:"timestamp"`
	HasOperations      bool          `json:"hasOperations,omitempty"`
	Type               string        `json:"type,omitempty"`
	Status             string        `json:"status,omitempty"`
	Tokens             []string      `json:"tokens,omitempty" gorm:"serializer:json"`
	ESDTValues         []string      `json:"esdtValues,omitempty" gorm:"serializer:json"`
	Receivers          []string      `json:"receivers,omitempty" gorm:"serializer:json"`
	ReceiversShardIDs  []uint32      `json:"receiversShardIDs,omitempty" gorm:"serializer:json"`
	Operation          string        `json:"operation,omitempty"`
	Function           string        `json:"function,omitempty"`
	IsRelayed          bool          `json:"isRelayed,omitempty"`
	CanBeIgnored       bool          `json:"canBeIgnored,omitempty"`
	OriginalSender     string        `json:"originalSender,omitempty"`
	SenderAddressBytes []byte        `json:"-" gorm:"serializer:base64"`
}
