package logsevents

import (
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

type txHashStatusInfoProc struct {
	hashStatusInfo map[string]*outport.StatusInfo
}

// newTxHashStatusInfoProcessor will create a new instance of txHashStatusInfoProc
func newTxHashStatusInfoProcessor() *txHashStatusInfoProc {
	return &txHashStatusInfoProc{
		hashStatusInfo: make(map[string]*outport.StatusInfo),
	}
}

// addRecord will add a new record for the given hash
func (ths *txHashStatusInfoProc) addRecord(hash string, statusInfo *outport.StatusInfo) {
	statusInfoFromMap, found := ths.hashStatusInfo[hash]
	if !found {
		ths.hashStatusInfo[hash] = statusInfo
		return
	}

	if statusInfoFromMap.Status != transaction.TxStatusFail.String() {
		statusInfoFromMap.Status = statusInfo.Status
	}

	statusInfoFromMap.ErrorEvent = statusInfoFromMap.ErrorEvent || statusInfo.ErrorEvent
	statusInfoFromMap.CompletedEvent = statusInfoFromMap.CompletedEvent || statusInfo.CompletedEvent
}

// getAllRecords will return all the records
func (ths *txHashStatusInfoProc) getAllRecords() map[string]*outport.StatusInfo {
	return ths.hashStatusInfo
}
