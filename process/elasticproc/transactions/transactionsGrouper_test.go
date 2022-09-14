package transactions

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/receipt"
	"github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestGroupNormalTxs(t *testing.T) {
	t.Parallel()

	parser := createDataFieldParserMock()
	txBuilder := newTransactionDBBuilder(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{}, parser)
	grouper := newTxsGrouper(txBuilder, &mock.HasherMock{}, &mock.MarshalizerMock{})

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	mb := &block.MiniBlock{
		TxHashes: [][]byte{txHash1, txHash2},
		Type:     block.TxBlock,
	}
	header := &block.Header{}
	txs := map[string]coreData.TransactionHandlerWithGasUsedAndFee{
		string(txHash1): outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
			SndAddr: []byte("sender1"),
			RcvAddr: []byte("receiver1"),
		}, 0, big.NewInt(0)),
		string(txHash2): outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
			SndAddr: []byte("sender2"),
			RcvAddr: []byte("receiver2"),
		}, 0, big.NewInt(0)),
	}
	alteredAddresses := data.NewAlteredAccounts()

	normalTxs, _ := grouper.groupNormalTxs(0, mb, header, txs, alteredAddresses, false)
	require.Len(t, normalTxs, 2)
	require.Equal(t, 4, alteredAddresses.Len())
}

func TestGroupRewardsTxs(t *testing.T) {
	t.Parallel()

	parser := createDataFieldParserMock()
	txBuilder := newTransactionDBBuilder(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{}, parser)
	grouper := newTxsGrouper(txBuilder, &mock.HasherMock{}, &mock.MarshalizerMock{})

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	mb := &block.MiniBlock{
		TxHashes: [][]byte{txHash1, txHash2},
		Type:     block.RewardsBlock,
	}
	header := &block.Header{}
	txs := map[string]coreData.TransactionHandlerWithGasUsedAndFee{
		string(txHash1): outport.NewTransactionHandlerWithGasAndFee(&rewardTx.RewardTx{
			RcvAddr: []byte("receiver1"),
		}, 0, big.NewInt(0)),
		string(txHash2): outport.NewTransactionHandlerWithGasAndFee(&rewardTx.RewardTx{
			RcvAddr: []byte("receiver2"),
		}, 0, big.NewInt(0)),
	}
	alteredAddresses := data.NewAlteredAccounts()

	normalTxs, _ := grouper.groupRewardsTxs(0, mb, header, txs, alteredAddresses, false)
	require.Len(t, normalTxs, 2)
	require.Equal(t, 2, alteredAddresses.Len())
}

func TestGroupInvalidTxs(t *testing.T) {
	t.Parallel()

	parser := createDataFieldParserMock()
	txBuilder := newTransactionDBBuilder(mock.NewPubkeyConverterMock(32), &mock.ShardCoordinatorMock{}, parser)
	grouper := newTxsGrouper(txBuilder, &mock.HasherMock{}, &mock.MarshalizerMock{})

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	mb := &block.MiniBlock{
		TxHashes: [][]byte{txHash1, txHash2},
		Type:     block.InvalidBlock,
	}
	header := &block.Header{}
	txs := map[string]coreData.TransactionHandlerWithGasUsedAndFee{
		string(txHash1): outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
			SndAddr: []byte("sender1"),
			RcvAddr: []byte("receiver1"),
		}, 0, big.NewInt(0)),
		string(txHash2): outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
			SndAddr: []byte("sender2"),
			RcvAddr: []byte("receiver2"),
		}, 0, big.NewInt(0)),
	}
	alteredAddresses := data.NewAlteredAccounts()

	normalTxs, _ := grouper.groupInvalidTxs(0, mb, header, txs, alteredAddresses)
	require.Len(t, normalTxs, 2)
	require.Equal(t, 2, alteredAddresses.Len())
}

func TestGroupReceipts(t *testing.T) {
	t.Parallel()

	parser := createDataFieldParserMock()
	txBuilder := newTransactionDBBuilder(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{}, parser)
	grouper := newTxsGrouper(txBuilder, &mock.HasherMock{}, &mock.MarshalizerMock{})

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	header := &block.Header{}
	txs := map[string]coreData.TransactionHandlerWithGasUsedAndFee{
		string(txHash1): outport.NewTransactionHandlerWithGasAndFee(&receipt.Receipt{
			SndAddr: []byte("sender1"),
		}, 0, big.NewInt(0)),
		string(txHash2): outport.NewTransactionHandlerWithGasAndFee(&receipt.Receipt{
			SndAddr: []byte("sender2"),
		}, 0, big.NewInt(0)),
	}

	normalTxs := grouper.groupReceipts(header, txs)
	require.Len(t, normalTxs, 2)
}
