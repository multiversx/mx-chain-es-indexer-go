package data

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
)

const (
	metaESDT        = "MetaESDT"
	dynamicMetaESDT = "DynamicMetaESDT"
)

// NFTDataUpdate will contain the update information for an NFT or SFT
type NFTDataUpdate struct {
	Identifier    string
	Address       string
	NewCreator    string
	URIsToAdd     [][]byte
	NewAttributes []byte
	Freeze        bool
	UnFreeze      bool
	Pause         bool
	UnPause       bool
	SetURIs       bool
	NewRoyalties  core.OptionalUint32
	NewMetaData   *TokenMetaData
}

// ResponseTokens is the structure for the tokens response
type ResponseTokens struct {
	Docs []ResponseTokenDB `json:"docs"`
}

// ResponseTokenDB is the structure for the token response
type ResponseTokenDB struct {
	Found  bool        `json:"found"`
	ID     string      `json:"_id"`
	Source SourceToken `json:"_source"`
}

// SourceToken is the structure for the source body of a token
type SourceToken struct {
	Type         string `json:"type"`
	CurrentOwner string `json:"currentOwner"`
}

// ResponseTokenInfo is the structure for the tokens info response
type ResponseTokenInfo struct {
	Docs []ResponseTokenInfoDB `json:"docs"`
}

// ResponseTokenInfoDB is the structure for the token info response
type ResponseTokenInfoDB struct {
	Found  bool      `json:"found"`
	ID     string    `json:"_id"`
	Source TokenInfo `json:"_source"`
}

// TokenInfo is a structure that is needed to store information about a token
type TokenInfo struct {
	Name              string           `json:"name,omitempty"`
	Ticker            string           `json:"ticker,omitempty"`
	Identifier        string           `json:"identifier,omitempty"`
	Token             string           `json:"token,omitempty"`
	Issuer            string           `json:"issuer,omitempty"`
	CurrentOwner      string           `json:"currentOwner,omitempty"`
	NumDecimals       uint64           `json:"numDecimals"`
	Type              string           `json:"type,omitempty"`
	Nonce             uint64           `json:"nonce,omitempty"`
	Timestamp         time.Duration    `json:"timestamp,omitempty"`
	Data              *TokenMetaData   `json:"data,omitempty"`
	OwnersHistory     []*OwnerData     `json:"ownersHistory,omitempty"`
	TransferOwnership bool             `json:"-"`
	ChangeToDynamic   bool             `json:"-"`
	Properties        *TokenProperties `json:"properties,omitempty"`
}

// TokenProperties is a structure that is needed to store all properties of a token
type TokenProperties struct {
	Mintable                 bool `json:"canMint"`
	Burnable                 bool `json:"canBurn"`
	Upgradable               bool `json:"canUpgrade"`
	CanTransferNFTCreateRole bool `json:"canTransferNFTCreateRole"`
	CanAddSpecialRoles       bool `json:"canAddSpecialRoles"`
	CanPause                 bool `json:"canPause"`
	CanFreeze                bool `json:"canFreeze"`
	CanWipe                  bool `json:"canWipe"`
	CanChangeOwner           bool `json:"canChangeOwner"`
	CanCreateMultiShard      bool `json:"canCreateMultiShard"`
}

// OwnerData is a structure that is needed to store information about an owner
type OwnerData struct {
	TxHash    string        `json:"txHash,omitempty"`
	Address   string        `json:"address"`
	Timestamp time.Duration `json:"timestamp"`
}

// TokensHandler defines the actions that a tokens' handler should do
type TokensHandler interface {
	Add(tokenInfo *TokenInfo)
	Len() int
	AddTypeAndOwnerFromResponse(res *ResponseTokens)
	PutTypeAndOwnerInAccountsESDT(accountsESDT map[string]*AccountInfo)
	GetAllTokens() []string
	GetAll() []*TokenInfo
	GetAllWithoutMetaESDT() []*TokenInfo
	IsInterfaceNil() bool
}

type tokensInfo struct {
	tokensInfo map[string]*TokenInfo
}

// NewTokensInfo will create a new instance of tokensInfo
func NewTokensInfo() *tokensInfo {
	return &tokensInfo{
		tokensInfo: make(map[string]*TokenInfo),
	}
}

// Add will add tokenInfo
func (ti *tokensInfo) Add(tokenInfo *TokenInfo) {
	mapKey := tokenInfo.Token
	if tokenInfo.Identifier != "" {
		mapKey = tokenInfo.Identifier
	}

	ti.tokensInfo[mapKey] = tokenInfo
}

// GetAll will return all tokens information
func (ti *tokensInfo) GetAll() []*TokenInfo {
	tokens := make([]*TokenInfo, 0, len(ti.tokensInfo))
	for _, tokenData := range ti.tokensInfo {
		tokens = append(tokens, tokenData)
	}

	return tokens
}

// GetAllWithoutMetaESDT will return all tokens except metaESDT tokens
func (ti *tokensInfo) GetAllWithoutMetaESDT() []*TokenInfo {
	tokens := make([]*TokenInfo, 0)
	for _, tokenData := range ti.tokensInfo {
		if tokenData.Type == metaESDT || tokenData.Type == dynamicMetaESDT {
			continue
		}

		tokens = append(tokens, tokenData)
	}

	return tokens
}

// GetAllTokens wil return all tokens names
func (ti *tokensInfo) GetAllTokens() []string {
	tokensMap := make(map[string]struct{})
	for _, tokenData := range ti.tokensInfo {
		tokensMap[tokenData.Token] = struct{}{}
	}

	tokensSlice := make([]string, 0, len(tokensMap))
	for token := range tokensMap {
		tokensSlice = append(tokensSlice, token)
	}

	return tokensSlice
}

// AddTypeAndOwnerFromResponse will add token type and current owner from response
func (ti *tokensInfo) AddTypeAndOwnerFromResponse(res *ResponseTokens) {
	keyTokenValueIdentifiers := make(map[string][]string)
	for identifier, tokenData := range ti.tokensInfo {
		keyTokenValueIdentifiers[tokenData.Token] = append(keyTokenValueIdentifiers[tokenData.Token], identifier)
	}

	if res == nil {
		return
	}

	for _, tokenData := range res.Docs {
		if !tokenData.Found {
			continue
		}

		for _, identifier := range keyTokenValueIdentifiers[tokenData.ID] {
			_, ok := ti.tokensInfo[identifier]
			if !ok {
				continue
			}

			ti.tokensInfo[identifier].Type = tokenData.Source.Type
			ti.tokensInfo[identifier].CurrentOwner = tokenData.Source.CurrentOwner
		}
	}
}

// PutTypeAndOwnerInAccountsESDT will put in the provided accounts ESDT map token type and current owner
func (ti *tokensInfo) PutTypeAndOwnerInAccountsESDT(accountsESDT map[string]*AccountInfo) {
	for _, accountESDT := range accountsESDT {
		tokenData, ok := ti.tokensInfo[accountESDT.TokenIdentifier]
		if !ok {
			continue
		}

		accountESDT.Type = tokenData.Type
		accountESDT.CurrentOwner = tokenData.CurrentOwner
	}
}

// Len will return the number of tokens
func (ti *tokensInfo) Len() int {
	return len(ti.tokensInfo)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ti *tokensInfo) IsInterfaceNil() bool {
	return ti == nil
}
