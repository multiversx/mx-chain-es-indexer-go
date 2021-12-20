package datafield

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
)

func (odp *operationDataFieldParser) parseESDTTransfer(args [][]byte, sender, receiver []byte) *ResponseParseData {
	responseParse := &ResponseParseData{
		Operation: core.BuiltInFunctionESDTTransfer,
	}

	parsedESDTTransfers, err := odp.esdtTransferParser.ParseESDTTransfers(sender, receiver, core.BuiltInFunctionESDTTransfer, args)
	if err != nil {
		return responseParse
	}

	if core.IsSmartContractAddress(receiver) {
		responseParse.Function = parsedESDTTransfers.CallFunction
	}

	if len(parsedESDTTransfers.ESDTTransfers) == 0 {
		return responseParse
	}
	responseParse.Tokens = append(responseParse.Tokens, string(parsedESDTTransfers.ESDTTransfers[0].ESDTTokenName))
	responseParse.ESDTValues = append(responseParse.ESDTValues, parsedESDTTransfers.ESDTTransfers[0].ESDTValue.String())

	return responseParse
}
