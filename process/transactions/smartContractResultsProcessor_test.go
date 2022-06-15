package transactions

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-vm-common/parsers/dataField"
	"github.com/stretchr/testify/require"
)

func createDataFieldParserMock() DataFieldParser {
	args := &datafield.ArgsOperationDataFieldParser{
		Marshalizer:      &mock.MarshalizerMock{},
		ShardCoordinator: &mock.ShardCoordinatorMock{},
		AddressLength:    32,
	}
	parser, _ := datafield.NewOperationDataFieldParser(args)

	return parser
}

func TestPrepareSmartContractResult(t *testing.T) {
	t.Parallel()

	parser := createDataFieldParserMock()
	pubKeyConverter := &mock.PubkeyConverterMock{}
	scrsProc := newSmartContractResultsProcessor(pubKeyConverter, &mock.ShardCoordinatorMock{}, &mock.MarshalizerMock{}, &mock.HasherMock{}, parser)

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
		CallType:   1,
	}
	header := &block.Header{TimeStamp: 100}

	mbHash := []byte("hash")
	scRes := scrsProc.prepareSmartContractResult([]byte(scHash), mbHash, smartContractRes, header, 0, 1)
	expectedTx := &data.ScResult{
		Nonce:              nonce,
		Hash:               hex.EncodeToString([]byte(scHash)),
		PrevTxHash:         hex.EncodeToString(txHash),
		MBHash:             hex.EncodeToString(mbHash),
		Code:               string(code),
		Data:               make([]byte, 0),
		Sender:             pubKeyConverter.Encode(sndAddr),
		Receiver:           pubKeyConverter.Encode(rcvAddr),
		Value:              "<nil>",
		CallType:           "1",
		Timestamp:          time.Duration(100),
		SenderShard:        0,
		ReceiverShard:      1,
		Operation:          "transfer",
		SenderAddressBytes: sndAddr,
		Receivers:          []string{},
	}

	require.Equal(t, expectedTx, scRes)
}

func TestAddScrsReceiverToAlteredAccounts_ShouldWork(t *testing.T) {
	t.Parallel()

	parser := createDataFieldParserMock()
	scrsProc := newSmartContractResultsProcessor(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{}, &mock.MarshalizerMock{}, &mock.HasherMock{}, parser)

	alteredAddress := data.NewAlteredAccounts()
	scrs := []*data.ScResult{
		{
			Sender:   "010101",
			Receiver: "020202",
			Data:     []byte("ESDTTransfer@544b4e2d626231323061@010f0cf064dd59200000"),
			Value:    "1",
		},
	}
	scrsProc.addScrsReceiverToAlteredAccounts(alteredAddress, scrs)
	require.Equal(t, 1, alteredAddress.Len())

	_, ok := alteredAddress.Get("020202")
	require.True(t, ok)
}
