package data

import (
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/tokeninfo"
)

// Logs holds all the fields needed for a logs structure
type Logs struct {
	UUID           string   `json:"uuid"`
	ID             string   `json:"-"`
	OriginalTxHash string   `json:"originalTxHash,omitempty"`
	Address        string   `json:"address"`
	Events         []*Event `json:"events"`
	Timestamp      uint64   `json:"timestamp,omitempty"`
	TimestampMs    uint64   `json:"timestampMs,omitempty"`
}

// Event holds all the fields needed for an event structure
type Event struct {
	Address        string   `json:"address"`
	Identifier     string   `json:"identifier"`
	Topics         [][]byte `json:"topics"`
	Data           []byte   `json:"data"`
	AdditionalData [][]byte `json:"additionalData,omitempty"`
	Order          int      `json:"order"`
}

// PreparedLogsResults is the DTO that holds all the results after processing
type PreparedLogsResults struct {
	Tokens                  TokensHandler
	TokensSupply            TokensHandler
	ScDeploys               map[string]*ScDeployInfo
	ChangeOwnerOperations   map[string]*OwnerData
	Delegators              map[string]*Delegator
	TxHashStatusInfo        map[string]*outport.StatusInfo
	TokensInfo              []*TokenInfo
	NFTsDataUpdates         []*NFTDataUpdate
	TokenRolesAndProperties *tokeninfo.TokenRolesAndProperties
	DBLogs                  []*Logs
	DBEvents                []*LogEvent
}
