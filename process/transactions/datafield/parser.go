package datafield

import (
	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm-common/parsers"
	"math/big"
)

const OperationTransfer = `transfer`

type operationDataFieldParser struct {
	argsParser         vmcommon.CallArgsParser
	pubKeyConverter    core.PubkeyConverter
	shardCoordinator   indexer.ShardCoordinator
	esdtTransferParser vmcommon.ESDTTransferParser
}

// NewOperationDataFieldParser will return a new instance of operationDataFieldParser
func NewOperationDataFieldParser(args *ArgsOperationDataFieldParser) (*operationDataFieldParser, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, indexer.ErrNilShardCoordinator
	}
	if check.IfNil(args.PubKeyConverter) {
		return nil, indexer.ErrNilPubkeyConverter
	}
	if check.IfNil(args.Marshalizer) {
		return nil, indexer.ErrNilMarshalizer
	}

	argsParser := parsers.NewCallArgsParser()
	esdtTransferParser, err := parsers.NewESDTTransferParser(args.Marshalizer)
	if err != nil {
		return nil, err
	}

	return &operationDataFieldParser{
		argsParser:         argsParser,
		pubKeyConverter:    args.PubKeyConverter,
		shardCoordinator:   args.ShardCoordinator,
		esdtTransferParser: esdtTransferParser,
	}, nil
}

// Parse will parse the provided data field
func (odp *operationDataFieldParser) Parse(dataField []byte, sender, receiver []byte) *ResponseParseData {
	function, args, err := odp.argsParser.ParseData(string(dataField))
	if err != nil {
		return &ResponseParseData{
			Operation: OperationTransfer,
		}
	}

	switch function {
	case core.BuiltInFunctionESDTTransfer:
		return odp.parseESDTTransfer(args, sender, receiver)
	case core.BuiltInFunctionESDTNFTTransfer:
		return odp.parseESDTNFTTransfer(args, sender, receiver)
	case core.BuiltInFunctionMultiESDTNFTTransfer:
		return odp.parseMultiESDTNFTTransfer(args, sender, receiver)
	case core.BuiltInFunctionESDTLocalBurn, core.BuiltInFunctionESDTLocalMint:
		return parseQuantityOperationESDT(args, function)
	case core.BuiltInFunctionESDTWipe, core.BuiltInFunctionESDTFreeze, core.BuiltInFunctionESDTUnFreeze:
		return parseBlockingOperationESDT(args, function)
	case core.BuiltInFunctionESDTNFTCreate, core.BuiltInFunctionESDTNFTBurn, core.BuiltInFunctionESDTNFTAddQuantity:
		return parseQuantityOperationNFT(args, function)
	default:
		return &ResponseParseData{
			Operation: OperationTransfer,
			Function:  function,
		}
	}
}

func parseBlockingOperationESDT(args [][]byte, funcName string) *ResponseParseData {
	responseData := &ResponseParseData{
		Operation: funcName,
	}

	if len(args) == 0 {
		return responseData
	}

	responseData.Tokens = append(responseData.Tokens, string(args[0]))
	return responseData
}

func parseQuantityOperationESDT(args [][]byte, funcName string) *ResponseParseData {
	responseData := &ResponseParseData{
		Operation: funcName,
	}

	if len(args) < 2 {
		return responseData
	}

	responseData.Tokens = append(responseData.Tokens, string(args[0]))
	responseData.ESDTValues = append(responseData.ESDTValues, big.NewInt(0).SetBytes(args[1]).String())

	return responseData
}

func parseQuantityOperationNFT(args [][]byte, funcName string) *ResponseParseData {
	responseData := &ResponseParseData{
		Operation: funcName,
	}

	if len(args) < 3 {
		return responseData
	}

	nonce := big.NewInt(0).SetBytes(args[1]).Uint64()
	token := converters.ComputeTokenIdentifier(string(args[0]), nonce)
	responseData.Tokens = append(responseData.Tokens, token)

	value := big.NewInt(0).SetBytes(args[2]).String()
	responseData.ESDTValues = append(responseData.ESDTValues, value)

	return responseData
}