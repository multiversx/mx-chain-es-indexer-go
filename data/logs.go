package data

import "github.com/ElrondNetwork/elastic-indexer-go/process/tags"

// Logs holds all the fields needed for a logs structure
type Logs struct {
	ID      string   `json:"-"`
	Address string   `json:"address"`
	Events  []*Event `json:"events"`
}

// Event holds all the fields needed for an event structure
type Event struct {
	Address    string   `json:"address"`
	Identifier string   `json:"identifier"`
	Topics     [][]byte `json:"topics"`
	Data       []byte   `json:"data"`
}

// PreparedLogsResults is the DTO that holds all the results after processing
type PreparedLogsResults struct {
	Tokens          TokensHandler
	TagsCount       tags.CountTags
	ScDeploys       map[string]*ScDeployInfo
	PendingBalances map[string]*AccountInfo
}
