package datafield

import (
	"bytes"
	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elrond-go-core/core"
)

func (odp *operationDataFieldParser) parseESDTNFTTransfer(args [][]byte, sender, receiver []byte) *ResponseParseData {
	responseParse := &ResponseParseData{
		Operation: core.BuiltInFunctionESDTNFTTransfer,
	}

	parsedESDTTransfers, err := odp.esdtTransferParser.ParseESDTTransfers(sender, receiver, core.BuiltInFunctionESDTNFTTransfer, args)
	if err != nil {
		return responseParse
	}

	if len(parsedESDTTransfers.ESDTTransfers) == 0 || !isASCIIString(string(parsedESDTTransfers.ESDTTransfers[0].ESDTTokenName)) {
		return responseParse
	}

	if core.IsSmartContractAddress(parsedESDTTransfers.RcvAddr) && isASCIIString(parsedESDTTransfers.CallFunction) {
		responseParse.Function = parsedESDTTransfers.CallFunction
	}

	rcvAddr := receiver
	if bytes.Equal(sender, receiver) {
		rcvAddr = parsedESDTTransfers.RcvAddr
	}

	esdtNFTTransfer := parsedESDTTransfers.ESDTTransfers[0]
	receiverEncoded := odp.pubKeyConverter.Encode(rcvAddr)
	receiverShardID := odp.shardCoordinator.ComputeId(rcvAddr)
	token := converters.ComputeTokenIdentifier(string(esdtNFTTransfer.ESDTTokenName), esdtNFTTransfer.ESDTTokenNonce)

	responseParse.Tokens = append(responseParse.Tokens, token)
	responseParse.ESDTValues = append(responseParse.ESDTValues, esdtNFTTransfer.ESDTValue.String())
	responseParse.Receivers = append(responseParse.Receivers, receiverEncoded)
	responseParse.ReceiversShardID = append(responseParse.ReceiversShardID, receiverShardID)

	return responseParse
}
