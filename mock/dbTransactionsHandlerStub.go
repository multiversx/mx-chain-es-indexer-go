package mock

import (
	"bytes"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/indexer"
)

type DBTransactionProcessorStub struct {
}

func (tps *DBTransactionProcessorStub) PrepareTransactionsForDatabase(_ *block.Body, _ nodeData.HeaderHandler, _ *indexer.Pool) *data.PreparedResults {
	panic("implement me")
}

func (tps *DBTransactionProcessorStub) GetRewardsTxsHashesHexEncoded(_ nodeData.HeaderHandler, _ *block.Body) []string {
	panic("implement me")
}

func (tps *DBTransactionProcessorStub) SerializeReceipts(_ []*data.Receipt) ([]*bytes.Buffer, error) {
	panic("implement me")
}

func (tps *DBTransactionProcessorStub) SerializeTransactions(_ []*data.Transaction, _ uint32, _ map[string]bool) ([]*bytes.Buffer, error) {
	panic("implement me")
}

func (tps *DBTransactionProcessorStub) SerializeScResults(_ []*data.ScResult) ([]*bytes.Buffer, error) {
	panic("implement me")
}
