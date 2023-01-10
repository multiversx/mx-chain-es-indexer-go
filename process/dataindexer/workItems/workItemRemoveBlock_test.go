package workItems_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elastic-indexer-go/process/dataindexer/workItems"
	"github.com/multiversx/mx-chain-core-go/data"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"
)

func TestItemRemoveBlock_Save(t *testing.T) {
	countCalled := 0
	itemRemove := workItems.NewItemRemoveBlock(
		&mock.ElasticProcessorStub{
			RemoveHeaderCalled: func(header data.HeaderHandler) error {
				countCalled++
				return nil
			},
			RemoveMiniblocksCalled: func(header data.HeaderHandler, body *dataBlock.Body) error {
				countCalled++
				return nil
			},
			RemoveTransactionsCalled: func(header data.HeaderHandler, body *dataBlock.Body) error {
				countCalled++
				return nil
			},
		},
		&dataBlock.Body{},
		&dataBlock.Header{},
	)
	require.False(t, itemRemove.IsInterfaceNil())

	err := itemRemove.Save()
	require.NoError(t, err)
	require.Equal(t, 3, countCalled)
}

func TestItemRemoveBlock_SaveRemoveHeaderShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	itemRemove := workItems.NewItemRemoveBlock(
		&mock.ElasticProcessorStub{
			RemoveHeaderCalled: func(header data.HeaderHandler) error {
				return localErr
			},
		},
		&dataBlock.Body{},
		&dataBlock.Header{},
	)
	require.False(t, itemRemove.IsInterfaceNil())

	err := itemRemove.Save()
	require.Equal(t, localErr, err)
}

func TestItemRemoveBlock_SaveRemoveMiniblocksShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	itemRemove := workItems.NewItemRemoveBlock(
		&mock.ElasticProcessorStub{
			RemoveMiniblocksCalled: func(header data.HeaderHandler, body *dataBlock.Body) error {
				return localErr
			},
		},
		&dataBlock.Body{},
		&dataBlock.Header{},
	)
	require.False(t, itemRemove.IsInterfaceNil())

	err := itemRemove.Save()
	require.Equal(t, localErr, err)
}
