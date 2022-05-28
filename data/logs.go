package data

import (
	"time"
)

// Logs holds all the fields needed for a logs structure
type Logs struct {
	ID             string        `json:"-" gorm:"primaryKey;unique"`
	OriginalTxHash string        `json:"originalTxHash,omitempty"`
	Address        string        `json:"address"`
	Events         []*Event      `json:"events" gorm:"foreignKey:Address"`
	Timestamp      time.Duration `json:"timestamp,omitempty"`
}

// Event holds all the fields needed for an event structure
type Event struct {
	Address    string   `json:"address" gorm:"primaryKey;unique"`
	Identifier string   `json:"identifier"`
	Topics     [][]byte `json:"topics" gorm:"serializer:base64"`
	Data       []byte   `json:"data" gorm:"serializer:base64"`
	Order      int      `json:"order"`
}

// PreparedLogsResults is the DTO that holds all the results after processing
type PreparedLogsResults struct {
	Tokens          TokensHandler
	TokensSupply    TokensHandler
	TagsCount       CountTags
	ScDeploys       map[string]*ScDeployInfo
	Delegators      map[string]*Delegator
	TokensInfo      []*TokenInfo
	NFTsDataUpdates []*NFTDataUpdate
	RolesData       RolesData
}
