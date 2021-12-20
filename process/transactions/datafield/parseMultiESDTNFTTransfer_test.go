package datafield

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/stretchr/testify/require"
)

func TestMultiESDTNFTTransferParse(t *testing.T) {
	t.Parallel()

	args := &ArgsOperationDataFieldParser{
		PubKeyConverter:  pubKeyConv,
		Marshalizer:      &mock.MarshalizerMock{},
		ShardCoordinator: &mock.ShardCoordinatorMock{},
	}

	parser, _ := NewOperationDataFieldParser(args)

	t.Run("MultiNFTTransferWithSCCall", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("MultiESDTNFTTransfer@000000000000000005001e2a1428dd1e3a5146b3960d9e0f4a50369904ee5483@02@4c4b4d45582d616162393130@0d3d@058184103ad80ffb19f7@4c4b4641524d2d396431656138@1ecf06@0423fc01830d455ee5510c@656e7465724661726d416e644c6f636b5265776172647350726f7879@00000000000000000500656d0acc53561c5d6f6fd7d7e82bf13247014f615483")
		res := parser.Parse(dataField, sender, sender)
		require.Equal(t, &ResponseParseData{
			Operation:        "MultiESDTNFTTransfer",
			Function:         "enterFarmAndLockRewardsProxy",
			ESDTValues:       []string{"26000978570569047546359", "5005634793810936671326476"},
			Tokens:           []string{"LKMEX-aab910-0d3d", "LKFARM-9d1ea8-1ecf06"},
			Receivers:        []string{"erd1qqqqqqqqqqqqqpgqrc4pg2xarca9z34njcxeur622qmfjp8w2jps89fxnl", "erd1qqqqqqqqqqqqqpgqrc4pg2xarca9z34njcxeur622qmfjp8w2jps89fxnl"},
			ReceiversShardID: []uint32{0, 0},
		}, res)
	})

	t.Run("MultiNFTTransfer", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("MultiESDTNFTTransfer@000000000000000005001e2a1428dd1e3a5146b3960d9e0f4a50369904ee5483@02@4d4949552d61626364@00@01@4d4949552d616263646566@02@05")
		res := parser.Parse(dataField, sender, sender)
		require.Equal(t, &ResponseParseData{
			Operation:        "MultiESDTNFTTransfer",
			ESDTValues:       []string{"1", "5"},
			Tokens:           []string{"MIIU-abcd", "MIIU-abcdef-02"},
			Receivers:        []string{"erd1qqqqqqqqqqqqqpgqrc4pg2xarca9z34njcxeur622qmfjp8w2jps89fxnl", "erd1qqqqqqqqqqqqqpgqrc4pg2xarca9z34njcxeur622qmfjp8w2jps89fxnl"},
			ReceiversShardID: []uint32{0, 0},
		}, res)
	})

	t.Run("MultiNFTTransferNonHexArguments", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("MultiESDTNFTTransfer@000000000000000005001e2a1428dd1e3a5146b3960d9e0f4a50369904ee5483@02@4d4949552d61626364@00@01@4d4949552d616263646566@02@05@1")
		res := parser.Parse(dataField, sender, sender)
		require.Equal(t, &ResponseParseData{
			Operation: "transfer",
		}, res)
	})
	t.Run("MultiNFTTransferInvalidNumberOfArguments", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("MultiESDTNFTTransfer@000000000000000005001e2a1428dd1e3a5146b3960d9e0f4a50369904ee5483@02@4d4949552d61626364@00@01@4d4949552d616263646566@02")
		res := parser.Parse(dataField, sender, sender)
		require.Equal(t, &ResponseParseData{
			Operation: "MultiESDTNFTTransfer",
		}, res)
	})

	t.Run("MultiNFTTransferEmptyArguments", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("MultiESDTNFTTransfer@@@@@@@")
		res := parser.Parse(dataField, sender, sender)
		require.Equal(t, &ResponseParseData{
			Operation: "MultiESDTNFTTransfer",
		}, res)
	})
}
