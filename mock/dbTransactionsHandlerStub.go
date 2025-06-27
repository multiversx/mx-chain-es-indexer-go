package mock

import (
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
)

// DBTransactionProcessorStub -
type DBTransactionProcessorStub struct {
	PrepareTransactionsForDatabaseCalled func(mbs []*block.MiniBlock, header coreData.HeaderHandler, pool *outport.TransactionPool) *data.PreparedResults
	SerializeReceiptsCalled              func(recs []*data.Receipt, buffSlice *data.BufferSlice, index string) error
	SerializeScResultsCalled             func(scrs []*data.ScResult, buffSlice *data.BufferSlice, index string) error
}

// SerializeTransactionsFeeData -
func (tps *DBTransactionProcessorStub) SerializeTransactionsFeeData(_ map[string]*data.FeeData, _ *data.BufferSlice, _ string) error {
	return nil
}

// PrepareTransactionsForDatabase -
func (tps *DBTransactionProcessorStub) PrepareTransactionsForDatabase(mbs []*block.MiniBlock, header coreData.HeaderHandler, pool *outport.TransactionPool, _ bool, _ uint32, _ uint64) *data.PreparedResults {
	if tps.PrepareTransactionsForDatabaseCalled != nil {
		return tps.PrepareTransactionsForDatabaseCalled(mbs, header, pool)
	}

	return nil
}

// GetHexEncodedHashesForRemove -
func (tps *DBTransactionProcessorStub) GetHexEncodedHashesForRemove(_ coreData.HeaderHandler, _ *block.Body) ([]string, []string) {
	return nil, nil
}

// SerializeReceipts -
func (tps *DBTransactionProcessorStub) SerializeReceipts(recs []*data.Receipt, buffSlice *data.BufferSlice, index string) error {
	if tps.SerializeReceiptsCalled != nil {
		return tps.SerializeReceiptsCalled(recs, buffSlice, index)
	}

	return nil
}

// SerializeTransactions -
func (tps *DBTransactionProcessorStub) SerializeTransactions(_ []*data.Transaction, _ map[string]*outport.StatusInfo, _ uint32, _ *data.BufferSlice, _ string) error {
	return nil
}

// SerializeScResults -
func (tps *DBTransactionProcessorStub) SerializeScResults(scrs []*data.ScResult, buffSlice *data.BufferSlice, index string) error {
	if tps.SerializeScResultsCalled != nil {
		return tps.SerializeScResultsCalled(scrs, buffSlice, index)
	}

	return nil
}

// SerializeDeploysData -
func (tps *DBTransactionProcessorStub) SerializeDeploysData(_ []*data.ScDeployInfo, _ *data.BufferSlice, _ string) error {
	return nil
}

// SerializeTokens -
func (tps *DBTransactionProcessorStub) SerializeTokens(_ []*data.TokenInfo, _ *data.BufferSlice, _ string) error {
	return nil
}
