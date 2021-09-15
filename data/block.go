package data

import (
	"time"
)

// Block is a structure containing all the fields that need
//  to be saved for a block. It has all the default fields
//  plus some extra information for ease of search and filter
type Block struct {
	Nonce                 uint64          `json:"nonce"`
	Round                 uint64          `json:"round"`
	Epoch                 uint32          `json:"epoch"`
	Hash                  string          `json:"-"`
	MiniBlocksHashes      []string        `json:"miniBlocksHashes"`
	NotarizedBlocksHashes []string        `json:"notarizedBlocksHashes"`
	Proposer              uint64          `json:"proposer"`
	Validators            []uint64        `json:"validators"`
	PubKeyBitmap          string          `json:"pubKeyBitmap"`
	Size                  int64           `json:"size"`
	SizeTxs               int64           `json:"sizeTxs"`
	Timestamp             time.Duration   `json:"timestamp"`
	StateRootHash         string          `json:"stateRootHash"`
	PrevHash              string          `json:"prevHash"`
	ShardID               uint32          `json:"shardId"`
	TxCount               uint32          `json:"txCount"`
	AccumulatedFees       string          `json:"accumulatedFees"`
	DeveloperFees         string          `json:"developerFees"`
	EpochStartBlock       bool            `json:"epochStartBlock"`
	SearchOrder           uint64          `json:"searchOrder"`
	EpochStartInfo        *EpochStartInfo `json:"epochStartInfo,omitempty"`
	GasConsumed           uint64          `json:"gasConsumed"`
	MaxGasLimit           uint64          `json:"maxGasLimit"`
}

// EpochStartInfo is a structure that hold information about epoch start meta block
type EpochStartInfo struct {
	TotalSupply                      string `json:"totalSupply"`
	TotalToDistribute                string `json:"totalToDistribute"`
	TotalNewlyMinted                 string `json:"totalNewlyMinted"`
	RewardsPerBlock                  string `json:"rewardsPerBlock"`
	RewardsForProtocolSustainability string `json:"rewardsForProtocolSustainability"`
	NodePrice                        string `json:"nodePrice"`
	PrevEpochStartRound              uint64 `json:"prevEpochStartRound"`
	PrevEpochStartHash               string `json:"prevEpochStartHash"`
}

// Miniblock is a structure containing miniblock information
type Miniblock struct {
	Hash              string        `json:"-"`
	SenderShardID     uint32        `json:"senderShard"`
	ReceiverShardID   uint32        `json:"receiverShard"`
	SenderBlockHash   string        `json:"senderBlockHash"`
	ReceiverBlockHash string        `json:"receiverBlockHash"`
	Type              string        `json:"type"`
	Timestamp         time.Duration `json:"timestamp"`
}
