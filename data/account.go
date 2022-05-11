package data

import (
	"time"

	coreData "github.com/ElrondNetwork/elrond-go-core/data"
)

// AccountInfo holds (serializable) data about an account
type AccountInfo struct {
	Address                  string         `json:"address,omitempty" gorm:"primaryKey;unique"`
	Nonce                    uint64         `json:"nonce,omitempty"`
	Balance                  string         `json:"balance"`
	BalanceNum               float64        `json:"balanceNum"`
	TokenName                string         `json:"token,omitempty"`
	TokenIdentifier          string         `json:"identifier,omitempty"`
	TokenNonce               uint64         `json:"tokenNonce,omitempty"`
	Properties               string         `json:"properties,omitempty"`
	IsSender                 bool           `json:"-" gorm:"-"`
	IsSmartContract          bool           `json:"-" gorm:"-"`
	TotalBalanceWithStake    string         `json:"totalBalanceWithStake,omitempty"`
	TotalBalanceWithStakeNum float64        `json:"totalBalanceWithStakeNum,omitempty"`
	DataID                   int            `json:"-"`
	Data                     *TokenMetaData `json:"data,omitempty" gorm:"-"`
}

// TokenMetaData holds data about a token metadata
type TokenMetaData struct {
	Name               string   `json:"name,omitempty"`
	Creator            string   `json:"creator,omitempty"`
	Royalties          uint32   `json:"royalties,omitempty"`
	Hash               []byte   `json:"hash,omitempty" gorm:"serializer:json"`
	URIs               [][]byte `json:"uris,omitempty" gorm:"serializer:json"`
	Tags               []string `json:"tags,omitempty" gorm:"serializer:json"`
	Attributes         []byte   `json:"attributes,omitempty" gorm:"serializer:json"`
	MetaData           string   `json:"metadata,omitempty"`
	NonEmptyURIs       bool     `json:"nonEmptyURIs"`
	WhiteListedStorage bool     `json:"whiteListedStorage"`
}

// AccountBalanceHistory represents an entry in the user accounts balances history
type AccountBalanceHistory struct {
	Address         string        `json:"address"`
	Timestamp       time.Duration `json:"timestamp"`
	Balance         string        `json:"balance"`
	Token           string        `json:"token,omitempty"`
	Identifier      string        `json:"identifier,omitempty"`
	TokenNonce      uint64        `json:"tokenNonce,omitempty"`
	IsSender        bool          `json:"isSender,omitempty"`
	IsSmartContract bool          `json:"isSmartContract,omitempty"`
}

// Account is a structure that is needed for regular accounts
type Account struct {
	UserAccount coreData.UserAccountHandler
	IsSender    bool
}

// AccountESDT is a structure that is needed for ESDT accounts
type AccountESDT struct {
	Account         coreData.UserAccountHandler
	TokenIdentifier string
	NFTNonce        uint64
	IsSender        bool
	IsNFTOperation  bool
	Type            string
}
