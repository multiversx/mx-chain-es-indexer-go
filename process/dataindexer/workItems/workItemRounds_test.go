package workItems_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer/workItems"
	"github.com/stretchr/testify/require"
)

func TestItemRounds_Save(t *testing.T) {
	called := false
	itemRounds := workItems.NewItemRounds(
		&mock.ElasticProcessorStub{
			SaveRoundsInfoCalled: func(_ *outport.RoundsInfo) error {
				called = true
				return nil
			},
		},
		&outport.RoundsInfo{},
	)
	require.False(t, itemRounds.IsInterfaceNil())

	err := itemRounds.Save()
	require.NoError(t, err)
	require.True(t, called)
}

func TestItemRounds_SaveRoundsShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	itemRounds := workItems.NewItemRounds(
		&mock.ElasticProcessorStub{
			SaveRoundsInfoCalled: func(_ *outport.RoundsInfo) error {
				return localErr
			},
		},
		&outport.RoundsInfo{},
	)
	require.False(t, itemRounds.IsInterfaceNil())

	err := itemRounds.Save()
	require.Equal(t, localErr, err)
}
