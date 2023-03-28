package workItems

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

type itemRemoveBlock struct {
	indexer removeIndexer
	header  data.HeaderHandler
	body    *block.Body
}

// NewItemRemoveBlock will create a new instance of itemRemoveBlock
func NewItemRemoveBlock(
	indexer removeIndexer,
	header data.HeaderHandler,
	body *block.Body,
) WorkItemHandler {
	return &itemRemoveBlock{
		indexer: indexer,
		header:  header,
		body:    body,
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (wirb *itemRemoveBlock) IsInterfaceNil() bool {
	return wirb == nil
}

// Save will remove a block and miniblocks from elasticsearch database
func (wirb *itemRemoveBlock) Save() error {
	err := wirb.indexer.RemoveHeader(wirb.header)
	if err != nil {
		return err
	}

	err = wirb.indexer.RemoveMiniblocks(wirb.header, wirb.body)
	if err != nil {
		return err
	}

	err = wirb.indexer.RemoveTransactions(wirb.header, wirb.body)
	if err != nil {
		return err
	}

	return wirb.indexer.RemoveAccountsESDT(wirb.header.GetTimeStamp(), wirb.header.GetShardID())
}
