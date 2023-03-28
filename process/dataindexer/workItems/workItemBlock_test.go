package workItems_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer/workItems"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestItemBlock_SaveNilHeaderShouldRetNil(t *testing.T) {
	itemBlock := workItems.NewItemBlock(
		&mock.ElasticProcessorStub{},
		&outport.OutportBlockWithHeader{},
	)
	require.False(t, itemBlock.IsInterfaceNil())

	err := itemBlock.Save()
	assert.Nil(t, err)
}

func TestItemBlock_SaveHeaderShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	itemBlock := workItems.NewItemBlock(
		&mock.ElasticProcessorStub{
			SaveHeaderCalled: func(_ *outport.OutportBlockWithHeader) error {
				return localErr
			},
		},
		&outport.OutportBlockWithHeader{
			OutportBlock: &outport.OutportBlock{
				BlockData: &outport.BlockData{},
			},
			Header: &dataBlock.Header{},
		},
	)
	require.False(t, itemBlock.IsInterfaceNil())

	err := itemBlock.Save()
	require.True(t, errors.Is(err, localErr))
}

func TestItemBlock_SaveNoMiniblocksShoulCallSaveHeader(t *testing.T) {
	countCalled := 0
	itemBlock := workItems.NewItemBlock(
		&mock.ElasticProcessorStub{
			SaveHeaderCalled: func(_ *outport.OutportBlockWithHeader) error {
				countCalled++
				return nil
			},
			SaveMiniblocksCalled: func(header data.HeaderHandler, body *dataBlock.Body) error {
				countCalled++
				return nil
			},
			SaveTransactionsCalled: func(_ *outport.OutportBlockWithHeader) error {
				countCalled++
				return nil
			},
		},
		&outport.OutportBlockWithHeader{
			OutportBlock: &outport.OutportBlock{
				BlockData: &outport.BlockData{
					Body: &dataBlock.Body{MiniBlocks: []*dataBlock.MiniBlock{}},
				},
			},
			Header: &dataBlock.Header{},
		},
	)
	require.False(t, itemBlock.IsInterfaceNil())

	err := itemBlock.Save()
	require.NoError(t, err)
	require.Equal(t, 1, countCalled)
}

func TestItemBlock_SaveMiniblocksShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	itemBlock := workItems.NewItemBlock(
		&mock.ElasticProcessorStub{
			SaveMiniblocksCalled: func(header data.HeaderHandler, body *dataBlock.Body) error {
				return localErr
			},
		},
		&outport.OutportBlockWithHeader{
			OutportBlock: &outport.OutportBlock{
				BlockData: &outport.BlockData{
					Body: &dataBlock.Body{MiniBlocks: []*dataBlock.MiniBlock{{}}},
				},
			},
			Header: &dataBlock.Header{},
		},
	)
	require.False(t, itemBlock.IsInterfaceNil())

	err := itemBlock.Save()
	require.True(t, errors.Is(err, localErr))
}

func TestItemBlock_SaveTransactionsShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	itemBlock := workItems.NewItemBlock(
		&mock.ElasticProcessorStub{
			SaveTransactionsCalled: func(_ *outport.OutportBlockWithHeader) error {
				return localErr
			},
		},
		&outport.OutportBlockWithHeader{
			OutportBlock: &outport.OutportBlock{
				BlockData: &outport.BlockData{
					Body: &dataBlock.Body{MiniBlocks: []*dataBlock.MiniBlock{{}}},
				},
			},
			Header: &dataBlock.Header{},
		},
	)
	require.False(t, itemBlock.IsInterfaceNil())

	err := itemBlock.Save()
	require.True(t, errors.Is(err, localErr))
}

func TestItemBlock_SaveShouldWork(t *testing.T) {
	countCalled := 0
	itemBlock := workItems.NewItemBlock(
		&mock.ElasticProcessorStub{
			SaveHeaderCalled: func(_ *outport.OutportBlockWithHeader) error {
				countCalled++
				return nil
			},
			SaveMiniblocksCalled: func(header data.HeaderHandler, body *dataBlock.Body) error {
				countCalled++
				return nil
			},
			SaveTransactionsCalled: func(_ *outport.OutportBlockWithHeader) error {
				countCalled++
				return nil
			},
		},
		&outport.OutportBlockWithHeader{
			OutportBlock: &outport.OutportBlock{
				BlockData: &outport.BlockData{
					Body: &dataBlock.Body{MiniBlocks: []*dataBlock.MiniBlock{{}}},
				},
			},
			Header: &dataBlock.Header{},
		},
	)
	require.False(t, itemBlock.IsInterfaceNil())

	err := itemBlock.Save()
	require.NoError(t, err)
	require.Equal(t, 3, countCalled)
}
