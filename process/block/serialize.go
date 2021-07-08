package block

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go/core/check"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// SerializeBlock will serialize a block for database
func (bp *blockProcessor) SerializeBlock(elasticBlock *data.Block) (*bytes.Buffer, error) {
	if elasticBlock == nil {
		return nil, indexer.ErrNilElasticBlock
	}

	blockBytes, err := json.Marshal(elasticBlock)
	if err != nil {
		return nil, err
	}

	buff := &bytes.Buffer{}

	buff.Grow(len(blockBytes))
	_, err = buff.Write(blockBytes)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

// SerializeEpochInfoData will serialize information about current epoch
func (bp *blockProcessor) SerializeEpochInfoData(header nodeData.HeaderHandler) (*bytes.Buffer, error) {
	if check.IfNil(header) {
		return nil, indexer.ErrNilHeaderHandler
	}

	metablock, ok := header.(*block.MetaBlock)
	if !ok {
		return nil, fmt.Errorf("%w in blockProcessor.SerializeEpochInfoData", indexer.ErrHeaderTypeAssertion)
	}

	epochInfo := &data.EpochInfo{
		AccumulatedFees: metablock.AccumulatedFeesInEpoch.String(),
		DeveloperFees:   metablock.DevFeesInEpoch.String(),
	}

	epochInfoBytes, err := json.Marshal(epochInfo)
	if err != nil {
		return nil, err
	}

	buff := &bytes.Buffer{}
	buff.Grow(len(epochInfoBytes))
	_, err = buff.Write(epochInfoBytes)
	if err != nil {
		return nil, err
	}

	return buff, nil
}
