package tokens

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/esdt"

	"github.com/multiversx/mx-chain-es-indexer-go/client"
	"github.com/multiversx/mx-chain-es-indexer-go/client/logging"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	indexerdata "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/factory"
)

type sovereignIndexTokensHandler struct {
	indexingEnabled        bool
	mainChainElasticClient elasticproc.DatabaseClientHandler
	esdtPrefix             string
}

// NewSovereignIndexTokensHandler creates a new sovereign index tokens handler
func NewSovereignIndexTokensHandler(mainChainElastic factory.ElasticConfig, esdtPrefix string) (*sovereignIndexTokensHandler, error) {
	var mainChainElasticClient elasticproc.DatabaseClientHandler
	if mainChainElastic.Enabled {
		var err error
		argsEsClient := elasticsearch.Config{
			Addresses:     []string{mainChainElastic.Url},
			Username:      mainChainElastic.UserName,
			Password:      mainChainElastic.Password,
			Logger:        &logging.CustomLogger{},
			RetryOnStatus: []int{http.StatusConflict},
			RetryBackoff:  retryBackOff,
		}
		mainChainElasticClient, err = client.NewElasticClient(argsEsClient)
		if err != nil {
			return nil, err
		}
	}

	return &sovereignIndexTokensHandler{
		indexingEnabled:        mainChainElastic.Enabled,
		mainChainElasticClient: mainChainElasticClient,
		esdtPrefix:             esdtPrefix,
	}, nil
}

func retryBackOff(attempt int) time.Duration {
	return time.Duration(math.Exp2(float64(attempt))) * time.Second
}

// IndexCrossChainTokens will index the new tokens properties
func (sit *sovereignIndexTokensHandler) IndexCrossChainTokens(elasticClient elasticproc.DatabaseClientHandler, scrs []*data.ScResult, buffSlice *data.BufferSlice) error {
	if !sit.indexingEnabled {
		return nil
	}

	notFoundTokens, err := sit.getTokensFromScrs(elasticClient, scrs)
	if err != nil {
		return err
	}

	if len(notFoundTokens) == 0 { // no new tokens
		return nil
	}

	// get tokens from main chain elastic db
	mainChainTokens := &data.ResponseTokenInfo{}
	err = sit.mainChainElasticClient.DoMultiGet(context.Background(), notFoundTokens, indexerdata.TokensIndex, true, mainChainTokens)
	if err != nil {
		return err
	}

	return sit.indexNewTokens(mainChainTokens.Docs, buffSlice)
}

func (sit *sovereignIndexTokensHandler) getTokensFromScrs(elasticClient elasticproc.DatabaseClientHandler, scrs []*data.ScResult) ([]string, error) {
	receivedTokensIDs := make([]string, 0)
	for _, scr := range scrs {
		if scr.SenderShard == core.MainChainShardId {
			for _, token := range scr.Tokens {
				tokenPrefix, hasPrefix := esdt.IsValidPrefixedToken(token)
				if !hasPrefix || tokenPrefix != sit.esdtPrefix {
					receivedTokensIDs = append(receivedTokensIDs, token)
				}
				if isEsdt, tokenCollection := getTokenCollection(hasPrefix, token); isEsdt {
					receivedTokensIDs = append(receivedTokensIDs, tokenCollection)
				}
			}
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

func getTokenCollection(hasPrefix bool, tokenIdentifier string) (bool, string) {
	tokenSplit := strings.Split(tokenIdentifier, "-")
	if !hasPrefix && len(tokenSplit) == 3 {
		return true, tokenSplit[0] + "-" + tokenSplit[1]
	}
	if hasPrefix && len(tokenSplit) == 4 {
		return true, tokenSplit[1] + "-" + tokenSplit[2]
	}
	return false, ""
}

func (sit *sovereignIndexTokensHandler) indexNewTokens(responseTokensInfo []data.ResponseTokenInfoDB, buffSlice *data.BufferSlice) error {
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
