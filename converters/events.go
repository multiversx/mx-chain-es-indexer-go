package converters

import "github.com/ElrondNetwork/elastic-indexer-go/data"

// ConvertSliceTxsInMap will convert slice of provided transaction in a map with key transaction hash hex encoded
func ConvertSliceTxsInMap(txs []*data.Transaction) map[string]*data.Transaction {
	mapTxs := make(map[string]*data.Transaction, len(txs))

	for idx := 0; idx < len(txs); idx++ {
		tx := txs[idx]
		mapTxs[tx.Hash] = tx
	}

	return mapTxs
}

// ConvertSliceScrInMap will convert slice of provided smart contract results in a map with key scr hash hex encoded
func ConvertSliceScrInMap(scrs []*data.ScResult) map[string]*data.ScResult {
	mapSCRs := make(map[string]*data.ScResult, len(scrs))

	for idx := 0; idx < len(scrs); idx++ {
		scr := scrs[idx]
		mapSCRs[scr.Hash] = scr
	}

	return mapSCRs
}
