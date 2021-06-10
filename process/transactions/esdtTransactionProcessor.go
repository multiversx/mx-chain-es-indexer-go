package transactions

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/parsers"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const (
	builtInFuncWipeSingleNFT = "wipeSingleNFT"
	emptyString              = ""
)

type esdtTransactionProcessor struct {
	esdtOperations         map[string]struct{}
	esdtNFTOperations      map[string]struct{}
	argumentParserExtended *argumentsParserExtended
	pubKeyConverter        core.PubkeyConverter
	shardCoordinator       sharding.Coordinator
}

func newEsdtTransactionHandler(
	pubKeyConverter core.PubkeyConverter,
	shardCoordinator sharding.Coordinator,
) *esdtTransactionProcessor {
	argsParser := parsers.NewCallArgsParser()
	esdtTxProc := &esdtTransactionProcessor{
		argumentParserExtended: newArgumentsParser(argsParser),
		pubKeyConverter:        pubKeyConverter,
		shardCoordinator:       shardCoordinator,
	}

	esdtTxProc.initESDTOperations()
	esdtTxProc.initESDTNFTOperations()

	return esdtTxProc
}

func (etp *esdtTransactionProcessor) initESDTOperations() {
	etp.esdtOperations = map[string]struct{}{
		core.BuiltInFunctionESDTTransfer: {},
		core.BuiltInFunctionESDTBurn:     {},
		core.BuiltInFunctionESDTFreeze:   {},
		core.BuiltInFunctionESDTUnFreeze: {},
		core.BuiltInFunctionESDTWipe:     {},
		core.BuiltInFunctionESDTPause:    {},
		core.BuiltInFunctionESDTUnPause:  {},
	}
}

func (etp *esdtTransactionProcessor) initESDTNFTOperations() {
	etp.esdtNFTOperations = map[string]struct{}{
		core.BuiltInFunctionESDTNFTCreate:      {},
		core.BuiltInFunctionESDTNFTTransfer:    {},
		core.BuiltInFunctionESDTNFTAddQuantity: {},
		builtInFuncWipeSingleNFT:               {},
	}
}

func (etp *esdtTransactionProcessor) getFunctionName(txData []byte) string {
	function, _, err := etp.argumentParserExtended.ParseData(string(txData))
	if err != nil {
		return emptyString
	}

	return function
}

func (etp *esdtTransactionProcessor) isESDTTx(txData []byte) bool {
	functionName := etp.getFunctionName(txData)
	if functionName == emptyString {
		return false
	}

	_, isEsdtFunc := etp.esdtOperations[functionName]
	_, isNftFunc := etp.esdtNFTOperations[functionName]

	return isEsdtFunc || isNftFunc
}

func (etp *esdtTransactionProcessor) isNFTTx(txData []byte) bool {
	functionName := etp.getFunctionName(txData)
	if functionName == emptyString {
		return false
	}

	_, ok := etp.esdtNFTOperations[functionName]
	return ok
}

func (etp *esdtTransactionProcessor) getTokenIdentifier(txData []byte) string {
	_, arguments, err := etp.argumentParserExtended.ParseData(string(txData))
	if err != nil {
		return emptyString
	}

	if len(arguments) < 1 {
		return emptyString
	}

	return string(arguments[0])
}

func (etp *esdtTransactionProcessor) getNFTTxInfo(txData []byte) (tokenIdentifier string, nonce uint64) {
	function, arguments, err := etp.argumentParserExtended.ParseData(string(txData))
	if err != nil {
		return
	}

	if len(arguments) < 1 {
		return
	}

	_, isNftFunc := etp.esdtNFTOperations[function]
	if !isNftFunc {
		return
	}

	tokenIdentifier = string(arguments[0])

	if len(arguments) < 2 {
		return
	}

	switch function {
	case core.BuiltInFunctionESDTNFTTransfer, builtInFuncWipeSingleNFT, core.BuiltInFunctionESDTNFTAddQuantity:
		nonce = big.NewInt(0).SetBytes(arguments[1]).Uint64()
	}

	return
}
