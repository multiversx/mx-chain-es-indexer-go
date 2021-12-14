package datafield

import (
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
		arguments := createMockArgumentsOperationParser()
		arguments.Marshalizer = nil

		_, err := NewOperationDataFieldParser(arguments)
		require.Equal(t, indexer.ErrNilMarshalizer, err)
	})

	t.Run("NilPubKeyConverter", func(t *testing.T) {
		arguments := createMockArgumentsOperationParser()
		arguments.PubKeyConverter = nil

		_, err := NewOperationDataFieldParser(arguments)
		require.Equal(t, indexer.ErrNilPubkeyConverter, err)
	})

	t.Run("NilShardCoordinator", func(t *testing.T) {
		arguments := createMockArgumentsOperationParser()
		arguments.ShardCoordinator = nil

		_, err := NewOperationDataFieldParser(arguments)
		require.Equal(t, indexer.ErrNilShardCoordinator, err)
	})

	t.Run("ShouldWork", func(t *testing.T) {
		arguments := createMockArgumentsOperationParser()

		parser, err := NewOperationDataFieldParser(arguments)
		require.NotNil(t, parser)
		require.Nil(t, err)
	})
}

func TestParseQuantityOperationsESDT(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsOperationParser()
	parser, _ := NewOperationDataFieldParser(arguments)

	t.Run("ESDTLocalBurn", func(t *testing.T) {
		dataField := []byte("ESDTLocalBurn@4d4949552d616263646566@0102")
		res := parser.Parse(dataField, sender, sender)
		require.Equal(t, &ResponseParseData{
			Operation:  "ESDTLocalBurn",
			ESDTValues: []string{"258"},
			Tokens:     []string{"MIIU-abcdef"},
		}, res)
	})

	t.Run("ESDTLocalMint", func(t *testing.T) {
		dataField := []byte("ESDTLocalMint@4d4949552d616263646566@1122")
		res := parser.Parse(dataField, sender, sender)
		require.Equal(t, &ResponseParseData{
			Operation:  "ESDTLocalMint",
			ESDTValues: []string{"4386"},
			Tokens:     []string{"MIIU-abcdef"},
		}, res)
	})

	t.Run("ESDTLocalMintNotEnoughArguments", func(t *testing.T) {
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
		dataField := []byte("ESDTNFTCreate@4d494841492d316630666638@01@54657374@03e8@516d664132487465726e674d6242655467506b3261327a6f4d357965616f33456f61373678513775346d63646947@746167733a746573742c667265652c66756e3b6d657461646174613a5468697320697320612074657374206465736372697074696f6e20666f7220616e20617765736f6d65206e6674@0101")
		res := parser.Parse(dataField, sender, sender)
		require.Equal(t, &ResponseParseData{
			Operation:  "ESDTNFTCreate",
			ESDTValues: []string{"1415934836"},
			Tokens:     []string{"MIHAI-1f0ff8-01"},
		}, res)
	})

	t.Run("ESDTNFTBurn", func(t *testing.T) {
		dataField := []byte("ESDTNFTBurn@54494b4954414b41@0102@123456")
		res := parser.Parse(dataField, sender, sender)
		require.Equal(t, &ResponseParseData{
			Operation:  "ESDTNFTBurn",
			ESDTValues: []string{"1193046"},
			Tokens:     []string{"TIKITAKA-0102"},
		}, res)
	})

	t.Run("ESDTNFTAddQuantity", func(t *testing.T) {
		dataField := []byte("ESDTNFTAddQuantity@54494b4954414b41@02@03")
		res := parser.Parse(dataField, sender, sender)
		require.Equal(t, &ResponseParseData{
			Operation:  "ESDTNFTAddQuantity",
			ESDTValues: []string{"3"},
			Tokens:     []string{"TIKITAKA-02"},
		}, res)
	})

	t.Run("ESDTNFTAddQuantityNotEnoughtArguments", func(t *testing.T) {
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
		dataField := []byte("ESDTFreeze@54494b4954414b41")
		res := parser.Parse(dataField, sender, receiver)
		require.Equal(t, &ResponseParseData{
			Operation: "ESDTFreeze",
			Tokens:    []string{"TIKITAKA"},
		}, res)
	})

	t.Run("ESDTFreezeNFT", func(t *testing.T) {
		dataField := []byte("ESDTFreeze@544f4b454e2d616263642d3031")
		res := parser.Parse(dataField, sender, receiver)
		require.Equal(t, &ResponseParseData{
			Operation: "ESDTFreeze",
			Tokens:    []string{"TOKEN-abcd-01"},
		}, res)
	})

	t.Run("ESDTWipe", func(t *testing.T) {
		dataField := []byte("ESDTWipe@534b4537592d37336262636404")
		res := parser.Parse(dataField, sender, receiver)
		require.Equal(t, &ResponseParseData{
			Operation: "ESDTWipe",
			Tokens:    []string{"SKE7Y-73bbcd-04"},
		}, res)
	})

	t.Run("ESDTWipe", func(t *testing.T) {
		dataField := []byte("ESDTFreeze")
		res := parser.Parse(dataField, sender, receiver)
		require.Equal(t, &ResponseParseData{
			Operation: "ESDTFreeze",
		}, res)
	})

	t.Run("SCCall", func(t *testing.T) {
		dataField := []byte("callMe@01")
		res := parser.Parse(dataField, sender, receiverSC)
		require.Equal(t, &ResponseParseData{
			Operation: operationTransfer,
			Function:  "callMe",
		}, res)
	})
}

func TestRelayedOperation(t *testing.T) {
	t.Parallel()

	arguments := createMockArgumentsOperationParser()
	parser, _ := NewOperationDataFieldParser(arguments)

	t.Run("RelayedTxV1", func(t *testing.T) {
		dataField := []byte("relayedTx@7b226e6f6e6365223a322c2276616c7565223a302c227265636569766572223a22414141414141414141414146414974673738352f736c73554148686b57334569624c6e47524b76496f4e4d3d222c2273656e646572223a22726b6e534a477a343769534e794b43642f504f717075776b5477684534306d7a476a585a51686e622b724d3d222c226761735072696365223a313030303030303030302c226761734c696d6974223a31353030303030302c2264617461223a22633246325a5546306447567a644746306157397551444668597a49314d6a5935596d51335a44497759324a6959544d31596d566c4f4459314d4464684f574e6a4e7a677a5a4755774f445a694e4445334e546b345a54517a59544e6b5a6a566a593245795a5468684d6a6c414d6a51344e54677a4d574e6d4d5445304d54566d596a41354d6a63774e4451324e5755324e7a597a59574d314f4445345a5467314e4751345957526d4e54417a596a63354d6a6c6b4f54526c4e6d49794e6a49775a673d3d222c22636861696e4944223a224d513d3d222c2276657273696f6e223a312c227369676e6174757265223a225239462b34546352415a386d7771324559303163596c337662716c46657176387a76474a775a6833594d4f556b4234643451574e66376744626c484832576b71614a76614845744356617049713365356562384e41773d3d227d")
		res := parser.Parse(dataField, sender, receiverSC)
		require.Equal(t, &ResponseParseData{
			Operation: operationRelayedTx,
		}, res)
	})
}
