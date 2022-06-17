package process

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
)

func (ei *elasticProcessor) indexTokens(tokensData []*data.TokenInfo, updateNFTData []*data.NFTDataUpdate, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(data.TokensIndex) {
		return nil
	}

	err := ei.logsAndEventsProc.SerializeTokens(tokensData, updateNFTData, buffSlice, data.TokensIndex)
	if err != nil {
		return err
	}

	err = ei.addTokenType(tokensData, data.AccountsESDTIndex)
	if err != nil {
		return err
	}

	return ei.addTokenType(tokensData, data.TokensIndex)
}

func (ei *elasticProcessor) addTokenType(tokensData []*data.TokenInfo, index string) error {
	if len(tokensData) == 0 {
		return nil
	}

	defer func(startTime time.Time) {
		log.Debug("elasticProcessor.addTokenType", "index", index, "duration", time.Since(startTime))
	}(time.Now())

	for _, td := range tokensData {
		if td.Type == core.FungibleESDT {
			continue
		}

		handlerFunc := func(responseBytes []byte) error {
			responseScroll := &data.ResponseScroll{}
			err := json.Unmarshal(responseBytes, responseScroll)
			if err != nil {
				return err
			}

			ids := make([]string, 0, len(responseScroll.Hits.Hits))
			for _, res := range responseScroll.Hits.Hits {
				ids = append(ids, res.ID)
			}

			buffSlice := data.NewBufferSlice(ei.bulkRequestMaxSize)
			err = ei.accountsProc.SerializeTypeForProvidedIDs(ids, td.Type, buffSlice, index)
			if err != nil {
				return err
			}

			return ei.doBulkRequests(index, buffSlice.Buffers())
		}

		query := fmt.Sprintf(`{"query": {"bool": {"must": [{"match": {"token": {"query": "%s","operator": "AND"}}}],"must_not":[{"exists": {"field": "type"}}]}}}`, td.Token)
		resultsCount, err := ei.elasticClient.DoCountRequest(index, []byte(query))
		if err != nil || resultsCount == 0 {
			return err
		}

		err = ei.elasticClient.DoScrollRequest(index, []byte(query), false, handlerFunc)
		if err != nil {
			return err
		}
	}

	return nil
}
