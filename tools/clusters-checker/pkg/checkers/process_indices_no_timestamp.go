package checkers

import (
	"encoding/json"
)

const (
	defaultSize = 9000
	sizeRating  = 5000
)

func (cc *clusterChecker) CompareIndicesNoTimestamp() error {
	for _, index := range cc.indicesNoTimestamp {
		err := cc.compareIndex(index)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cc *clusterChecker) compareIndex(index string) error {
	count := 0
	handlerFunc := func(responseBytes []byte) error {
		count++
		genericResponse := &generalElasticResponse{}
		err := json.Unmarshal(responseBytes, genericResponse)
		if err != nil {
			return err
		}

		log.Info(cc.logPrefix+": comparing", "bulk size", len(genericResponse.Hits.Hits), "count", count)

		return cc.processResponse(index, genericResponse)
	}

	size := defaultSize
	if index == "rating" {
		size = sizeRating
	}

	return cc.clientSource.DoScrollRequestAllDocuments(index, getAll(true), handlerFunc, size)
}

func (cc *clusterChecker) processResponse(index string, genericResponse *generalElasticResponse) error {
	mapResponseSource, ids := convertResponseInMap(genericResponse)

	genericResponseDestination := &generalElasticResponse{}
	err := cc.clientDestination.DoGetRequest(index, queryMultipleObj(ids, true), genericResponseDestination, len(ids))
	if err != nil {
		return err
	}

	mapResponseDestination, _ := convertResponseInMap(genericResponseDestination)

	cc.compareResultsNo(index, mapResponseSource, mapResponseDestination)

	return nil
}

func (cc *clusterChecker) compareResultsNo(index string, sourceRes, destinationRes map[string]json.RawMessage) {
	for id, rawDataSource := range sourceRes {
		rawDataDestination, found := destinationRes[id]
		if !found {
			log.Warn(cc.logPrefix+": cannot find document", "index", index, "id", id)
			continue
		}

		equal, err := areEqualJSON(rawDataSource, rawDataDestination)
		if err != nil {
			log.Error(cc.logPrefix+": cannot compare json", "error", err.Error(), "index", index, "id", id)
			continue
		}

		if !equal {
			log.Warn(cc.logPrefix+": different documents", "index", index, "id", id)
		}
	}
}

func convertResponseInMap(response *generalElasticResponse) (map[string]json.RawMessage, []string) {
	mapResponse := make(map[string]json.RawMessage, len(response.Hits.Hits))
	ids := make([]string, 0, len(response.Hits.Hits))

	for _, hit := range response.Hits.Hits {
		ids = append(ids, hit.ID)
		mapResponse[hit.ID] = hit.Source
	}

	return mapResponse, ids
}
