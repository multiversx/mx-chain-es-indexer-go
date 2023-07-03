package logsevents

import (
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
)

type txHashStatusInfoProc struct {
	hashStatusInfo map[string]*data.StatusInfo
}

// NewTxHashStatusInfo will create a new instance of TxHashStatusInfo
func newTxHashStatusInfo() *txHashStatusInfoProc {
	return &txHashStatusInfoProc{
		hashStatusInfo: make(map[string]*data.StatusInfo),
	}
}

// AddRecord will add a new record for the given hash
func (ths *txHashStatusInfoProc) addRecord(hash string, statusInfo *data.StatusInfo) {
	statusInfoFromMap, found := ths.hashStatusInfo[hash]
	if !found {
		ths.hashStatusInfo[hash] = statusInfo
	}

	if statusInfoFromMap.Status != transaction.TxStatusFail.String() {
		statusInfoFromMap.Status = statusInfo.Status
	}

	statusInfoFromMap.ErrorEvent = statusInfoFromMap.ErrorEvent || statusInfo.ErrorEvent
	statusInfoFromMap.CompletedEvent = statusInfoFromMap.CompletedEvent || statusInfo.CompletedEvent
}

// GetAllRecords will return all the records
func (ths *txHashStatusInfoProc) getAllRecords() map[string]*data.StatusInfo {
	return ths.hashStatusInfo
}
