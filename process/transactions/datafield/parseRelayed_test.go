package datafield

import (
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOperationDataFieldParser_ParseRelayed(t *testing.T) {
	t.Parallel()

	args := &ArgsOperationDataFieldParser{
		PubKeyConverter:  pubKeyConv,
		Marshalizer:      &mock.MarshalizerMock{},
		ShardCoordinator: &mock.ShardCoordinatorMock{},
	}

	parser, _ := NewOperationDataFieldParser(args)

	t.Run("RelayedTxOk", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("relayedTx@7b226e6f6e6365223a362c2276616c7565223a302c227265636569766572223a2241414141414141414141414641436e626331733351534939726e6d697a69684d7a3631665539446a71786b3d222c2273656e646572223a2248714b386459464a43474144346a756d4e4e742b314530745a6579736376714c7a38624c47574e774177453d222c226761735072696365223a313030303030303030302c226761734c696d6974223a31353030303030302c2264617461223a2252564e45564652795957357a5a6d56795144517a4e446330597a51304d6d517a4f544d794d7a677a4e444d354d7a4a414d444e6c4f4541324d6a63314e7a6b304d7a59344e6a55334d7a6330514745774d4441774d444177222c22636861696e4944223a2252413d3d222c2276657273696f6e223a312c227369676e6174757265223a2262367331755349396f6d4b63514448344337624f534a632f62343166577a3961584d777334526966552b71343870486d315430636f72744b727443484a4258724f67536b3651333254546f7a6e4e2b7074324f4644413d3d227d")

		res := parser.Parse(dataField, sender, receiver)
		require.Equal(t, &ResponseParseData{
			IsRelayed:        true,
			Operation:        "ESDTTransfer",
			Function:         "buyChest",
			Tokens:           []string{"CGLD-928492"},
			ESDTValues:       []string{"1000"},
			Receivers:        []string{"erd1qqqqqqqqqqqqqpgq98dhxkehgy3rmtne5t8zsnx04404858r4vvsamdlsv"},
			ReceiversShardID: []uint32{0},
		}, res)
	})
}
