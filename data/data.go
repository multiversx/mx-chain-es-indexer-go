package data

import (
	"encoding/json"
	"time"
)

// ValidatorsPublicKeys is a structure containing fields for validators public keys
type ValidatorsPublicKeys struct {
	PublicKeys []string `json:"publicKeys"`
}

// Response is a structure that holds response from Kibana
type Response struct {
	Error  interface{} `json:"error,omitempty"`
	Status int         `json:"status"`
}

// ValidatorRatingInfo is a structure containing validator rating information
type ValidatorRatingInfo struct {
	PublicKey string  `json:"-"`
	Rating    float32 `json:"rating"`
}

// RoundInfo is a structure containing block signers and shard id
type RoundInfo struct {
	Round            uint64        `json:"round"`
	SignersIndexes   []uint64      `json:"signersIndexes"`
	BlockWasProposed bool          `json:"blockWasProposed"`
	ShardId          uint32        `json:"shardId"`
	Epoch            uint32        `json:"epoch"`
	Timestamp        time.Duration `json:"timestamp"`
}

// EpochInfo holds the information about epoch
type EpochInfo struct {
	AccumulatedFees string `json:"accumulatedFees"`
	DeveloperFees   string `json:"developerFees"`
}

// ResponseScroll defines the generic structure for an Elasticsearch scroll request
type ResponseScroll struct {
	Hits struct {
		Hits []struct {
			ID     string          `json:"_id"`
			Source json.RawMessage `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// KeyValueObj is the dto for values index
type KeyValueObj struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
