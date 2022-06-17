package block

import (
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
)

// SerializeBlock will serialize a block for database
func (bp *blockProcessor) SerializeBlock(elasticBlock *data.Block, buffSlice *data.BufferSlice, index string) error {
	if elasticBlock == nil {
		return indexer.ErrNilElasticBlock
	}

	meta := []byte(fmt.Sprintf(`{ "index" : { "_index":"%s", "_id" : "%s" } }%s`, index, elasticBlock.Hash, "\n"))
	serializedData, errMarshal := json.Marshal(elasticBlock)
	if errMarshal != nil {
		return errMarshal
	}

	return buffSlice.PutData(meta, serializedData)
}

// SerializeEpochInfoData will serialize information about current epoch
func (bp *blockProcessor) SerializeEpochInfoData(header coreData.HeaderHandler, buffSlice *data.BufferSlice, index string) error {
	if check.IfNil(header) {
		return indexer.ErrNilHeaderHandler
	}

	metablock, ok := header.(*block.MetaBlock)
	if !ok {
		return fmt.Errorf("%w in blockProcessor.SerializeEpochInfoData", indexer.ErrHeaderTypeAssertion)
	}

	epochInfo := &data.EpochInfo{
		AccumulatedFees: metablock.AccumulatedFeesInEpoch.String(),
		DeveloperFees:   metablock.DevFeesInEpoch.String(),
	}

	id := header.GetEpoch()
	meta := []byte(fmt.Sprintf(`{ "index" : { "_index":"%s", "_id" : "%d" } }%s`, index, id, "\n"))
	serializedData, errMarshal := json.Marshal(epochInfo)
	if errMarshal != nil {
		return errMarshal
	}

	return buffSlice.PutData(meta, serializedData)
}
