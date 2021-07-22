package data

import (
	"time"
)

// ValidatorsPublicKeys is a structure containing fields for validators public keys
type ValidatorsPublicKeys struct {
	PublicKeys []string `json:"publicKeys"`
}

// KibanaResponse -
type KibanaResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// ValidatorRatingInfo is a structure containing validator rating information
type ValidatorRatingInfo struct {
	PublicKey string  `json:"publicKey"`
	Rating    float32 `json:"rating"`
}

// TODO this will be removed in the next PR
// ValidatorsRatingInfo is a structure containing validators information
type ValidatorsRatingInfo struct {
	ValidatorsInfos []*ValidatorRatingInfo `json:"validatorsRating"`
}

// RoundInfo is a structure containing block signers and shard id
type RoundInfo struct {
	Index            uint64        `json:"round"`
	SignersIndexes   []uint64      `json:"signersIndexes"`
	BlockWasProposed bool          `json:"blockWasProposed"`
	ShardId          uint32        `json:"shardId"`
	Timestamp        time.Duration `json:"timestamp"`
}

// EpochInfo holds the information about epoch
type EpochInfo struct {
	AccumulatedFees string `json:"accumulatedFees"`
	DeveloperFees   string `json:"developerFees"`
}
