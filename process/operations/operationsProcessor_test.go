package operations

import (
	"testing"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestNewOperationsProcessor(t *testing.T) {
	t.Parallel()

	op, err := NewOperationsProcessor(false, nil)
	require.Nil(t, op)
	require.Equal(t, indexer.ErrNilShardCoordinator, err)

	op, err = NewOperationsProcessor(false, &mock.ShardCoordinatorMock{})
	require.NotNil(t, op)
	require.Nil(t, err)
}

func TestOperationsProcessor_ProcessTransactionsAndSCRSTransactions(t *testing.T) {
	t.Parallel()

	op, _ := NewOperationsProcessor(true, &mock.ShardCoordinatorMock{})

	txs := []*data.Transaction{
		{},
		{
			ReceiverShard: 1,
		},
	}

	processedTxs, _ := op.ProcessTransactionsAndSCRS(txs, nil)
	require.Equal(t, []*data.Transaction{
		{Type: string(transaction.TxTypeNormal)},
	}, processedTxs)
}

func TestOperationsProcessor_ProcessTransactionsAndSCRSSmartContractResults(t *testing.T) {
	t.Parallel()

	op, _ := NewOperationsProcessor(true, &mock.ShardCoordinatorMock{})

	scrs := []*data.ScResult{
		{},
		{
			ReceiverShard: 1,
		},
	}

	_, processedSCRs := op.ProcessTransactionsAndSCRS(nil, scrs)
	require.Equal(t, []*data.ScResult{
		{Type: string(transaction.TxTypeUnsigned), Status: transaction.TxStatusSuccess.String()},
	}, processedSCRs)
}
