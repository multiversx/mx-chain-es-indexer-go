package dataindexer

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/stretchr/testify/require"
)

func NewDataIndexerArguments() ArgDataIndexer {
	return ArgDataIndexer{
		ElasticProcessor: &mock.ElasticProcessorStub{},
		HeaderMarshaller: &mock.MarshalizerMock{},
		BlockContainer:   &mock.BlockContainerStub{},
	}
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
	countMap := map[int]int{}

	arguments := NewDataIndexerArguments()
	arguments.BlockContainer = &mock.BlockContainerStub{
		GetCalled: func(headerType core.HeaderType) (dataBlock.EmptyBlockCreator, error) {
			return dataBlock.NewEmptyHeaderV2Creator(), nil
		},
	}

	arguments.ElasticProcessor = &mock.ElasticProcessorStub{
		SaveHeaderCalled: func(outportBlockWithHeader *outport.OutportBlockWithHeader) error {
			countMap[0]++
			return nil
		},
		SaveMiniblocksCalled: func(header coreData.HeaderHandler, miniBlocks []*dataBlock.MiniBlock, timestampMs uint64) error {
			countMap[1]++
			return nil
		},
		SaveTransactionsCalled: func(outportBlockWithHeader *outport.OutportBlockWithHeader) error {
			countMap[2]++
			return nil
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
	require.Equal(t, 1, countMap[0])
	require.Equal(t, 1, countMap[1])
	require.Equal(t, 1, countMap[2])
}

func TestDataIndexer_SaveRoundInfo(t *testing.T) {
	called := false

	arguments := NewDataIndexerArguments()

	arguments.HeaderMarshaller = &mock.MarshalizerMock{Fail: true}
	arguments.ElasticProcessor = &mock.ElasticProcessorStub{
		SaveRoundsInfoCalled: func(infos *outport.RoundsInfo) error {
			called = true
			return nil
		},
	}
	ei, _ := NewDataIndexer(arguments)
	_ = ei.Close()

	err := ei.SaveRoundsInfo(&outport.RoundsInfo{})
	require.True(t, called)
	require.Nil(t, err)
}

func TestDataIndexer_SaveValidatorsPubKeys(t *testing.T) {
	called := false

	arguments := NewDataIndexerArguments()
	arguments.ElasticProcessor = &mock.ElasticProcessorStub{
		SaveShardValidatorsPubKeysCalled: func(validators *outport.ValidatorsPubKeys) error {
			called = true
			return nil
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
	arguments.ElasticProcessor = &mock.ElasticProcessorStub{
		SaveValidatorsRatingCalled: func(validatorsRating *outport.ValidatorsRating) error {
			called = true
			return nil
		},
	}
	ei, _ := NewDataIndexer(arguments)

	err := ei.SaveValidatorsRating(&outport.ValidatorsRating{})
	require.True(t, called)
	require.Nil(t, err)
}

func TestDataIndexer_RevertIndexedBlock(t *testing.T) {
	countMap := map[int]int{}

	arguments := NewDataIndexerArguments()
	arguments.BlockContainer = &mock.BlockContainerStub{
		GetCalled: func(headerType core.HeaderType) (dataBlock.EmptyBlockCreator, error) {
			return dataBlock.NewEmptyHeaderV2Creator(), nil
		}}
	arguments.ElasticProcessor = &mock.ElasticProcessorStub{
		RemoveHeaderCalled: func(header coreData.HeaderHandler) error {
			countMap[0]++
			return nil
		},
		RemoveMiniblocksCalled: func(header coreData.HeaderHandler, body *dataBlock.Body) error {
			countMap[1]++
			return nil
		},
		RemoveTransactionsCalled: func(header coreData.HeaderHandler, body *dataBlock.Body) error {
			countMap[2]++
			return nil
		},
		RemoveAccountsESDTCalled: func(_ uint32, _ uint64) error {
			countMap[3]++
			return nil
		},
	}
	ei, _ := NewDataIndexer(arguments)

	err := ei.RevertIndexedBlock(&outport.BlockData{
		HeaderType:  string(core.ShardHeaderV2),
		Body:        &dataBlock.Body{MiniBlocks: []*dataBlock.MiniBlock{{}}},
		HeaderBytes: []byte("{}"),
	})
	require.Nil(t, err)
	require.Equal(t, 1, countMap[0])
	require.Equal(t, 1, countMap[1])
	require.Equal(t, 1, countMap[2])
	require.Equal(t, 1, countMap[3])
}
