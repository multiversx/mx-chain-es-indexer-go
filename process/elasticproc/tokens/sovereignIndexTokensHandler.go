package tokens

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
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
	mainChainElasticClient elasticproc.DatabaseClientHandler
	esdtPrefix             string
}

func NewSovereignIndexTokensHandler(mainChainElastic factory.MainChainElastic, esdtPrefix string) (*sovereignIndexTokensHandler, error) {
	argsEsClient := elasticsearch.Config{
		Addresses:     []string{mainChainElastic.Url},
		Username:      mainChainElastic.UserName,
		Password:      mainChainElastic.Password,
		Logger:        &logging.CustomLogger{},
		RetryOnStatus: []int{http.StatusConflict},
		RetryBackoff:  retryBackOff,
	}
	mainChainElasticClient, err := client.NewElasticClient(argsEsClient)
	if err != nil {
		return nil, err
	}

	return &sovereignIndexTokensHandler{
		mainChainElasticClient: mainChainElasticClient,
		esdtPrefix:             esdtPrefix,
	}, nil
}

func retryBackOff(attempt int) time.Duration {
	return time.Duration(math.Exp2(float64(attempt))) * time.Second
}

func (sit *sovereignIndexTokensHandler) IndexCrossChainTokens(elasticClient elasticproc.DatabaseClientHandler, scrs []*data.ScResult, buffSlice *data.BufferSlice) error {
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
		if token.Found == false {
			newTokens = append(newTokens, token.ID)
		}
	}

	return newTokens, nil
}

func (sit *sovereignIndexTokensHandler) indexNewTokens(responseTokensInfo []data.ResponseTokenInfoDB, buffSlice *data.BufferSlice) error {
	for _, responseToken := range responseTokensInfo {
		token := formatToken(responseToken)

		meta := []byte(fmt.Sprintf(`{ "index" : { "_index":"%s", "_id" : "%s" } }%s`, indexerdata.TokensIndex, converters.JsonEscape(token.Token), "\n"))
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

func formatToken(token data.ResponseTokenInfoDB) data.TokenInfo {
	token.Source.OwnersHistory = nil
	token.Source.Properties = nil

	return token.Source
}

// IsInterfaceNil returns true if there is no value under the interface
func (sit *sovereignIndexTokensHandler) IsInterfaceNil() bool {
	return sit == nil
}
