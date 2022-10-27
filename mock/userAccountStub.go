package mock

import (
	"math/big"
)

// UserAccountStub -
type UserAccountStub struct {
	GetBalanceCalled    func() *big.Int
	GetNonceCalled      func() uint64
	AddressBytesCalled  func() []byte
	RetrieveValueCalled func(key []byte) ([]byte, uint32, error)
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

// RetrieveValue -
func (u *UserAccountStub) RetrieveValue(key []byte) ([]byte, uint32, error) {
	if u.RetrieveValueCalled != nil {
		return u.RetrieveValueCalled(key)
	}

	return nil, 0, nil
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
