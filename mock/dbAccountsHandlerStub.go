package mock

import (
	"bytes"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
)

// DBAccountsHandlerStub -
type DBAccountsHandlerStub struct {
	PrepareAccountsHistoryCalled   func(timestamp uint64, accounts map[string]*data.AccountInfo) map[string]*data.AccountBalanceHistory
	SerializeAccountsHistoryCalled func(accounts map[string]*data.AccountBalanceHistory) ([]*bytes.Buffer, error)
}

// GetAccounts -
func (dba *DBAccountsHandlerStub) GetAccounts(_ data.AlteredAccountsHandler, _ map[string]*indexer.AlteredAccount) ([]*data.Account, []*data.AccountESDT) {
	return nil, nil
}

// PrepareRegularAccountsMap -
func (dba *DBAccountsHandlerStub) PrepareRegularAccountsMap(_ []*data.Account) map[string]*data.AccountInfo {
	return nil
}

// PrepareAccountsMapESDT -
func (dba *DBAccountsHandlerStub) PrepareAccountsMapESDT(_ []*data.AccountESDT) map[string]*data.AccountInfo {
	return nil
}

// PrepareAccountsHistory -
func (dba *DBAccountsHandlerStub) PrepareAccountsHistory(timestamp uint64, accounts map[string]*data.AccountInfo) map[string]*data.AccountBalanceHistory {
	if dba.PrepareAccountsHistoryCalled != nil {
		return dba.PrepareAccountsHistoryCalled(timestamp, accounts)
	}

	return nil
}

// SerializeAccountsHistory -
func (dba *DBAccountsHandlerStub) SerializeAccountsHistory(accounts map[string]*data.AccountBalanceHistory) ([]*bytes.Buffer, error) {
	if dba.SerializeAccountsHistoryCalled != nil {
		return dba.SerializeAccountsHistoryCalled(accounts)
	}
	return nil, nil
}

// SerializeAccounts -
func (dba *DBAccountsHandlerStub) SerializeAccounts(_ map[string]*data.AccountInfo, _ bool) ([]*bytes.Buffer, error) {
	return nil, nil
}

// SerializeNFTCreateInfo -
func (dba *DBAccountsHandlerStub) SerializeNFTCreateInfo(_ []*data.TokenInfo) ([]*bytes.Buffer, error) {
	return nil, nil
}
