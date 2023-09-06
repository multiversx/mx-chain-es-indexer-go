package data

import (
	"time"
)

// Block is a structure containing all the fields that need
//
//	to be saved for a block. It has all the default fields
//	plus some extra information for ease of search and filter
type Block struct {
	Nonce                 uint64                 `json:"nonce"`
	Round                 uint64                 `json:"round"`
	Epoch                 uint32                 `json:"epoch"`
	Hash                  string                 `json:"-"`
	MiniBlocksHashes      []string               `json:"miniBlocksHashes"`
	MiniBlocksDetails     []*MiniBlocksDetails   `json:"miniBlocksDetails,omitempty"`
	NotarizedBlocksHashes []string               `json:"notarizedBlocksHashes"`
	Proposer              uint64                 `json:"proposer"`
	Validators            []uint64               `json:"validators"`
	PubKeyBitmap          string                 `json:"pubKeyBitmap"`
	Size                  int64                  `json:"size"`
	SizeTxs               int64                  `json:"sizeTxs"`
	Timestamp             time.Duration          `json:"timestamp"`
	StateRootHash         string                 `json:"stateRootHash"`
	PrevHash              string                 `json:"prevHash"`
	ShardID               uint32                 `json:"shardId"`
	TxCount               uint32                 `json:"txCount"`
	NotarizedTxsCount     uint32                 `json:"notarizedTxsCount"`
	AccumulatedFees       string                 `json:"accumulatedFees"`
	DeveloperFees         string                 `json:"developerFees"`
	EpochStartBlock       bool                   `json:"epochStartBlock"`
	SearchOrder           uint64                 `json:"searchOrder"`
	EpochStartInfo        *EpochStartInfo        `json:"epochStartInfo,omitempty"`
	GasProvided           uint64                 `json:"gasProvided"`
	GasRefunded           uint64                 `json:"gasRefunded"`
	GasPenalized          uint64                 `json:"gasPenalized"`
	MaxGasLimit           uint64                 `json:"maxGasLimit"`
	ScheduledData         *ScheduledData         `json:"scheduledData,omitempty"`
	EpochStartShardsData  []*EpochStartShardData `json:"epochStartShardsData,omitempty"`
	RandSeed              string                 `json:"randSeed,omitempty"`
	PrevRandSeed          string                 `json:"prevRandSeed,omitempty"`
	Signature             string                 `json:"signature,omitempty"`
	LeaderSignature       string                 `json:"leaderSignature,omitempty"`
	ChainID               string                 `json:"chainID,omitempty"`
	SoftwareVersion       string                 `json:"softwareVersion,omitempty"`
	ReceiptsHash          string                 `json:"receiptsHash,omitempty"`
	Reserved              []byte                 `json:"reserved,omitempty"`
}

// MiniBlocksDetails is a structure that hold information about mini-blocks execution details
type MiniBlocksDetails struct {
	IndexFirstProcessedTx    int32    `json:"firstProcessedTx"`
	IndexLastProcessedTx     int32    `json:"lastProcessedTx"`
	SenderShardID            uint32   `json:"senderShard"`
	ReceiverShardID          uint32   `json:"receiverShard"`
	MBIndex                  int      `json:"mbIndex"`
	Type                     string   `json:"type"`
	ProcessingType           string   `json:"procType,omitempty"`
	TxsHashes                []string `json:"txsHashes,omitempty"`
	ExecutionOrderTxsIndices []int    `json:"executionOrderTxsIndices,omitempty"`
}

// ScheduledData is a structure that hold information about scheduled events
type ScheduledData struct {
	ScheduledRootHash        string `json:"rootHash,omitempty"`
	ScheduledAccumulatedFees string `json:"accumulatedFees,omitempty"`
	ScheduledDeveloperFees   string `json:"developerFees,omitempty"`
	ScheduledGasProvided     uint64 `json:"gasProvided,omitempty"`
	ScheduledGasPenalized    uint64 `json:"penalized,omitempty"`
	ScheduledGasRefunded     uint64 `json:"gasRefunded,omitempty"`
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

// EpochStartShardData is a structure that hold information about epoch start meta block shard data
type EpochStartShardData struct {
	ShardID                 uint32       `json:"shardID,omitempty"`
	Epoch                   uint32       `json:"epoch,omitempty"`
	Round                   uint64       `json:"round,omitempty"`
	Nonce                   uint64       `json:"nonce,omitempty"`
	HeaderHash              string       `json:"headerHash,omitempty"`
	RootHash                string       `json:"rootHash,omitempty"`
	ScheduledRootHash       string       `json:"scheduledRootHash,omitempty"`
	FirstPendingMetaBlock   string       `json:"firstPendingMetaBlock,omitempty"`
	LastFinishedMetaBlock   string       `json:"lastFinishedMetaBlock,omitempty"`
	PendingMiniBlockHeaders []*Miniblock `json:"pendingMiniBlockHeaders,omitempty"`
}

// Miniblock is a structure containing miniblock information
type Miniblock struct {
	Hash                        string        `json:"hash,omitempty"`
	SenderShardID               uint32        `json:"senderShard"`
	ReceiverShardID             uint32        `json:"receiverShard"`
	SenderBlockHash             string        `json:"senderBlockHash,omitempty"`
	ReceiverBlockHash           string        `json:"receiverBlockHash,omitempty"`
	Type                        string        `json:"type"`
	ProcessingTypeOnSource      string        `json:"procTypeS,omitempty"`
	ProcessingTypeOnDestination string        `json:"procTypeD,omitempty"`
	Timestamp                   time.Duration `json:"timestamp"`
	Reserved                    []byte        `json:"reserved,omitempty"`
}
