package transactions

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
	datafield "github.com/multiversx/mx-chain-vm-common-go/parsers/dataField"
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
	ap, _ := converters.NewBalanceConverter(18)
	scrsProc := newSmartContractResultsProcessor(pubKeyConverter, &mock.MarshalizerMock{}, &mock.HasherMock{}, parser, ap)

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

	scrInfo := &outport.SCRInfo{
		SmartContractResult: smartContractRes,
		FeeInfo: &outport.FeeInfo{
			Fee: big.NewInt(0),
		},
	}

	header := &block.Header{TimeStamp: 100}

	mbHash := []byte("hash")
	scRes := scrsProc.prepareSmartContractResult(scHash, mbHash, scrInfo, header, 0, 1, 3)

	senderAddr, err := pubKeyConverter.Encode(sndAddr)
	require.Nil(t, err)
	receiverAddr, err := pubKeyConverter.Encode(rcvAddr)
	require.Nil(t, err)

	expectedTx := &data.ScResult{
		Nonce:              nonce,
		Hash:               scHash,
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
		ESDTValuesNum:      []float64{},
		InitialTxFee:       "0",
	}

	require.Equal(t, expectedTx, scRes)
}
