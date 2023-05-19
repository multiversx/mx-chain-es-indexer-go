package workItems

import (
	"github.com/multiversx/mx-chain-core-go/data/outport"
)

type itemAccounts struct {
	indexer  saveAccountsIndexer
	accounts *outport.Accounts
}

// NewItemAccounts will create a new instance of itemAccounts
func NewItemAccounts(
	indexer saveAccountsIndexer,
	accounts *outport.Accounts,
) WorkItemHandler {
	return &itemAccounts{
		indexer:  indexer,
		accounts: accounts,
	}
}

// Save will save information about an account
func (wiv *itemAccounts) Save() error {
	if wiv.accounts == nil {
		return nil
	}

	err := wiv.indexer.SaveAccounts(wiv.accounts)
	if err != nil {
		log.Warn("itemAccounts.Save",
			"could not index account",
			"error", err.Error())
		return err
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (wiv *itemAccounts) IsInterfaceNil() bool {
	return wiv == nil
}
