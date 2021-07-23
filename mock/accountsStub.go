package mock

import (
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// AccountsStub -
type AccountsStub struct {
	LoadAccountCalled func(container []byte) (vmcommon.AccountHandler, error)
}

// LoadAccount -
func (as *AccountsStub) LoadAccount(address []byte) (vmcommon.AccountHandler, error) {
	if as.LoadAccountCalled != nil {
		return as.LoadAccountCalled(address)
	}
	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (as *AccountsStub) IsInterfaceNil() bool {
	return as == nil
}
