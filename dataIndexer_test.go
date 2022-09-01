package indexer

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elastic-indexer-go/workItems"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/stretchr/testify/require"
)

func NewDataIndexerArguments() ArgDataIndexer {
	return ArgDataIndexer{
		Marshalizer:      &mock.MarshalizerMock{},
		DataDispatcher:   &mock.DispatcherMock{},
		ElasticProcessor: &mock.ElasticProcessorStub{},
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
	arguments.Marshalizer = nil
	ei, err := NewDataIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, core.ErrNilMarshalizer, err)
}

func TestDataIndexer_NewIndexerWithCorrectParamsShouldWork(t *testing.T) {
	arguments := NewDataIndexerArguments()

	ei, err := NewDataIndexer(arguments)

	require.Nil(t, err)
	require.False(t, check.IfNil(ei))
	require.False(t, ei.IsNilIndexer())
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

	args := &outport.ArgsSaveBlockData{
		HeaderHash:             []byte("hash"),
		Body:                   &dataBlock.Body{MiniBlocks: []*dataBlock.MiniBlock{}},
		Header:                 nil,
		SignersIndexes:         nil,
		NotarizedHeadersHashes: nil,
		TransactionsPool:       nil,
	}
	err := ei.SaveBlock(args)
	require.True(t, called)
	require.Nil(t, err)
}

func TestDataIndexer_SaveRoundInfo(t *testing.T) {
	called := false

	arguments := NewDataIndexerArguments()
	arguments.DataDispatcher = &mock.DispatcherMock{
		AddCalled: func(item workItems.WorkItemHandler) {
			called = true
		},
	}

	arguments.Marshalizer = &mock.MarshalizerMock{Fail: true}
	ei, _ := NewDataIndexer(arguments)
	_ = ei.Close()

	err := ei.SaveRoundsInfo([]*outport.RoundInfo{})
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
	epoch := uint32(0)

	err := ei.SaveValidatorsPubKeys(valPubKey, epoch)
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

	err := ei.SaveValidatorsRating("ID", []*outport.ValidatorRatingInfo{
		{Rating: 1}, {Rating: 2},
	})
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

	err := ei.RevertIndexedBlock(&dataBlock.Header{}, &dataBlock.Body{})
	require.True(t, called)
	require.Nil(t, err)
}
