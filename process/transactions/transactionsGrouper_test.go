package transactions

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestGroupNormalTxs(t *testing.T) {
	t.Parallel()

	txBuilder := newTransactionDBBuilder(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{}, &mock.EconomicsHandlerStub{})
	grouper := newTxsGrouper(txBuilder, false, 0, &mock.HasherMock{}, &mock.MarshalizerMock{})

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	mb := &block.MiniBlock{
		TxHashes: [][]byte{txHash1, txHash2},
		Type:     block.TxBlock,
	}
	header := &block.Header{}
	txs := map[string]nodeData.TransactionHandler{
		string(txHash1): &transaction.Transaction{
			SndAddr: []byte("sender1"),
			RcvAddr: []byte("receiver1"),
		},
		string(txHash2): &transaction.Transaction{
			SndAddr: []byte("sender2"),
			RcvAddr: []byte("receiver2"),
		},
	}
	alteredAddresses := map[string]*data.AlteredAccount{}

	normalTxs := grouper.groupNormalTxs(mb, header, txs, alteredAddresses)
	require.Len(t, normalTxs, 2)
	require.Len(t, alteredAddresses, 4)
}

func TestGroupRewardsTxs(t *testing.T) {
	t.Parallel()

	txBuilder := newTransactionDBBuilder(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{}, &mock.EconomicsHandlerStub{})
	grouper := newTxsGrouper(txBuilder, false, 0, &mock.HasherMock{}, &mock.MarshalizerMock{})

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	mb := &block.MiniBlock{
		TxHashes: [][]byte{txHash1, txHash2},
		Type:     block.RewardsBlock,
	}
	header := &block.Header{}
	txs := map[string]nodeData.TransactionHandler{
		string(txHash1): &rewardTx.RewardTx{
			RcvAddr: []byte("receiver1"),
		},
		string(txHash2): &rewardTx.RewardTx{
			RcvAddr: []byte("receiver2"),
		},
	}
	alteredAddresses := map[string]*data.AlteredAccount{}

	normalTxs := grouper.groupRewardsTxs(mb, header, txs, alteredAddresses)
	require.Len(t, normalTxs, 2)
	require.Len(t, alteredAddresses, 2)
}

func TestGroupInvalidTxs(t *testing.T) {
	t.Parallel()

	txBuilder := newTransactionDBBuilder(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{}, &mock.EconomicsHandlerStub{})
	grouper := newTxsGrouper(txBuilder, false, 0, &mock.HasherMock{}, &mock.MarshalizerMock{})

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	mb := &block.MiniBlock{
		TxHashes: [][]byte{txHash1, txHash2},
		Type:     block.InvalidBlock,
	}
	header := &block.Header{}
	txs := map[string]nodeData.TransactionHandler{
		string(txHash1): &transaction.Transaction{
			SndAddr: []byte("sender1"),
			RcvAddr: []byte("receiver1"),
		},
		string(txHash2): &transaction.Transaction{
			SndAddr: []byte("sender2"),
			RcvAddr: []byte("receiver2"),
		},
	}
	alteredAddresses := map[string]*data.AlteredAccount{}

	normalTxs := grouper.groupInvalidTxs(mb, header, txs, alteredAddresses)
	require.Len(t, normalTxs, 2)
	require.Len(t, alteredAddresses, 2)
}

func TestGroupReceipts(t *testing.T) {
	t.Parallel()

	txBuilder := newTransactionDBBuilder(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{}, &mock.EconomicsHandlerStub{})
	grouper := newTxsGrouper(txBuilder, false, 0, &mock.HasherMock{}, &mock.MarshalizerMock{})

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	header := &block.Header{}
	txs := map[string]nodeData.TransactionHandler{
		string(txHash1): &receipt.Receipt{
			SndAddr: []byte("sender1"),
		},
		string(txHash2): &receipt.Receipt{
			SndAddr: []byte("sender2"),
		},
	}

	normalTxs := grouper.groupReceipts(header, txs)
	require.Len(t, normalTxs, 2)
}
