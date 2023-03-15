package workItems_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer/workItems"
	"github.com/stretchr/testify/require"
)

func TestItemAccounts_Save(t *testing.T) {
	called := false
	itemAccounts := workItems.NewItemAccounts(
		&mock.ElasticProcessorStub{
			SaveAccountsCalled: func(_ *outport.Accounts) error {
				called = true
				return nil
			},
		},
		&outport.Accounts{},
	)
	require.False(t, itemAccounts.IsInterfaceNil())

	err := itemAccounts.Save()
	require.NoError(t, err)
	require.True(t, called)
}

func TestItemAccounts_SaveAccountsShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	itemAccounts := workItems.NewItemAccounts(
		&mock.ElasticProcessorStub{
			SaveAccountsCalled: func(_ *outport.Accounts) error {
				return localErr
			},
		},
		&outport.Accounts{},
	)
	require.False(t, itemAccounts.IsInterfaceNil())

	err := itemAccounts.Save()
	require.Equal(t, localErr, err)
}
