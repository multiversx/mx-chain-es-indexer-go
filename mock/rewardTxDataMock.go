package mock

// RewardTxDataMock -
type RewardTxDataMock struct {
	GetSenderCalled func() string
}

// GetSender -
func (rtd *RewardTxDataMock) GetSender() string {
	if rtd.GetSenderCalled != nil {
		return rtd.GetSenderCalled()
	}

	return ""
}

// IsInterfaceNil returns true if there is no value under the interface
func (rtd *RewardTxDataMock) IsInterfaceNil() bool {
	return rtd == nil
}
