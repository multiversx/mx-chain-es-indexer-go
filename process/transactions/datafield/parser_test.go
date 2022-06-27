package datafield

import (
	"encoding/hex"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"testing"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/stretchr/testify/require"
)

func createMockArgumentsOperationParser() *ArgsOperationDataFieldParser {
	return &ArgsOperationDataFieldParser{
		PubKeyConverter:  &mock.PubkeyConverterMock{},
		Marshalizer:      &mock.MarshalizerMock{},
		ShardCoordinator: &mock.ShardCoordinatorMock{},
	}
}

func TestNewOperationDataFieldParser(t *testing.T) {
	t.Parallel()

	t.Run("NilMarshalizer", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgumentsOperationParser()
		arguments.Marshalizer = nil

		_, err := NewOperationDataFieldParser(arguments)
		require.Equal(t, indexer.ErrNilMarshalizer, err)
	})

	t.Run("NilPubKeyConverter", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgumentsOperationParser()
		arguments.PubKeyConverter = nil

		_, err := NewOperationDataFieldParser(arguments)
		require.Equal(t, indexer.ErrNilPubkeyConverter, err)
	})

	t.Run("NilShardCoordinator", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgumentsOperationParser()
		arguments.ShardCoordinator = nil

		_, err := NewOperationDataFieldParser(arguments)
		require.Equal(t, indexer.ErrNilShardCoordinator, err)
	})

	t.Run("ShouldWork", func(t *testing.T) {
		t.Parallel()

		arguments := createMockArgumentsOperationParser()

		parser, err := NewOperationDataFieldParser(arguments)
		require.NotNil(t, parser)
		require.Nil(t, err)
	})
}

func TestParseSCDeploy(t *testing.T) {
	arguments := createMockArgumentsOperationParser()
	parser, _ := NewOperationDataFieldParser(arguments)

	t.Run("ScDeploy", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("0101020304050607")
		rcvAddr := make([]byte, 0)

		res := parser.Parse(dataField, sender, rcvAddr)
		require.Equal(t, &ResponseParseData{
			Operation: operationDeploy,
		}, res)
	})
}

func TestParseQuantityOperationsESDT(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsOperationParser()
	parser, _ := NewOperationDataFieldParser(arguments)

	t.Run("ESDTLocalBurn", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("ESDTLocalBurn@4d4949552d616263646566@0102")
		res := parser.Parse(dataField, sender, sender)
		require.Equal(t, &ResponseParseData{
			Operation:  "ESDTLocalBurn",
			ESDTValues: []string{"258"},
			Tokens:     []string{"MIIU-abcdef"},
		}, res)
	})

	t.Run("ESDTLocalMint", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("ESDTLocalMint@4d4949552d616263646566@1122")
		res := parser.Parse(dataField, sender, sender)
		require.Equal(t, &ResponseParseData{
			Operation:  "ESDTLocalMint",
			ESDTValues: []string{"4386"},
			Tokens:     []string{"MIIU-abcdef"},
		}, res)
	})

	t.Run("ESDTLocalMintNotEnoughArguments", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("ESDTLocalMint@4d4949552d616263646566")
		res := parser.Parse(dataField, sender, sender)
		require.Equal(t, &ResponseParseData{
			Operation: "ESDTLocalMint",
		}, res)
	})
}

func TestParseQuantityOperationsNFT(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsOperationParser()
	parser, _ := NewOperationDataFieldParser(arguments)

	t.Run("ESDTNFTCreate", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("ESDTNFTCreate@4d494841492d316630666638@01@54657374@03e8@516d664132487465726e674d6242655467506b3261327a6f4d357965616f33456f61373678513775346d63646947@746167733a746573742c667265652c66756e3b6d657461646174613a5468697320697320612074657374206465736372697074696f6e20666f7220616e20617765736f6d65206e6674@0101")
		res := parser.Parse(dataField, sender, sender)
		require.Equal(t, &ResponseParseData{
			Operation:  "ESDTNFTCreate",
			ESDTValues: []string{"1415934836"},
			Tokens:     []string{"MIHAI-1f0ff8-01"},
		}, res)
	})

	t.Run("ESDTNFTBurn", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("ESDTNFTBurn@54494b4954414b41@0102@123456")
		res := parser.Parse(dataField, sender, sender)
		require.Equal(t, &ResponseParseData{
			Operation:  "ESDTNFTBurn",
			ESDTValues: []string{"1193046"},
			Tokens:     []string{"TIKITAKA-0102"},
		}, res)
	})

	t.Run("ESDTNFTAddQuantity", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("ESDTNFTAddQuantity@54494b4954414b41@02@03")
		res := parser.Parse(dataField, sender, sender)
		require.Equal(t, &ResponseParseData{
			Operation:  "ESDTNFTAddQuantity",
			ESDTValues: []string{"3"},
			Tokens:     []string{"TIKITAKA-02"},
		}, res)
	})

	t.Run("ESDTNFTAddQuantityNotEnoughtArguments", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("ESDTNFTAddQuantity@54494b4954414b41@02")
		res := parser.Parse(dataField, sender, sender)
		require.Equal(t, &ResponseParseData{
			Operation: "ESDTNFTAddQuantity",
		}, res)
	})
}

func TestParseBlockingOperationESDT(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsOperationParser()
	parser, _ := NewOperationDataFieldParser(arguments)

	t.Run("ESDTFreeze", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("ESDTFreeze@54494b4954414b41")
		res := parser.Parse(dataField, sender, receiver)
		require.Equal(t, &ResponseParseData{
			Operation: "ESDTFreeze",
			Tokens:    []string{"TIKITAKA"},
		}, res)
	})

	t.Run("ESDTFreezeNFT", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("ESDTFreeze@544f4b454e2d616263642d3031")
		res := parser.Parse(dataField, sender, receiver)
		require.Equal(t, &ResponseParseData{
			Operation: "ESDTFreeze",
			Tokens:    []string{"TOKEN-abcd-01"},
		}, res)
	})

	t.Run("ESDTWipe", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("ESDTWipe@534b4537592d37336262636404")
		res := parser.Parse(dataField, sender, receiver)
		require.Equal(t, &ResponseParseData{
			Operation: "ESDTWipe",
			Tokens:    []string{"SKE7Y-73bbcd-04"},
		}, res)
	})

	t.Run("ESDTWipe", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("ESDTFreeze")
		res := parser.Parse(dataField, sender, receiver)
		require.Equal(t, &ResponseParseData{
			Operation: "ESDTFreeze",
		}, res)
	})

	t.Run("SCCall", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("callMe@01")
		res := parser.Parse(dataField, sender, receiverSC)
		require.Equal(t, &ResponseParseData{
			Operation: operationTransfer,
			Function:  "callMe",
		}, res)
	})

	t.Run("ESDTTransferRole", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("ESDTNFTCreateRoleTransfer@01010101@020202")
		res := parser.Parse(dataField, sender, receiver)
		require.Equal(t, &ResponseParseData{
			Operation: "ESDTNFTCreateRoleTransfer",
		}, res)
	})

}

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

	t.Run("RelayedTxV2Ok", func(t *testing.T) {
		t.Parallel()

		dataField := []byte(core.RelayedTransactionV2 +
			"@" +
			hex.EncodeToString(receiverSC) +
			"@" +
			"0A" +
			"@" +
			hex.EncodeToString([]byte("callMe@02")) +
			"@" +
			"01a2")

		res := parser.Parse(dataField, sender, receiver)
		require.Equal(t, &ResponseParseData{
			IsRelayed:        true,
			Operation:        "transfer",
			Function:         "callMe",
			Receivers:        []string{"erd1qqqqqqqqqqqqqpgqp699jngundfqw07d8jzkepucvpzush6k3wvqyc44rx"},
			ReceiversShardID: []uint32{0},
		}, res)
	})

	t.Run("RelayedTxV2NotEnoughtArgs", func(t *testing.T) {
		t.Parallel()

		dataField := []byte(core.RelayedTransactionV2 + "@abcd")
		res := parser.Parse(dataField, sender, receiver)
		require.Equal(t, &ResponseParseData{
			IsRelayed: true,
		}, res)
	})

	t.Run("RelayedTxV1NoArguments", func(t *testing.T) {
		t.Parallel()

		dataField := []byte(core.RelayedTransaction)
		res := parser.Parse(dataField, sender, receiver)
		require.Equal(t, &ResponseParseData{
			IsRelayed: true,
		}, res)
	})

	t.Run("RelayedTxV2WithRelayedTxIn", func(t *testing.T) {
		t.Parallel()

		dataField := []byte(core.RelayedTransactionV2 +
			"@" +
			hex.EncodeToString(receiverSC) +
			"@" +
			"0A" +
			"@" +
			hex.EncodeToString([]byte(core.RelayedTransaction)) +
			"@" +
			"01a2")
		res := parser.Parse(dataField, sender, receiver)
		require.Equal(t, &ResponseParseData{
			IsRelayed: true,
		}, res)
	})

	t.Run("RelayedTxV2WithNFTTransfer", func(t *testing.T) {
		t.Parallel()

		nftTransferData := []byte("ESDTNFTTransfer@4c4b4641524d2d396431656138@34ae14@728faa2c8883760aaf53bb@000000000000000005001e2a1428dd1e3a5146b3960d9e0f4a50369904ee5483@636c61696d5265776172647350726f7879@00000000000000000500a655b2b534218d6d8cfa1f219960be2f462e92565483")
		dataField := []byte(core.RelayedTransactionV2 +
			"@" +
			hex.EncodeToString(receiver) +
			"@" +
			"0A" +
			"@" +
			hex.EncodeToString(nftTransferData) +
			"@" +
			"01a2")
		res := parser.Parse(dataField, sender, receiver)
		require.Equal(t, &ResponseParseData{
			IsRelayed:        true,
			Operation:        "ESDTNFTTransfer",
			ESDTValues:       []string{"138495980998569893315957691"},
			Tokens:           []string{"LKFARM-9d1ea8-34ae14"},
			Receivers:        []string{"erd1qqqqqqqqqqqqqpgqrc4pg2xarca9z34njcxeur622qmfjp8w2jps89fxnl"},
			ReceiversShardID: []uint32{0},
			Function:         "claimRewardsProxy",
		}, res)
	})
}
