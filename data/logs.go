package data

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/tokeninfo"
)

// Logs holds all the fields needed for a logs structure
type Logs struct {
	ID             string        `json:"-"`
	OriginalTxHash string        `json:"originalTxHash,omitempty"`
	Address        string        `json:"address"`
	Events         []*Event      `json:"events"`
	Timestamp      time.Duration `json:"timestamp,omitempty"`
}

// Event holds all the fields needed for an event structure
type Event struct {
	Address    string   `json:"address"`
	Identifier string   `json:"identifier"`
	Topics     [][]byte `json:"topics"`
	Data       []byte   `json:"data"`
	Order      int      `json:"order"`
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
}
