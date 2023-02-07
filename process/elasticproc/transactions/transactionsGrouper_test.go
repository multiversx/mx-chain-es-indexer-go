package transactions

import (
	"math/big"
	"testing"

	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/receipt"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
	"github.com/stretchr/testify/require"
)

func TestGroupNormalTxs(t *testing.T) {
	t.Parallel()

	parser := createDataFieldParserMock()
	ap, _ := converters.NewBalanceConverter(18)
	txBuilder := newTransactionDBBuilder(&mock.PubkeyConverterMock{}, parser, ap)
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

	normalTxs, _ := grouper.groupNormalTxs(0, mb, header, txs, false, 3)
	require.Len(t, normalTxs, 2)
}

func TestGroupRewardsTxs(t *testing.T) {
	t.Parallel()

	parser := createDataFieldParserMock()
	ap, _ := converters.NewBalanceConverter(18)
	txBuilder := newTransactionDBBuilder(&mock.PubkeyConverterMock{}, parser, ap)
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

	normalTxs, _ := grouper.groupRewardsTxs(0, mb, header, txs, false)
	require.Len(t, normalTxs, 2)
}

func TestGroupInvalidTxs(t *testing.T) {
	t.Parallel()

	parser := createDataFieldParserMock()
	ap, _ := converters.NewBalanceConverter(18)
	txBuilder := newTransactionDBBuilder(mock.NewPubkeyConverterMock(32), parser, ap)
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

	normalTxs, _ := grouper.groupInvalidTxs(0, mb, header, txs, 3)
	require.Len(t, normalTxs, 2)
}

func TestGroupReceipts(t *testing.T) {
	t.Parallel()

	parser := createDataFieldParserMock()
	ap, _ := converters.NewBalanceConverter(18)
	txBuilder := newTransactionDBBuilder(&mock.PubkeyConverterMock{}, parser, ap)
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
