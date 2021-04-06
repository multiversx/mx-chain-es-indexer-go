package mock

import (
	"bytes"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/indexer"
)

// DBTransactionProcessorStub -
type DBTransactionProcessorStub struct {
	PrepareTransactionsForDatabaseCalled func(body *block.Body, header nodeData.HeaderHandler, pool *indexer.Pool) *data.PreparedResults
	SerializeReceiptsCalled              func(recs []*data.Receipt) ([]*bytes.Buffer, error)
	SerializeScResultsCalled             func(scrs []*data.ScResult) ([]*bytes.Buffer, error)
}

// PrepareTransactionsForDatabase -
func (tps *DBTransactionProcessorStub) PrepareTransactionsForDatabase(body *block.Body, header nodeData.HeaderHandler, pool *indexer.Pool) *data.PreparedResults {
	if tps.PrepareTransactionsForDatabaseCalled != nil {
		return tps.PrepareTransactionsForDatabaseCalled(body, header, pool)
	}

	return nil
}

// GetRewardsTxsHashesHexEncoded -
func (tps *DBTransactionProcessorStub) GetRewardsTxsHashesHexEncoded(_ nodeData.HeaderHandler, _ *block.Body) []string {
	return nil
}

// SerializeReceipts -
func (tps *DBTransactionProcessorStub) SerializeReceipts(recs []*data.Receipt) ([]*bytes.Buffer, error) {
	if tps.SerializeReceiptsCalled != nil {
		return tps.SerializeReceiptsCalled(recs)
	}

	return nil, nil
}

// SerializeTransactions -
func (tps *DBTransactionProcessorStub) SerializeTransactions(_ []*data.Transaction, _ uint32, _ map[string]bool) ([]*bytes.Buffer, error) {
	return nil, nil
}

// SerializeScResults -
func (tps *DBTransactionProcessorStub) SerializeScResults(scrs []*data.ScResult) ([]*bytes.Buffer, error) {
	if tps.SerializeScResultsCalled != nil {
		return tps.SerializeScResultsCalled(scrs)
	}

	return nil, nil
}
