package transactions

import (
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/parsers"
)

const (
	builtInFuncWipeSingleNFT = "wipeSingleNFT"
	emptyString              = ""
)

type esdtTransactionProcessor struct {
	esdtOperations         map[string]struct{}
	esdtNFTOperations      map[string]struct{}
	argumentParserExtended *argumentsParserExtended
}

func newEsdtTransactionHandler() *esdtTransactionProcessor {
	argsParser := parsers.NewCallArgsParser()
	esdtTxProc := &esdtTransactionProcessor{
		argumentParserExtended: newArgumentsParser(argsParser),
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

func (etp *esdtTransactionProcessor) getNFTTxInfo(txData []byte) (tokenIdentifier, nonce string) {
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
		nonce = big.NewInt(0).SetBytes(arguments[1]).String()
	}

	return
}

func (etp *esdtTransactionProcessor) searchTxsWithNFTCreateAndPutNonceInAlteredAddress(
	alteredAddresses map[string]*data.AlteredAccount,
	txs map[string]*data.Transaction,
	scrs []*data.ScResult,
) {
	for txHash, tx := range txs {
		isBuiltinFunctionForNftCreate := strings.HasPrefix(string(tx.Data), core.BuiltInFunctionESDTNFTCreate)
		if !isBuiltinFunctionForNftCreate {
			continue
		}

		encodedHash := hex.EncodeToString([]byte(txHash))
		etp.searchSCRWithNonceOfNFT(encodedHash, alteredAddresses, scrs)
	}
}

func (etp *esdtTransactionProcessor) searchSCRSWithCreateNFTAndPutNonceInAlteredAddress(
	alteredAddresses map[string]*data.AlteredAccount,
	scrs []*data.ScResult,
) {
	for _, scr := range scrs {
		if !strings.HasPrefix(string(scr.Data), core.BuiltInFunctionESDTNFTCreate) {
			continue
		}

		etp.searchSCRWithNonceOfNFT(scr.OriginalTxHash, alteredAddresses, scrs)
	}
}

func (etp *esdtTransactionProcessor) searchSCRWithNonceOfNFT(txHash string, alteredAddresses map[string]*data.AlteredAccount, scrs []*data.ScResult) {
	for _, scr := range scrs {
		if scr.OriginalTxHash != txHash {
			continue
		}

		if scr.Receiver != scr.Sender {
			continue
		}

		altered, ok := alteredAddresses[scr.Sender]
		if !ok {
			return
		}

		altered.NFTNonceString = etp.extractNonceString(scr)
	}
}

func (etp *esdtTransactionProcessor) extractNonceString(scr *data.ScResult) string {
	scrDataSplit := etp.argumentParserExtended.split(string(scr.Data))
	if len(scrDataSplit) < 3 {
		return ""
	}

	if scrDataSplit[1] != okHexConst {
		return ""
	}

	nonceBytes, err := hex.DecodeString(scrDataSplit[2])
	if err != nil {
		return ""
	}

	return big.NewInt(0).SetBytes(nonceBytes).String()
}
