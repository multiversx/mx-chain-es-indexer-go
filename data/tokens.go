package data

import "time"

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
	Type string `json:"type"`
}

// TokenInfo is a structure that is needed to store information about a token
type TokenInfo struct {
	Name       string         `json:"name,omitempty"`
	Ticker     string         `json:"ticker,omitempty"`
	Identifier string         `json:"identifier,omitempty"`
	Token      string         `json:"token,omitempty"`
	Issuer     string         `json:"issuer,omitempty"`
	Type       string         `json:"type,omitempty"`
	Nonce      uint64         `json:"nonce,omitempty"`
	Timestamp  time.Duration  `json:"timestamp,omitempty"`
	Data       *TokenMetaData `json:"data,omitempty"`
}

// TokensHandler defines the actions that an tokens handler should do
type TokensHandler interface {
	Add(tokenInfo *TokenInfo)
	Len() int
	AddTypeFromResponse(res *ResponseTokens)
	GetAllTokens() []string
	GetAll() []*TokenInfo
}

type tokensInfo struct {
	tokensInfo map[string]*TokenInfo
}

// NewTokensInfo will create a new instance of tokensInfo
func NewTokensInfo() *tokensInfo {
	return &tokensInfo{
		tokensInfo: make(map[string]*TokenInfo, 0),
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

// GetAllTokens wil return all tokens names
func (ti *tokensInfo) GetAllTokens() []string {
	tokens := make([]string, 0, len(ti.tokensInfo))
	for _, tokenData := range ti.tokensInfo {
		tokens = append(tokens, tokenData.Token)
	}

	return tokens
}

// AddTypeFromResponse will add token type from response
func (ti *tokensInfo) AddTypeFromResponse(res *ResponseTokens) {
	if res == nil {
		return
	}

	for _, tokenData := range res.Docs {
		if !tokenData.Found {
			continue
		}

		_, ok := ti.tokensInfo[tokenData.ID]
		if !ok {
			continue
		}

		ti.tokensInfo[tokenData.ID].Type = tokenData.Source.Type
	}
}

// Len will return the number of tokens
func (ti *tokensInfo) Len() int {
	return len(ti.tokensInfo)
}
