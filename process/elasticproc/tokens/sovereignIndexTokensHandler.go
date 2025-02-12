package tokens

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/esdt"

	"github.com/multiversx/mx-chain-es-indexer-go/data"
	indexerdata "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
)

type sovereignIndexTokensHandler struct {
	mainChainElasticClient elasticproc.MainChainDatabaseClientHandler
	esdtPrefix             string
}

// NewSovereignIndexTokensHandler creates a new sovereign index tokens handler
func NewSovereignIndexTokensHandler(mainChainElasticClient elasticproc.MainChainDatabaseClientHandler, esdtPrefix string) (*sovereignIndexTokensHandler, error) {
	return &sovereignIndexTokensHandler{
		mainChainElasticClient: mainChainElasticClient,
		esdtPrefix:             esdtPrefix,
	}, nil
}

// IndexCrossChainTokens will index the new tokens properties
func (sit *sovereignIndexTokensHandler) IndexCrossChainTokens(elasticClient elasticproc.DatabaseClientHandler, scrs []*data.ScResult, buffSlice *data.BufferSlice) error {
	if !sit.mainChainElasticClient.IsEnabled() {
		return nil
	}

	newTokens, err := sit.getNewTokensFromSCRs(elasticClient, scrs)
	if err != nil {
		return err
	}

	if len(newTokens) == 0 { // no new tokens
		return nil
	}

	// get tokens from main chain elastic db
	mainChainTokens := &data.ResponseTokenInfo{}
	err = sit.mainChainElasticClient.DoMultiGet(context.Background(), newTokens, indexerdata.TokensIndex, true, mainChainTokens)
	if err != nil {
		return err
	}

	return sit.serializeNewTokens(mainChainTokens.Docs, buffSlice)
}

func (sit *sovereignIndexTokensHandler) getNewTokensFromSCRs(elasticClient elasticproc.DatabaseClientHandler, scrs []*data.ScResult) ([]string, error) {
	receivedTokensIDs := make([]string, 0)
	for _, scr := range scrs {
		if scr.SenderShard == core.MainChainShardId {
			receivedTokensIDs = append(receivedTokensIDs, sit.extractNewSovereignTokens(scr.Tokens)...)
		}
	}

	if len(receivedTokensIDs) == 0 {
		return make([]string, 0), nil
	}

	responseTokens := &data.ResponseTokens{}
	err := elasticClient.DoMultiGet(context.Background(), receivedTokensIDs, indexerdata.TokensIndex, true, responseTokens)
	if err != nil {
		return nil, err
	}

	newTokens := make([]string, 0)
	for _, token := range responseTokens.Docs {
		if !token.Found {
			newTokens = append(newTokens, token.ID)
		}
	}

	return newTokens, nil
}

func (sit *sovereignIndexTokensHandler) extractNewSovereignTokens(tokens []string) []string {
	receivedTokensIDs := make([]string, 0)
	for _, token := range tokens {
		tokenPrefix, hasPrefix := esdt.IsValidPrefixedToken(token)
		if !hasPrefix || tokenPrefix != sit.esdtPrefix {
			receivedTokensIDs = append(receivedTokensIDs, token)
		}
		if tokenCollection := getTokenCollection(hasPrefix, token); tokenCollection != "" {
			receivedTokensIDs = append(receivedTokensIDs, tokenCollection)
		}
	}

	return receivedTokensIDs
}

func getTokenCollection(hasPrefix bool, tokenIdentifier string) string {
	tokenSplit := strings.Split(tokenIdentifier, "-")
	if !hasPrefix && len(tokenSplit) == 3 {
		return tokenSplit[0] + "-" + tokenSplit[1]
	}
	if hasPrefix && len(tokenSplit) == 4 {
		return tokenSplit[1] + "-" + tokenSplit[2]
	}
	return ""
}

func (sit *sovereignIndexTokensHandler) serializeNewTokens(responseTokensInfo []data.ResponseTokenInfoDB, buffSlice *data.BufferSlice) error {
	for _, responseToken := range responseTokensInfo {
		token, identifier := formatToken(responseToken)

		meta := []byte(fmt.Sprintf(`{ "index" : { "_index":"%s", "_id" : "%s" } }%s`, indexerdata.TokensIndex, converters.JsonEscape(identifier), "\n"))
		serializedTokenData, err := json.Marshal(token)
		if err != nil {
			return err
		}

		err = buffSlice.PutData(meta, serializedTokenData)
		if err != nil {
			return err
		}
	}

	return nil
}

func formatToken(token data.ResponseTokenInfoDB) (data.TokenInfo, string) {
	token.Source.OwnersHistory = nil
	token.Source.Properties = nil

	identifier := token.Source.Identifier // for NFTs
	if identifier == "" {
		identifier = token.Source.Token // for tokens/collections
	}
	return token.Source, identifier
}

// IsInterfaceNil returns true if there is no value under the interface
func (sit *sovereignIndexTokensHandler) IsInterfaceNil() bool {
	return sit == nil
}
