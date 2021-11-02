package mock

import (
	"bytes"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
)

// DBTransactionProcessorStub -
type DBTransactionProcessorStub struct {
	PrepareTransactionsForDatabaseCalled func(body *block.Body, header coreData.HeaderHandler, pool *indexer.Pool) *data.PreparedResults
	SerializeReceiptsCalled              func(recs []*data.Receipt) ([]*bytes.Buffer, error)
	SerializeScResultsCalled             func(scrs []*data.ScResult) ([]*bytes.Buffer, error)
}

func (tps *DBTransactionProcessorStub) SerializeTransactionWithRefund(_ map[string]*data.Transaction, _ map[string]*data.RefundData) ([]*bytes.Buffer, error) {
	return nil, nil
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
func (tps *DBTransactionProcessorStub) SerializeReceipts(recs []*data.Receipt) ([]*bytes.Buffer, error) {
	if tps.SerializeReceiptsCalled != nil {
		return tps.SerializeReceiptsCalled(recs)
	}

	return nil, nil
}

// SerializeTransactions -
func (tps *DBTransactionProcessorStub) SerializeTransactions(_ []*data.Transaction, _ map[string]string, _ uint32, _ []*data.ScResult) ([]*bytes.Buffer, error) {
	return nil, nil
}

// SerializeScResults -
func (tps *DBTransactionProcessorStub) SerializeScResults(scrs []*data.ScResult) ([]*bytes.Buffer, error) {
	if tps.SerializeScResultsCalled != nil {
		return tps.SerializeScResultsCalled(scrs)
	}

	return nil, nil
}

// SerializeDeploysData -
func (tps *DBTransactionProcessorStub) SerializeDeploysData(_ []*data.ScDeployInfo) ([]*bytes.Buffer, error) {
	return nil, nil
}

// SerializeTokens -
func (tps *DBTransactionProcessorStub) SerializeTokens(_ []*data.TokenInfo) ([]*bytes.Buffer, error) {
	return nil, nil
}
