package transactions

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	datafield "github.com/ElrondNetwork/elrond-vm-common/parsers/dataField"
	"github.com/stretchr/testify/require"
)

func createDataFieldParserMock() DataFieldParser {
	args := &datafield.ArgsOperationDataFieldParser{
		AddressLength: 32,
		Marshalizer:   &mock.MarshalizerMock{},
	}
	parser, _ := datafield.NewOperationDataFieldParser(args)

	return parser
}

func TestPrepareSmartContractResult(t *testing.T) {
	t.Parallel()

	parser := createDataFieldParserMock()
	pubKeyConverter := &mock.PubkeyConverterMock{}
	scrsProc := newSmartContractResultsProcessor(pubKeyConverter, &mock.MarshalizerMock{}, &mock.HasherMock{}, parser)

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
	scRes := scrsProc.prepareSmartContractResult([]byte(scHash), mbHash, smartContractRes, header, 0, 1, big.NewInt(0), 0, 3)
	senderAddr, err := pubKeyConverter.Encode(sndAddr)
	require.Nil(t, err)
	receiverAddr, err := pubKeyConverter.Encode(rcvAddr)
	require.Nil(t, err)
	expectedTx := &data.ScResult{
		Nonce:              nonce,
		Hash:               hex.EncodeToString([]byte(scHash)),
		PrevTxHash:         hex.EncodeToString(txHash),
		MBHash:             hex.EncodeToString(mbHash),
		Code:               string(code),
		Data:               make([]byte, 0),
		Sender:             senderAddr,
		Receiver:           receiverAddr,
		Value:              "<nil>",
		CallType:           "1",
		Timestamp:          time.Duration(100),
		SenderShard:        0,
		ReceiverShard:      1,
		Operation:          "transfer",
		SenderAddressBytes: sndAddr,
		Receivers:          []string{},
		InitialTxFee:       "0",
	}

	require.Equal(t, expectedTx, scRes)
}
