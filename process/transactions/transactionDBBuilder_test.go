package transactions

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon/economicsmocks"
	"github.com/stretchr/testify/require"
)

func createCommonProcessor() dbTransactionBuilder {
	return dbTransactionBuilder{
		addressPubkeyConverter: mock.NewPubkeyConverterMock(32),
		txFeeCalculator: &economicsmocks.EconomicsHandlerStub{
			ComputeTxFeeBasedOnGasUsedCalled: func(tx process.TransactionWithFeeHandler, gasUsed uint64) *big.Int {
				return big.NewInt(100)
			},
			ComputeGasLimitCalled: func(tx process.TransactionWithFeeHandler) uint64 {
				return 500
			},
		},
		shardCoordinator: &mock.ShardCoordinatorMock{},
		esdtProc:         newEsdtTransactionHandler(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{}),
	}
}

func TestGetMoveBalanceTransaction(t *testing.T) {
	t.Parallel()

	txHash := []byte("txHash")
	mbHash := []byte("mbHash")
	mb := &block.MiniBlock{TxHashes: [][]byte{txHash}}
	header := &block.Header{Nonce: 2}
	status := "Success"
	gasPrice := uint64(1000)
	gasLimit := uint64(1000)
	cp := createCommonProcessor()

	tx := &transaction.Transaction{
		Nonce:       1,
		Value:       big.NewInt(1000),
		RcvAddr:     []byte("receiver"),
		SndAddr:     []byte("sender"),
		GasPrice:    gasPrice,
		GasLimit:    gasLimit,
		Data:        []byte("data"),
		ChainID:     []byte("1"),
		Version:     1,
		Signature:   []byte("signature"),
		RcvUserName: []byte("rcv"),
		SndUserName: []byte("snd"),
	}

	expectedTx := &data.Transaction{
		Hash:                 hex.EncodeToString(txHash),
		MBHash:               hex.EncodeToString(mbHash),
		Nonce:                tx.Nonce,
		Round:                header.Round,
		Value:                tx.Value.String(),
		Receiver:             cp.addressPubkeyConverter.Encode(tx.RcvAddr),
		Sender:               cp.addressPubkeyConverter.Encode(tx.SndAddr),
		ReceiverShard:        mb.ReceiverShardID,
		SenderShard:          mb.SenderShardID,
		GasPrice:             gasPrice,
		GasLimit:             gasLimit,
		GasUsed:              uint64(500),
		Data:                 tx.Data,
		Signature:            hex.EncodeToString(tx.Signature),
		Timestamp:            time.Duration(header.GetTimeStamp()),
		Status:               status,
		ReceiverAddressBytes: []byte("receiver"),
		Fee:                  "100",
		ReceiverUserName:     []byte("rcv"),
		SenderUserName:       []byte("snd"),
	}

	dbTx := cp.prepareTransaction(tx, txHash, mbHash, mb, header, status)
	require.Equal(t, expectedTx, dbTx)
}

func TestGetTransactionByType_SC(t *testing.T) {
	t.Parallel()

	cp := createCommonProcessor()

	nonce := uint64(10)
	txHash := []byte("txHash")
	code := []byte("code")
	sndAddr, rcvAddr := []byte("snd"), []byte("rec")
	scHash := "scHash"
	smartContractRes := &smartContractResult.SmartContractResult{
		Nonce:      nonce,
		PrevTxHash: txHash,
		Code:       code,
		Data:       []byte(""),
		SndAddr:    sndAddr,
		RcvAddr:    rcvAddr,
		CallType:   vmcommon.CallType(1),
	}
	header := &block.Header{TimeStamp: 100}

	scRes := cp.prepareSmartContractResult(scHash, smartContractRes, header)
	expectedTx := &data.ScResult{
		Nonce:      nonce,
		Hash:       hex.EncodeToString([]byte(scHash)),
		PrevTxHash: hex.EncodeToString(txHash),
		Code:       string(code),
		Data:       make([]byte, 0),
		Sender:     cp.addressPubkeyConverter.Encode(sndAddr),
		Receiver:   cp.addressPubkeyConverter.Encode(rcvAddr),
		Value:      "<nil>",
		CallType:   "1",
		Timestamp:  time.Duration(100),
	}

	require.Equal(t, expectedTx, scRes)
}

func TestGetTransactionByType_RewardTx(t *testing.T) {
	t.Parallel()

	cp := createCommonProcessor()

	round := uint64(10)
	rcvAddr := []byte("receiver")
	rwdTx := &rewardTx.RewardTx{Round: round, RcvAddr: rcvAddr}
	txHash := []byte("txHash")
	mbHash := []byte("mbHash")
	mb := &block.MiniBlock{TxHashes: [][]byte{txHash}}
	header := &block.Header{Nonce: 2}
	status := "Success"

	resultTx := cp.prepareRewardTransaction(rwdTx, txHash, mbHash, mb, header, status)
	expectedTx := &data.Transaction{
		Hash:     hex.EncodeToString(txHash),
		MBHash:   hex.EncodeToString(mbHash),
		Round:    round,
		Receiver: hex.EncodeToString(rcvAddr),
		Status:   status,
		Value:    "<nil>",
		Sender:   fmt.Sprintf("%d", core.MetachainShardId),
		Data:     make([]byte, 0),
	}

	require.Equal(t, expectedTx, resultTx)
}

func TestAddScrsReceiverToAlteredAccounts_ShouldWork(t *testing.T) {
	t.Parallel()

	txBuilder := newTransactionDBBuilder(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{}, &mock.EconomicsHandlerStub{})

	alteredAddress := data.NewAlteredAccounts()
	scrs := []*data.ScResult{
		{
			Sender:              "sender",
			Receiver:            "receiver",
			EsdtTokenIdentifier: "my-token",
			Data:                []byte("ESDTTransfer@544b4e2d626231323061@010f0cf064dd59200000"),
		},
	}
	txBuilder.addScrsReceiverToAlteredAccounts(alteredAddress, scrs)
	require.Equal(t, 2, alteredAddress.Len())

	_, ok := alteredAddress.Get("sender")
	require.True(t, ok)

	_, ok = alteredAddress.Get("receiver")
	require.True(t, ok)
}
