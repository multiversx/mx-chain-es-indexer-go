package workItems_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer/workItems"
	"github.com/stretchr/testify/require"
)

func TestItemValidators_Save(t *testing.T) {
	called := false
	itemValidators := workItems.NewItemValidators(
		&mock.ElasticProcessorStub{
			SaveShardValidatorsPubKeysCalled: func(_ *outport.ValidatorsPubKeys) error {
				called = true
				return nil
			},
		},
		&outport.ValidatorsPubKeys{},
	)
	require.False(t, itemValidators.IsInterfaceNil())

	err := itemValidators.Save()
	require.NoError(t, err)
	require.True(t, called)
}

func TestItemValidators_SaveValidatorsShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	itemValidators := workItems.NewItemValidators(
		&mock.ElasticProcessorStub{
			SaveShardValidatorsPubKeysCalled: func(_ *outport.ValidatorsPubKeys) error {
				return localErr
			},
		},
		&outport.ValidatorsPubKeys{},
	)
	require.False(t, itemValidators.IsInterfaceNil())

	err := itemValidators.Save()
	require.Equal(t, localErr, err)
}
