package transactions

import (
	"encoding/hex"
	"testing"
	"time"

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
	header := &block.Header{TimeStamp: 1234}
	txs := map[string]*outport.TxInfo{
		hex.EncodeToString(txHash1): {
			Transaction: &transaction.Transaction{
				SndAddr: []byte("sender1"),
				RcvAddr: []byte("receiver1"),
			},
			FeeInfo: &outport.FeeInfo{},
		},
		hex.EncodeToString(txHash2): {
			Transaction: &transaction.Transaction{
				SndAddr: []byte("sender2"),
				RcvAddr: []byte("receiver2"),
			},
			FeeInfo: &outport.FeeInfo{},
		},
	}

	normalTxs, _ := grouper.groupNormalTxs(0, mb, header, txs, false, 3, 1234000)
	require.Len(t, normalTxs, 2)
	require.Equal(t, uint64(1234), normalTxs[string(txHash1)].Timestamp)
	require.Equal(t, uint64(1234000), normalTxs[string(txHash1)].TimestampMs)
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
	header := &block.Header{TimeStamp: 1234}
	txs := map[string]*outport.RewardInfo{
		hex.EncodeToString(txHash1): {Reward: &rewardTx.RewardTx{
			RcvAddr: []byte("receiver1"),
		}},
		hex.EncodeToString(txHash2): {Reward: &rewardTx.RewardTx{
			RcvAddr: []byte("receiver2"),
		}},
	}

	normalTxs, _ := grouper.groupRewardsTxs(0, mb, header, txs, false, 1234000)
	require.Len(t, normalTxs, 2)
	require.Equal(t, uint64(1234), normalTxs[string(txHash1)].Timestamp)
	require.Equal(t, uint64(1234000), normalTxs[string(txHash1)].TimestampMs)
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
	header := &block.Header{TimeStamp: 1234}
	txs := map[string]*outport.TxInfo{
		hex.EncodeToString(txHash1): {
			Transaction: &transaction.Transaction{
				SndAddr: []byte("sender1"),
				RcvAddr: []byte("receiver1"),
			}, FeeInfo: &outport.FeeInfo{}},
		hex.EncodeToString(txHash2): {
			Transaction: &transaction.Transaction{
				SndAddr: []byte("sender2"),
				RcvAddr: []byte("receiver2"),
			}, FeeInfo: &outport.FeeInfo{}},
	}

	normalTxs, _ := grouper.groupInvalidTxs(0, mb, header, txs, 3, 1234000)
	require.Len(t, normalTxs, 2)
	require.Equal(t, uint64(1234), normalTxs[string(txHash1)].Timestamp)
	require.Equal(t, uint64(1234000), normalTxs[string(txHash1)].TimestampMs)
}

func TestGroupReceipts(t *testing.T) {
	t.Parallel()

	parser := createDataFieldParserMock()
	ap, _ := converters.NewBalanceConverter(18)
	txBuilder := newTransactionDBBuilder(&mock.PubkeyConverterMock{}, parser, ap)
	grouper := newTxsGrouper(txBuilder, &mock.HasherMock{}, &mock.MarshalizerMock{})

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")
	header := &block.Header{TimeStamp: 1234}
	txs := map[string]*receipt.Receipt{
		hex.EncodeToString(txHash1): {
			SndAddr: []byte("sender1"),
		},
		hex.EncodeToString(txHash2): {
			SndAddr: []byte("sender2"),
		},
	}

	receipts := grouper.groupReceipts(header, txs, 1234000)
	require.Len(t, receipts, 2)
	require.Equal(t, time.Duration(1234), receipts[0].Timestamp)
	require.Equal(t, time.Duration(1234000), receipts[0].TimestampMs)
}
