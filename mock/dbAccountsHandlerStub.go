package mock

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

// DBAccountsHandlerStub -
type DBAccountsHandlerStub struct {
	PrepareAccountsHistoryCalled   func(timestamp uint64, accounts map[string]*data.AccountInfo) map[string]*data.AccountBalanceHistory
	SerializeAccountsHistoryCalled func(accounts map[string]*data.AccountBalanceHistory, buffSlice *data.BufferSlice, index string) error
}

// GetAccounts -
func (dba *DBAccountsHandlerStub) GetAccounts(_ data.AlteredAccountsHandler) ([]*data.Account, []*data.AccountESDT) {
	return nil, nil
}

// PrepareRegularAccountsMap -
func (dba *DBAccountsHandlerStub) PrepareRegularAccountsMap(_ uint64, _ []*data.Account) map[string]*data.AccountInfo {
	return nil
}

// PrepareAccountsMapESDT -
func (dba *DBAccountsHandlerStub) PrepareAccountsMapESDT(_ uint64, _ []*data.AccountESDT, _ data.CountTags) (map[string]*data.AccountInfo, data.TokensHandler) {
	return nil, nil
}

// PrepareAccountsHistory -
func (dba *DBAccountsHandlerStub) PrepareAccountsHistory(timestamp uint64, accounts map[string]*data.AccountInfo) map[string]*data.AccountBalanceHistory {
	if dba.PrepareAccountsHistoryCalled != nil {
		return dba.PrepareAccountsHistoryCalled(timestamp, accounts)
	}

	return nil
}

// SerializeAccountsHistory -
func (dba *DBAccountsHandlerStub) SerializeAccountsHistory(accounts map[string]*data.AccountBalanceHistory, buffSlice *data.BufferSlice, index string) error {
	if dba.SerializeAccountsHistoryCalled != nil {
		return dba.SerializeAccountsHistoryCalled(accounts, buffSlice, index)
	}
	return nil
}

// SerializeAccounts -
func (dba *DBAccountsHandlerStub) SerializeAccounts(_ map[string]*data.AccountInfo, _ *data.BufferSlice, _ string) error {
	return nil
}

// SerializeAccountsESDT -
func (dba *DBAccountsHandlerStub) SerializeAccountsESDT(_ map[string]*data.AccountInfo, _ []*data.NFTDataUpdate, _ *data.BufferSlice, _ string) error {
	return nil
}

// SerializeNFTCreateInfo -
func (dba *DBAccountsHandlerStub) SerializeNFTCreateInfo(_ []*data.TokenInfo, _ *data.BufferSlice, _ string) error {
	return nil
}

// PutTokenMedataDataInTokens -
func (dba *DBAccountsHandlerStub) PutTokenMedataDataInTokens(_ []*data.TokenInfo) {
}

// SerializeTypeForProvidedIDs -
func (dba *DBAccountsHandlerStub) SerializeTypeForProvidedIDs(_ []string, _ string, _ *data.BufferSlice, _ string) error {
	return nil
}
