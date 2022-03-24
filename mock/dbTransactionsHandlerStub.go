package mock

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
)

// DBTransactionProcessorStub -
type DBTransactionProcessorStub struct {
	PrepareTransactionsForDatabaseCalled func(body *block.Body, header coreData.HeaderHandler, pool *indexer.Pool) *data.PreparedResults
	SerializeReceiptsCalled              func(recs []*data.Receipt, buffSlice *data.BufferSlice, index string) error
	SerializeScResultsCalled             func(scrs []*data.ScResult, buffSlice *data.BufferSlice, index string) error
}

func (tps *DBTransactionProcessorStub) SerializeTransactionWithRefund(_ map[string]*data.Transaction, _ map[string]*data.RefundData, _ *data.BufferSlice, _ string) error {
	return nil
}

// PrepareTransactionsForDatabase -
func (tps *DBTransactionProcessorStub) PrepareTransactionsForDatabase(body *block.Body, header coreData.HeaderHandler, pool *indexer.Pool) *data.PreparedResults {
	if tps.PrepareTransactionsForDatabaseCalled != nil {
		return tps.PrepareTransactionsForDatabaseCalled(body, header, pool)
	}

	return nil
}

// GetRewardsTxsHashesHexEncoded -
func (tps *DBTransactionProcessorStub) GetRewardsTxsHashesHexEncoded(_ coreData.HeaderHandler, _ *block.Body) []string {
	return nil
}

// SerializeReceipts -
func (tps *DBTransactionProcessorStub) SerializeReceipts(recs []*data.Receipt, buffSlice *data.BufferSlice, index string) error {
	if tps.SerializeReceiptsCalled != nil {
		return tps.SerializeReceiptsCalled(recs, buffSlice, index)
	}

	return nil
}

// SerializeTransactions -
func (tps *DBTransactionProcessorStub) SerializeTransactions(_ []*data.Transaction, _ map[string]string, _ uint32, _ *data.BufferSlice, _ string) error {
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
