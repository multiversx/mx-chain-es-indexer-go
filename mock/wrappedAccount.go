package mock

// AccountWrapMock -
type AccountWrapMock struct {
}

// IsInterfaceNil -
func (awm *AccountWrapMock) IsInterfaceNil() bool {
	return awm == nil
}

// AddressBytes -
func (awm *AccountWrapMock) AddressBytes() []byte {
	return nil
}

// IncreaseNonce adds the given value to the current nonce
func (awm *AccountWrapMock) IncreaseNonce(_ uint64) {
}

// GetNonce gets the nonce of the account
func (awm *AccountWrapMock) GetNonce() uint64 {
	return 0
}
