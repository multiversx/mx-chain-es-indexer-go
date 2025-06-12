package data

import (
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
)

// AccountInfo holds (serializable) data about an account
type AccountInfo struct {
	Address             string         `json:"address,omitempty"`
	Nonce               uint64         `json:"nonce,omitempty"`
	Balance             string         `json:"balance"`
	BalanceNum          float64        `json:"balanceNum"`
	TokenName           string         `json:"token,omitempty"`
	TokenIdentifier     string         `json:"identifier,omitempty"`
	TokenNonce          uint64         `json:"tokenNonce,omitempty"`
	Properties          string         `json:"properties,omitempty"`
	Frozen              bool           `json:"frozen,omitempty"`
	Owner               string         `json:"owner,omitempty"`
	UserName            string         `json:"userName,omitempty"`
	DeveloperRewards    string         `json:"developerRewards,omitempty"`
	DeveloperRewardsNum float64        `json:"developerRewardsNum,omitempty"`
	Data                *TokenMetaData `json:"data,omitempty"`
	Timestamp           uint64         `json:"timestamp,omitempty"`
	TimestampMs         uint64         `json:"timestampMs,omitempty"`
	Type                string         `json:"type,omitempty"`
	CurrentOwner        string         `json:"currentOwner,omitempty"`
	ShardID             uint32         `json:"shardID"`
	RootHash            []byte         `json:"rootHash,omitempty"`
	CodeHash            []byte         `json:"codeHash,omitempty"`
	CodeMetadata        []byte         `json:"codeMetadata,omitempty"`
	IsSender            bool           `json:"-"`
	IsSmartContract     bool           `json:"-"`
	IsNFTCreate         bool           `json:"-"`
}

// TokenMetaData holds data about a token metadata
type TokenMetaData struct {
	Name               string   `json:"name,omitempty"`
	Creator            string   `json:"creator,omitempty"`
	Royalties          uint32   `json:"royalties,omitempty"`
	Hash               []byte   `json:"hash,omitempty"`
	URIs               [][]byte `json:"uris,omitempty"`
	Tags               []string `json:"tags,omitempty"`
	Attributes         []byte   `json:"attributes,omitempty"`
	MetaData           string   `json:"metadata,omitempty"`
	NonEmptyURIs       bool     `json:"nonEmptyURIs"`
	WhiteListedStorage bool     `json:"whiteListedStorage"`
}

// AccountBalanceHistory represents an entry in the user accounts balances history
type AccountBalanceHistory struct {
	Address         string `json:"address"`
	Timestamp       uint64 `json:"timestamp"`
	TimestampMs     uint64 `json:"timestampMs,omitempty"`
	Balance         string `json:"balance"`
	Token           string `json:"token,omitempty"`
	Identifier      string `json:"identifier,omitempty"`
	TokenNonce      uint64 `json:"tokenNonce,omitempty"`
	IsSender        bool   `json:"isSender,omitempty"`
	IsSmartContract bool   `json:"isSmartContract,omitempty"`
	ShardID         uint32 `json:"shardID"`
}

// Account is a structure that is needed for regular accounts
type Account struct {
	UserAccount *alteredAccount.AlteredAccount
	IsSender    bool
}

// AccountESDT is a structure that is needed for ESDT accounts
type AccountESDT struct {
	Account         *alteredAccount.AlteredAccount
	TokenIdentifier string
	NFTNonce        uint64
	IsSender        bool
	IsNFTOperation  bool
	IsNFTCreate     bool
}
