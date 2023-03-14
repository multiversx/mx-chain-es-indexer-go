package workItems

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/unmarshal"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/marshal"
)

type itemRemoveBlock struct {
	indexer    removeIndexer
	marshaller marshal.Marshalizer
	blockData  *outport.BlockData
}

// NewItemRemoveBlock will create a new instance of itemRemoveBlock
func NewItemRemoveBlock(
	indexer removeIndexer,
	marshaller marshal.Marshalizer,
	blockData *outport.BlockData,
) WorkItemHandler {
	return &itemRemoveBlock{
		indexer:    indexer,
		marshaller: marshaller,
		blockData:  blockData,
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (wirb *itemRemoveBlock) IsInterfaceNil() bool {
	return wirb == nil
}

// Save will remove a block and miniblocks from elasticsearch database
func (wirb *itemRemoveBlock) Save() error {
	header, err := unmarshal.GetHeaderFromBytes(wirb.marshaller, core.HeaderType(wirb.blockData.HeaderType), wirb.blockData.HeaderBytes)
	if err != nil {
		return err
	}

	err = wirb.indexer.RemoveHeader(header)
	if err != nil {
		return err
	}

	err = wirb.indexer.RemoveMiniblocks(header, wirb.blockData.Body)
	if err != nil {
		return err
	}

	err = wirb.indexer.RemoveTransactions(header, wirb.blockData.Body)
	if err != nil {
		return err
	}

	return wirb.indexer.RemoveAccountsESDT(header.GetTimeStamp(), header.GetShardID())
}
