package block

import (
	"encoding/json"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
)

// SerializeBlock will serialize a block for database
func (bp *blockProcessor) SerializeBlock(elasticBlock *data.Block, buffSlice *data.BufferSlice, index string) error {
	if elasticBlock == nil {
		return dataindexer.ErrNilElasticBlock
	}

	meta := []byte(fmt.Sprintf(`{ "index" : { "_index":"%s", "_id" : "%s" } }%s`, index, converters.JsonEscape(elasticBlock.Hash), "\n"))
	serializedData, errMarshal := json.Marshal(elasticBlock)
	if errMarshal != nil {
		return errMarshal
	}

	return buffSlice.PutData(meta, serializedData)
}

// SerializeExecutionResults will serialize execution results slice for database
func (bp *blockProcessor) SerializeExecutionResults(executionResults []*data.ExecutionResult, buffSlice *data.BufferSlice, index string) error {
	for _, result := range executionResults {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_index":"%s", "_id" : "%s" } }%s`, index, converters.JsonEscape(result.Hash), "\n"))
		serializedData, err := json.Marshal(result)
		if err != nil {
			return err
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			return err
		}
	}
	return nil
}

// SerializeEpochInfoData will serialize information about current epoch
func (bp *blockProcessor) SerializeEpochInfoData(header coreData.HeaderHandler, buffSlice *data.BufferSlice, index string) error {
	if check.IfNil(header) {
		return dataindexer.ErrNilHeaderHandler
	}

	epochInfo, err := getEpochInfoDataFromHeader(header)
	if err != nil {
		return err
	}

	id := header.GetEpoch()
	meta := []byte(fmt.Sprintf(`{ "index" : { "_index":"%s", "_id" : "%d" } }%s`, index, id, "\n"))
	serializedData, errMarshal := json.Marshal(epochInfo)
	if errMarshal != nil {
		return errMarshal
	}

	return buffSlice.PutData(meta, serializedData)
}

func getEpochInfoDataFromHeader(header coreData.HeaderHandler) (*data.EpochInfo, error) {
	epochInfo := &data.EpochInfo{
		AccumulatedFees: "0",
		DeveloperFees:   "0",
	}

	switch meta := header.(type) {
	case *block.MetaBlock:
		epochInfo.AccumulatedFees = meta.AccumulatedFeesInEpoch.String()
		epochInfo.DeveloperFees = meta.DevFeesInEpoch.String()
	case *block.MetaBlockV3:
		if check.IfNil(meta.LastExecutionResult) {
			break
		}
		epochInfo.AccumulatedFees = meta.LastExecutionResult.ExecutionResult.AccumulatedFeesInEpoch.String()
		epochInfo.DeveloperFees = meta.LastExecutionResult.ExecutionResult.DevFeesInEpoch.String()
	default:
		return nil, fmt.Errorf("%w in blockProcessor.SerializeEpochInfoData", dataindexer.ErrHeaderTypeAssertion)
	}

	return epochInfo, nil
}
