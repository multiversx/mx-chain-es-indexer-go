package checkers

import (
	"encoding/json"

	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var (
	log = logger.GetOrCreate("checkers")
)

type clusterChecker struct {
	missingFromSource      map[string]json.RawMessage
	missingFromDestination map[string]json.RawMessage

	clientSource         ESClient
	clientDestination    ESClient
	indicesNoTimestamp   []string
	indicesWithTimestamp []string

	startTimestamp, stopTimestamp int

	logPrefix string
}

func (cc *clusterChecker) CompareIndicesWithTimestamp() error {
	for _, index := range cc.indicesWithTimestamp {
		err := cc.compareIndexWithTimestamp(index)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cc *clusterChecker) compareIndexWithTimestamp(index string) error {
	rspSource := &generalElasticResponse{}
	nextScrollIDSource, _, err := cc.clientSource.InitializeScroll(
		index,
		getAllSortTimestampASC(true, cc.startTimestamp, cc.stopTimestamp),
		rspSource,
	)
	if err != nil {
		return err
	}

	rspDestination := &generalElasticResponse{}
	nextScrollIDDestination, _, err := cc.clientDestination.InitializeScroll(
		index,
		getAllSortTimestampASC(true, cc.startTimestamp, cc.stopTimestamp),
		rspDestination,
	)
	if err != nil {
		return err
	}

	cc.compareResults(index, rspSource, rspDestination)

	cc.continueReading(index, nextScrollIDSource, nextScrollIDDestination)

	return nil
}

func (cc *clusterChecker) continueReading(index string, scrollIDSource, scrollIDDestination string) {
	sourceID := scrollIDSource
	destinationID := scrollIDDestination
	var errSource, errDestination error
	var doneSource, doneDestination bool

	count := 0
	for {
		count++

		chanResponseSource := make(chan *generalElasticResponse)
		chanResponseDestination := make(chan *generalElasticResponse)

		go func() {
			responseS := &generalElasticResponse{}
			var nextScroll string
			if !doneSource {
				nextScroll, doneSource, errSource = cc.clientSource.DoScrollRequestV2(sourceID, responseS)
				if errSource != nil {
					log.Error(cc.logPrefix+": cannot read from source", "index", index, "error", errSource.Error())
				}
			}
			sourceID = nextScroll
			chanResponseSource <- responseS
		}()

		go func() {
			var nextScroll string
			responseD := &generalElasticResponse{}
			if !doneDestination {
				nextScroll, doneDestination, errDestination = cc.clientDestination.DoScrollRequestV2(destinationID, responseD)
				if errDestination != nil {
					log.Error(cc.logPrefix+": cannot read from destination", "index", index, "error", errDestination.Error())
				}
			}
			destinationID = nextScroll
			chanResponseDestination <- responseD
		}()

		rspFromSource := <-chanResponseSource
		rspFromDestination := <-chanResponseDestination

		cc.compareResults(index, rspFromSource, rspFromDestination)
		log.Info(cc.logPrefix+": comparing results", "index", index, "count", count)
		if count%10 == 0 {
			cc.checkMaps(index, false)
		}

		if doneSource && doneDestination {
			cc.checkMaps(index, true)
			return
		}
	}
}

func (cc *clusterChecker) compareResults(index string, respSource, respDestination *generalElasticResponse) {
	mapSource, _ := convertResponseInMap(respSource)
	mapDestination, _ := convertResponseInMap(respDestination)

	for id, rawDataSource := range mapSource {
		rawDataDestination, found := mapDestination[id]
		if !found {
			cc.missingFromSource[id] = rawDataSource
			continue
		}

		delete(mapDestination, id)

		equal, err := areEqualJSON(rawDataSource, rawDataDestination)
		if err != nil {
			log.Error(cc.logPrefix+": cannot compare json", "error", err.Error(), "index", index, "id", id)
			continue
		}

		if !equal {
			log.Warn(cc.logPrefix+": different documents", "index", index, "id", id)
			continue
		}
	}

	for id, rawDataSource := range mapDestination {
		cc.missingFromDestination[id] = rawDataSource
	}
}

func (cc *clusterChecker) checkMaps(index string, finish bool) {
	log.Info(cc.logPrefix+": missing from source",
		"num", len(cc.missingFromDestination),
		"missing from destination num", len(cc.missingFromDestination),
	)
	for id, rawDataSource := range cc.missingFromSource {
		rawDataDestination, found := cc.missingFromDestination[id]
		if !found {
			if finish {
				log.Warn(cc.logPrefix+": cannot find document source", "index", index, "id", id)
			}
			continue
		}

		delete(cc.missingFromSource, id)
		delete(cc.missingFromDestination, id)

		equal, err := areEqualJSON(rawDataSource, rawDataDestination)
		if err != nil {
			log.Error(cc.logPrefix+": cannot compare json", "error", err.Error(), "index", index, "id", id)
			continue
		}

		if !equal {
			log.Warn(cc.logPrefix+": different documents", "index", index, "id", id)
			continue
		}
	}
	if finish {
		for id := range cc.missingFromDestination {
			log.Warn(cc.logPrefix+": cannot find document destination", "index", index, "id", id)
		}
	}

	if finish {
		log.Info(cc.logPrefix + ": DONE")
		cc.missingFromDestination = make(map[string]json.RawMessage, 0)
		cc.missingFromSource = make(map[string]json.RawMessage, 0)
	}
}
