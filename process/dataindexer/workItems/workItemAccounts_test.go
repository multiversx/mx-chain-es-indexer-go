package workItems_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elastic-indexer-go/process/dataindexer/workItems"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/stretchr/testify/require"
)

func TestItemAccounts_Save(t *testing.T) {
	called := false
	itemAccounts := workItems.NewItemAccounts(
		&mock.ElasticProcessorStub{
			SaveAccountsCalled: func(_ uint64, _ []*data.Account) error {
				called = true
				return nil
			},
		},
		0,
		make(map[string]*outport.AlteredAccount), 0,
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
			SaveAccountsCalled: func(_ uint64, _ []*data.Account) error {
				return localErr
			},
		},
		0,
		make(map[string]*outport.AlteredAccount), 0,
	)
	require.False(t, itemAccounts.IsInterfaceNil())

	err := itemAccounts.Save()
	require.Equal(t, localErr, err)
}
