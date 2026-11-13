package converters

import "github.com/multiversx/mx-chain-es-indexer-go/data"

// ConvertTxsSliceIntoMap will convert the slice of the provided transactions into a map where the key represents the hash of the transaction and the value is the transaction
func ConvertTxsSliceIntoMap(txs []*data.Transaction) map[string]*data.Transaction {
	mapTxs := make(map[string]*data.Transaction, len(txs))

	for _, tx := range txs {
		mapTxs[tx.Hash] = tx
	}

	return mapTxs
}

// ConvertScrsSliceIntoMap will convert the slice of provided smart contract results into a map where the key represents the hash of the scr and the value is the scr.
func ConvertScrsSliceIntoMap(scrs []*data.ScResult) map[string]*data.ScResult {
	mapSCRs := make(map[string]*data.ScResult, len(scrs))

	for _, scr := range scrs {
		mapSCRs[scr.Hash] = scr
	}

	return mapSCRs
}
