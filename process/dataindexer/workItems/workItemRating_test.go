package workItems_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer/workItems"
	"github.com/stretchr/testify/require"
)

func TestItemRating_Save(t *testing.T) {
	called := false
	itemRating := workItems.NewItemRating(
		&mock.ElasticProcessorStub{
			SaveValidatorsRatingCalled: func(_ *outport.ValidatorsRating) error {
				called = true
				return nil
			},
		},
		&outport.ValidatorsRating{},
	)
	require.False(t, itemRating.IsInterfaceNil())

	err := itemRating.Save()
	require.NoError(t, err)
	require.True(t, called)
}

func TestItemRating_SaveShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	itemRating := workItems.NewItemRating(
		&mock.ElasticProcessorStub{
			SaveValidatorsRatingCalled: func(_ *outport.ValidatorsRating) error {
				return localErr
			},
		},
		&outport.ValidatorsRating{},
	)
	require.False(t, itemRating.IsInterfaceNil())

	err := itemRating.Save()
	require.Equal(t, localErr, err)
}
