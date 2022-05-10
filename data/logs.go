package data

import (
	"time"
)

// Logs holds all the fields needed for a logs structure
type Logs struct {
	ID        string        `json:"-"`
	Address   string        `json:"address"`
	Events    []*Event      `json:"events" gorm:"foreignKey:Address"`
	Timestamp time.Duration `json:"timestamp,omitempty"`
}

// Event holds all the fields needed for an event structure
type Event struct {
	Address    string   `json:"address"`
	Identifier string   `json:"identifier"`
	Topics     [][]byte `json:"topics" gorm:"serializer:json"`
	Data       []byte   `json:"data"`
	Order      int      `json:"order"`
}

// PreparedLogsResults is the DTO that holds all the results after processing
type PreparedLogsResults struct {
	Tokens          TokensHandler
	TagsCount       CountTags
	ScDeploys       map[string]*ScDeployInfo
	PendingBalances map[string]*AccountInfo
	Delegators      map[string]*Delegator
	TokensInfo      []*TokenInfo
}
