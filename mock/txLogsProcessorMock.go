package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.TransactionLogProcessorDatabase = (*TxLogProcessorStub)(nil)

// TxLogProcessorStub -
type TxLogProcessorStub struct {
	GetLogFromCacheCalled func(txHash []byte) (data.LogHandler, bool)
}

// GetLogFromCache -
func (t *TxLogProcessorStub) GetLogFromCache(txHash []byte) (data.LogHandler, bool) {
	if t.GetLogFromCacheCalled != nil {
		return t.GetLogFromCacheCalled(txHash)
	}

	return nil, false
}

// EnableLogToBeSavedInCache -
func (t *TxLogProcessorStub) EnableLogToBeSavedInCache() {
	return
}

// Clean -
func (t *TxLogProcessorStub) Clean() {
	return
}

// IsInterfaceNil -
func (t *TxLogProcessorStub) IsInterfaceNil() bool {
	return t == nil
}
