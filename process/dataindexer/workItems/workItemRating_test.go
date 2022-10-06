package workItems_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elastic-indexer-go/process/dataindexer/workItems"
	"github.com/stretchr/testify/require"
)

func TestItemRating_Save(t *testing.T) {
	id := "0_1"
	called := false
	itemRating := workItems.NewItemRating(
		&mock.ElasticProcessorStub{
			SaveValidatorsRatingCalled: func(index string, validatorsRatingInfo []*data.ValidatorRatingInfo) error {
				require.Equal(t, id, index)
				called = true
				return nil
			},
		},
		id,
		[]*data.ValidatorRatingInfo{
			{PublicKey: "pub-key", Rating: 100},
		},
	)
	require.False(t, itemRating.IsInterfaceNil())

	err := itemRating.Save()
	require.NoError(t, err)
	require.True(t, called)
}

func TestItemRating_SaveShouldErr(t *testing.T) {
	id := "0_1"
	localErr := errors.New("local err")
	itemRating := workItems.NewItemRating(
		&mock.ElasticProcessorStub{
			SaveValidatorsRatingCalled: func(index string, validatorsRatingInfo []*data.ValidatorRatingInfo) error {
				return localErr
			},
		},
		id,
		[]*data.ValidatorRatingInfo{
			{PublicKey: "pub-key", Rating: 100},
		},
	)
	require.False(t, itemRating.IsInterfaceNil())

	err := itemRating.Save()
	require.Equal(t, localErr, err)
}
