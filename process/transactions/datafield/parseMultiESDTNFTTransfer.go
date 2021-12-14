package datafield

import (
	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elrond-go-core/core"
)

func (odp *operationDataFieldParser) parseMultiESDTNFTTransfer(args [][]byte, sender, receiver []byte) *ResponseParseData {
	responseParse := &ResponseParseData{
		Operation: core.BuiltInFunctionMultiESDTNFTTransfer,
	}

	parsedESDTTransfers, err := odp.esdtTransferParser.ParseESDTTransfers(sender, receiver, core.BuiltInFunctionMultiESDTNFTTransfer, args)
	if err != nil {
		return responseParse
	}

	if core.IsSmartContractAddress(parsedESDTTransfers.RcvAddr) {
		responseParse.Function = parsedESDTTransfers.CallFunction
	}

	receiverEncoded := odp.pubKeyConverter.Encode(parsedESDTTransfers.RcvAddr)
	receiverShardID := odp.shardCoordinator.ComputeId(parsedESDTTransfers.RcvAddr)

	for _, esdtTransferData := range parsedESDTTransfers.ESDTTransfers {
		token := string(esdtTransferData.ESDTTokenName)
		if esdtTransferData.ESDTTokenNonce != 0 {
			token = converters.ComputeTokenIdentifier(string(esdtTransferData.ESDTTokenName), esdtTransferData.ESDTTokenNonce)
		}

		responseParse.Tokens = append(responseParse.Tokens, token)
		responseParse.ESDTValues = append(responseParse.ESDTValues, esdtTransferData.ESDTValue.String())
		responseParse.Receivers = append(responseParse.Receivers, receiverEncoded)
		responseParse.ReceiversShardID = append(responseParse.ReceiversShardID, receiverShardID)
	}

	return responseParse
}
