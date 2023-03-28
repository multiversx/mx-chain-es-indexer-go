package dataindexer

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer/workItems"
	"github.com/stretchr/testify/require"
)

func NewDataIndexerArguments() ArgDataIndexer {
	return ArgDataIndexer{
		DataDispatcher:   &mock.DispatcherMock{},
		ElasticProcessor: &mock.ElasticProcessorStub{},
		HeaderMarshaller: &mock.MarshalizerMock{},
	}
}

func TestDataIndexer_NewIndexerWithNilDataDispatcherShouldErr(t *testing.T) {
	arguments := NewDataIndexerArguments()
	arguments.DataDispatcher = nil
	ei, err := NewDataIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, ErrNilDataDispatcher, err)
}

func TestDataIndexer_NewIndexerWithNilElasticProcessorShouldErr(t *testing.T) {
	arguments := NewDataIndexerArguments()
	arguments.ElasticProcessor = nil
	ei, err := NewDataIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, ErrNilElasticProcessor, err)
}

func TestDataIndexer_NewIndexerWithNilMarshalizerShouldErr(t *testing.T) {
	arguments := NewDataIndexerArguments()
	arguments.HeaderMarshaller = nil
	ei, err := NewDataIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, core.ErrNilMarshalizer, err)
}

func TestDataIndexer_NewIndexerWithCorrectParamsShouldWork(t *testing.T) {
	arguments := NewDataIndexerArguments()

	ei, err := NewDataIndexer(arguments)

	require.Nil(t, err)
	require.False(t, check.IfNil(ei))
}

func TestDataIndexer_SaveBlock(t *testing.T) {
	called := false

	arguments := NewDataIndexerArguments()
	arguments.DataDispatcher = &mock.DispatcherMock{
		AddCalled: func(item workItems.WorkItemHandler) {
			called = true
		},
	}
	ei, _ := NewDataIndexer(arguments)

	args := &outport.OutportBlock{
		BlockData: &outport.BlockData{
			HeaderType:  string(core.ShardHeaderV2),
			Body:        &dataBlock.Body{MiniBlocks: []*dataBlock.MiniBlock{{}}},
			HeaderBytes: []byte("{}"),
		},
	}
	err := ei.SaveBlock(args)
	require.Nil(t, err)
	require.True(t, called)
}

func TestDataIndexer_SaveRoundInfo(t *testing.T) {
	called := false

	arguments := NewDataIndexerArguments()
	arguments.DataDispatcher = &mock.DispatcherMock{
		AddCalled: func(item workItems.WorkItemHandler) {
			called = true
		},
	}

	arguments.HeaderMarshaller = &mock.MarshalizerMock{Fail: true}
	ei, _ := NewDataIndexer(arguments)
	_ = ei.Close()

	err := ei.SaveRoundsInfo(&outport.RoundsInfo{})
	require.True(t, called)
	require.Nil(t, err)
}

func TestDataIndexer_SaveValidatorsPubKeys(t *testing.T) {
	called := false

	arguments := NewDataIndexerArguments()
	arguments.DataDispatcher = &mock.DispatcherMock{
		AddCalled: func(item workItems.WorkItemHandler) {
			called = true
		},
	}
	ei, _ := NewDataIndexer(arguments)

	valPubKey := make(map[uint32][][]byte)

	keys := [][]byte{[]byte("key")}
	valPubKey[0] = keys

	err := ei.SaveValidatorsPubKeys(&outport.ValidatorsPubKeys{})
	require.True(t, called)
	require.Nil(t, err)
}

func TestDataIndexer_SaveValidatorsRating(t *testing.T) {
	called := false

	arguments := NewDataIndexerArguments()
	arguments.DataDispatcher = &mock.DispatcherMock{
		AddCalled: func(item workItems.WorkItemHandler) {
			called = true
		},
	}
	ei, _ := NewDataIndexer(arguments)

	err := ei.SaveValidatorsRating(&outport.ValidatorsRating{})
	require.True(t, called)
	require.Nil(t, err)
}

func TestDataIndexer_RevertIndexedBlock(t *testing.T) {
	called := false

	arguments := NewDataIndexerArguments()
	arguments.DataDispatcher = &mock.DispatcherMock{
		AddCalled: func(item workItems.WorkItemHandler) {
			called = true
		},
	}
	ei, _ := NewDataIndexer(arguments)

	err := ei.RevertIndexedBlock(&outport.BlockData{
		HeaderType:  string(core.ShardHeaderV2),
		Body:        &dataBlock.Body{MiniBlocks: []*dataBlock.MiniBlock{{}}},
		HeaderBytes: []byte("{}"),
	})
	require.Nil(t, err)
	require.True(t, called)
}
