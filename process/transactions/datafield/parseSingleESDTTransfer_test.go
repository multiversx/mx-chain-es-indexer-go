package datafield

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/stretchr/testify/require"
)

func TestParseESDTTransfer(t *testing.T) {
	t.Parallel()

	args := &ArgsOperationDataFieldParser{
		PubKeyConverter:  pubKeyConv,
		Marshalizer:      &mock.MarshalizerMock{},
		ShardCoordinator: &mock.ShardCoordinatorMock{},
	}

	parser, _ := NewOperationDataFieldParser(args)

	t.Run("TransferNonHexArguments", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("ESDTTransfer@1234@011")
		res := parser.Parse(dataField, sender, receiver)
		require.Equal(t, &ResponseParseData{
			Operation: operationTransfer,
		}, res)
	})

	t.Run("TransferNotEnoughtArguments", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("ESDTTransfer@1234")
		res := parser.Parse(dataField, sender, receiver)
		require.Equal(t, &ResponseParseData{
			Operation: "ESDTTransfer",
		}, res)
	})

	t.Run("TransferEmptyArguments", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("ESDTTransfer@544f4b454e@")
		res := parser.Parse(dataField, sender, receiver)
		require.Equal(t, &ResponseParseData{
			Operation:  "ESDTTransfer",
			Tokens:     []string{"TOKEN"},
			ESDTValues: []string{"0"},
		}, res)
	})

	t.Run("TransferWithSCCall", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("ESDTTransfer@544f4b454e@01@63616c6c4d65")
		res := parser.Parse(dataField, sender, receiverSC)
		require.Equal(t, &ResponseParseData{
			Operation:  "ESDTTransfer",
			Function:   "callMe",
			ESDTValues: []string{"1"},
			Tokens:     []string{"TOKEN"},
		}, res)
	})

	t.Run("TransferNonAsciStringToken", func(t *testing.T) {
		dataField := []byte("ESDTTransfer@055de6a779bbac0000@01")
		res := parser.Parse(dataField, sender, receiverSC)
		require.Equal(t, &ResponseParseData{
			Operation: "ESDTTransfer",
		}, res)
	})
}
