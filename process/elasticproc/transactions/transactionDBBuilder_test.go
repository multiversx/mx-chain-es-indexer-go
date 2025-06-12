package transactions

import (
	"encoding/hex"
	"fmt"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func createCommonProcessor() dbTransactionBuilder {
	ap, _ := converters.NewBalanceConverter(18)
	return dbTransactionBuilder{
		addressPubkeyConverter: mock.NewPubkeyConverterMock(32),
		dataFieldParser:        createDataFieldParserMock(),
		balanceConverter:       ap,
	}
}

func TestGetMoveBalanceTransaction(t *testing.T) {
	t.Parallel()

	txHash := []byte("txHash")
	mbHash := []byte("mbHash")
	mb := &block.MiniBlock{TxHashes: [][]byte{txHash}}
	header := &block.Header{Nonce: 2, TimeStamp: 1234}
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

	txInfo := &outport.TxInfo{
		Transaction: tx,
		FeeInfo: &outport.FeeInfo{
			GasUsed:        500,
			Fee:            big.NewInt(100),
			InitialPaidFee: big.NewInt(100),
		},
		ExecutionOrder: 0,
	}

	senderAddr, err := cp.addressPubkeyConverter.Encode(tx.RcvAddr)
	require.Nil(t, err)
	receiverAddr, err := cp.addressPubkeyConverter.Encode(tx.SndAddr)
	require.Nil(t, err)

	expectedTx := &data.Transaction{
		Hash:             hex.EncodeToString(txHash),
		MBHash:           hex.EncodeToString(mbHash),
		Nonce:            tx.Nonce,
		Round:            header.Round,
		Value:            tx.Value.String(),
		ValueNum:         1e-15,
		Receiver:         senderAddr,
		Sender:           receiverAddr,
		ReceiverShard:    mb.ReceiverShardID,
		SenderShard:      mb.SenderShardID,
		GasPrice:         gasPrice,
		GasLimit:         gasLimit,
		GasUsed:          uint64(500),
		InitialPaidFee:   "100",
		Data:             tx.Data,
		Signature:        hex.EncodeToString(tx.Signature),
		Status:           status,
		Fee:              "100",
		FeeNum:           1e-16,
		ReceiverUserName: []byte("rcv"),
		SenderUserName:   []byte("snd"),
		ESDTValuesNum:    []float64{},
		Operation:        "transfer",
		Version:          1,
		Receivers:        []string{},
		Timestamp:        1234,
		TimestampMs:      1234000,
	}

	dbTx := cp.prepareTransaction(txInfo, txHash, mbHash, mb, header, status, 3, 1234000)
	dbTx.UUID = ""
	require.Equal(t, expectedTx, dbTx)
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
	header := &block.Header{Nonce: 2, TimeStamp: 1234}
	status := "Success"

	rewardInfo := &outport.RewardInfo{
		Reward: rwdTx,
	}
	resultTx := cp.prepareRewardTransaction(rewardInfo, txHash, mbHash, mb, header, status, 1234000)
	resultTx.UUID = ""
	expectedTx := &data.Transaction{
		Hash:        hex.EncodeToString(txHash),
		MBHash:      hex.EncodeToString(mbHash),
		Round:       round,
		Receiver:    hex.EncodeToString(rcvAddr),
		Status:      status,
		Value:       "<nil>",
		Sender:      fmt.Sprintf("%d", core.MetachainShardId),
		Data:        make([]byte, 0),
		Operation:   rewardsOperation,
		Timestamp:   1234,
		TimestampMs: 1234000,
	}

	require.Equal(t, expectedTx, resultTx)
}

func TestRelayedV3Transaction(t *testing.T) {
	t.Parallel()

	txHash := []byte("txHash")
	mbHash := []byte("mbHash")
	mb := &block.MiniBlock{TxHashes: [][]byte{txHash}, Type: block.InvalidBlock}
	header := &block.Header{Nonce: 2, TimeStamp: 1234}
	status := transaction.TxStatusInvalid.String()
	gasPrice := uint64(1000)
	gasLimit := uint64(1000)
	cp := createCommonProcessor()

	tx := &transaction.Transaction{
		Nonce:            1,
		Value:            big.NewInt(1000),
		RcvAddr:          []byte("receiver"),
		SndAddr:          []byte("sender"),
		GasPrice:         gasPrice,
		GasLimit:         gasLimit,
		Data:             []byte("data"),
		ChainID:          []byte("1"),
		Version:          1,
		Signature:        []byte("signature"),
		RcvUserName:      []byte("rcv"),
		SndUserName:      []byte("snd"),
		RelayerAddr:      []byte("relayer"),
		RelayerSignature: []byte("relayerSignature"),
	}

	expectedTx := &data.Transaction{
		Hash:             hex.EncodeToString(txHash),
		MBHash:           hex.EncodeToString(mbHash),
		Nonce:            tx.Nonce,
		Round:            header.Round,
		Value:            tx.Value.String(),
		ValueNum:         1e-15,
		Receiver:         cp.addressPubkeyConverter.SilentEncode(tx.RcvAddr, log),
		Sender:           cp.addressPubkeyConverter.SilentEncode(tx.SndAddr, log),
		ReceiverShard:    uint32(2),
		SenderShard:      mb.SenderShardID,
		GasPrice:         gasPrice,
		GasLimit:         gasLimit,
		GasUsed:          uint64(500),
		InitialPaidFee:   "100",
		Data:             tx.Data,
		Signature:        hex.EncodeToString(tx.Signature),
		Status:           status,
		Fee:              "100",
		FeeNum:           1e-16,
		ReceiverUserName: []byte("rcv"),
		SenderUserName:   []byte("snd"),
		Operation:        "transfer",
		Version:          1,
		Receivers:        []string{},
		ESDTValuesNum:    []float64{},
		RelayedAddr:      hex.EncodeToString(tx.RelayerAddr),
		RelayedSignature: hex.EncodeToString(tx.RelayerSignature),
		Timestamp:        1234,
		TimestampMs:      1234000,
	}

	txInfo := &outport.TxInfo{
		Transaction: tx,
		FeeInfo: &outport.FeeInfo{
			GasUsed:        500,
			Fee:            big.NewInt(100),
			InitialPaidFee: big.NewInt(100),
		},
		ExecutionOrder: 0,
	}

	dbTx := cp.prepareTransaction(txInfo, txHash, mbHash, mb, header, status, 3, 1234000)
	dbTx.UUID = ""
	require.Equal(t, expectedTx, dbTx)
}

func TestGetMoveBalanceTransactionInvalid(t *testing.T) {
	t.Parallel()

	txHash := []byte("txHash")
	mbHash := []byte("mbHash")
	mb := &block.MiniBlock{TxHashes: [][]byte{txHash}, Type: block.InvalidBlock}
	header := &block.Header{Nonce: 2, TimeStamp: 1234}
	status := transaction.TxStatusInvalid.String()
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
		Hash:             hex.EncodeToString(txHash),
		MBHash:           hex.EncodeToString(mbHash),
		Nonce:            tx.Nonce,
		Round:            header.Round,
		Value:            tx.Value.String(),
		ValueNum:         1e-15,
		Receiver:         cp.addressPubkeyConverter.SilentEncode(tx.RcvAddr, log),
		Sender:           cp.addressPubkeyConverter.SilentEncode(tx.SndAddr, log),
		ReceiverShard:    uint32(2),
		SenderShard:      mb.SenderShardID,
		GasPrice:         gasPrice,
		GasLimit:         gasLimit,
		GasUsed:          uint64(500),
		InitialPaidFee:   "100",
		Data:             tx.Data,
		Signature:        hex.EncodeToString(tx.Signature),
		Status:           status,
		Fee:              "100",
		FeeNum:           1e-16,
		ReceiverUserName: []byte("rcv"),
		SenderUserName:   []byte("snd"),
		Operation:        "transfer",
		Version:          1,
		Receivers:        []string{},
		ESDTValuesNum:    []float64{},
		Timestamp:        1234,
		TimestampMs:      1234000,
	}

	txInfo := &outport.TxInfo{
		Transaction: tx,
		FeeInfo: &outport.FeeInfo{
			GasUsed:        500,
			Fee:            big.NewInt(100),
			InitialPaidFee: big.NewInt(100),
		},
		ExecutionOrder: 0,
	}

	dbTx := cp.prepareTransaction(txInfo, txHash, mbHash, mb, header, status, 3, 1234000)
	dbTx.UUID = ""
	require.Equal(t, expectedTx, dbTx)
}
