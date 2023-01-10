package workItems

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/multiversx/mx-chain-core-go/data/outport"
)

type itemAccounts struct {
	indexer        saveAccountsIndexer
	blockTimestamp uint64
	accounts       map[string]*outport.AlteredAccount
	shardID        uint32
}

// NewItemAccounts will create a new instance of itemAccounts
func NewItemAccounts(
	indexer saveAccountsIndexer,
	blockTimestamp uint64,
	accounts map[string]*outport.AlteredAccount,
	shardID uint32,
) WorkItemHandler {
	return &itemAccounts{
		indexer:        indexer,
		accounts:       accounts,
		blockTimestamp: blockTimestamp,
		shardID:        shardID,
	}
}

// Save will save information about an account
func (wiv *itemAccounts) Save() error {
	accounts := make([]*data.Account, 0, len(wiv.accounts))
	for _, account := range wiv.accounts {
		accounts = append(accounts, &data.Account{
			UserAccount: account,
			IsSender:    false,
		})
	}

	err := wiv.indexer.SaveAccounts(wiv.blockTimestamp, accounts, wiv.shardID)
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
