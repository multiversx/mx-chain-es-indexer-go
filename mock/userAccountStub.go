package mock

import (
	"math/big"
)

// UserAccountStub -
type UserAccountStub struct {
	GetBalanceCalled                       func() *big.Int
	GetNonceCalled                         func() uint64
	AddressBytesCalled                     func() []byte
	RetrieveValueFromDataTrieTrackerCalled func(key []byte) ([]byte, error)
}

// IncreaseNonce -
func (u *UserAccountStub) IncreaseNonce(_ uint64) {
}

// GetBalance -
func (u *UserAccountStub) GetBalance() *big.Int {
	if u.GetBalanceCalled != nil {
		return u.GetBalanceCalled()
	}
	return nil
}

func (u *UserAccountStub) RetrieveValueFromDataTrieTracker(key []byte) ([]byte, error) {
	if u.RetrieveValueFromDataTrieTrackerCalled != nil {
		return u.RetrieveValueFromDataTrieTrackerCalled(key)
	}

	return nil, nil
}

// AddressBytes -
func (u *UserAccountStub) AddressBytes() []byte {
	if u.AddressBytesCalled != nil {
		return u.AddressBytesCalled()
	}
	return nil
}

// GetNonce -
func (u *UserAccountStub) GetNonce() uint64 {
	if u.GetNonceCalled != nil {
		return u.GetNonceCalled()
	}
	return 0
}

// IsInterfaceNil -
func (u *UserAccountStub) IsInterfaceNil() bool {
	return false
}
