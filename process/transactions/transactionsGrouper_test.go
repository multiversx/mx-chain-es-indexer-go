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

	normalTxs, _ := grouper.groupNormalTxs(mb, header, txs, alteredAddresses)
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

	normalTxs, _ := grouper.groupRewardsTxs(mb, header, txs, alteredAddresses)
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

	normalTxs, _ := grouper.groupInvalidTxs(mb, header, txs, alteredAddresses)
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

func TestAddToAlteredAddressesNFT(t *testing.T) {
	t.Parallel()

	grouper := txsGrouper{
		txBuilder: &dbTransactionBuilder{
			esdtProc: newEsdtTransactionHandler(),
		},
	}

	alteredAddresses := map[string]*data.AlteredAccount{}
	mb := &block.MiniBlock{
		SenderShardID:   0,
		ReceiverShardID: 0,
	}
	grouper.addToAlteredAddresses(&data.Transaction{
		Data:                []byte("ESDTNFTTransfer@4d494841492d666437653066@01@01@b7a5acba50ff6a2821876693a4e62d60ec8645af696591e04ead2e2cb6e4cb4f"),
		Sender:              "sender",
		Receiver:            "sender",
		EsdtTokenIdentifier: "MY-TOKEN",
	}, alteredAddresses, mb, 0, false)
	require.Equal(t, &data.AlteredAccount{
		IsNFTOperation:  true,
		IsESDTOperation: false,
		IsSender:        true,
		TokenIdentifier: "MY-TOKEN",
		NFTNonceString:  "1",
	}, alteredAddresses["sender"])
}

func TestAddToAlteredAddressesESDT(t *testing.T) {
	t.Parallel()

	grouper := txsGrouper{
		txBuilder: &dbTransactionBuilder{
			esdtProc: newEsdtTransactionHandler(),
		},
	}

	alteredAddresses := map[string]*data.AlteredAccount{}
	mb := &block.MiniBlock{
		SenderShardID:   0,
		ReceiverShardID: 0,
	}
	grouper.addToAlteredAddresses(&data.Transaction{
		Data:                []byte("ESDTTransfer@31323334352d373066366534@174876e800"),
		Sender:              "sender",
		Receiver:            "sender",
		EsdtTokenIdentifier: "MY-TOKEN",
	}, alteredAddresses, mb, 0, false)
	require.Equal(t, &data.AlteredAccount{
		IsNFTOperation:  false,
		IsESDTOperation: true,
		IsSender:        true,
		TokenIdentifier: "MY-TOKEN",
		NFTNonceString:  "",
	}, alteredAddresses["sender"])
}
