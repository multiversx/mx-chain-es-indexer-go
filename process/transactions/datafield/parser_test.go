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
			Operation: OperationTransfer,
			Function:  "callMe",
		}, res)
	})
}
