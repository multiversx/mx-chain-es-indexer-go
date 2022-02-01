package workItems

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
)

type itemAccounts struct {
	indexer        saveAccountsIndexer
	blockTimestamp uint64
	accounts       map[string]*indexer.AlteredAccount
}

// NewItemAccounts will create a new instance of itemAccounts
func NewItemAccounts(
	indexer saveAccountsIndexer,
	blockTimestamp uint64,
	accounts map[string]*indexer.AlteredAccount,
) WorkItemHandler {
	return &itemAccounts{
		indexer:        indexer,
		accounts:       accounts,
		blockTimestamp: blockTimestamp,
	}
}

// Save will save information about an account
func (wiv *itemAccounts) Save() error {
	accounts := make([]*data.Account, len(wiv.accounts))
	idx := 0
	for _, account := range wiv.accounts {
		accounts[idx] = &data.Account{
			UserAccount: account,
			IsSender:    false,
		}
		idx++
	}

	err := wiv.indexer.SaveAccounts(wiv.blockTimestamp, accounts)
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
