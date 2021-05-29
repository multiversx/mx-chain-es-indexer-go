package transactions

import (
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
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

func (etp *esdtTransactionProcessor) searchTxsWithNFTCreateAndPutNonceInAlteredAddress(
	alteredAccounts data.AlteredAccountsHandler,
	txs map[string]*data.Transaction,
	scrs []*data.ScResult,
) {
	for txHash, tx := range txs {
		encodedHash := hex.EncodeToString([]byte(txHash))
		etp.searchSCRWithNonceOfNFTAndPutInAlteredAddress(tx.Data, tx.EsdtTokenIdentifier, encodedHash, alteredAccounts, scrs)
	}
}

func (etp *esdtTransactionProcessor) searchSCRWithNonceOfNFTAndPutInAlteredAddress(
	dataField []byte,
	tokenIdentifier string,
	txHash string,
	alteredAccounts data.AlteredAccountsHandler,
	scrs []*data.ScResult,
) (found bool) {
	isBuiltinFunctionForNftCreate := strings.HasPrefix(string(dataField), core.BuiltInFunctionESDTNFTCreate)
	if !isBuiltinFunctionForNftCreate {
		return
	}

	for _, scr := range scrs {
		if scr.OriginalTxHash != txHash {
			continue
		}

		if scr.Receiver != scr.Sender {
			continue
		}

		nonceStr := etp.extractNonceString(scr)
		if nonceStr == 0 {
			continue
		}

		isNFT := etp.isNFTTx(dataField)
		isESDTNotNFT := !isNFT && etp.isESDTTx(dataField)
		alteredAccounts.Add(scr.Sender, &data.AlteredAccount{
			IsESDTOperation: isESDTNotNFT,
			IsNFTOperation:  isNFT,
			TokenIdentifier: tokenIdentifier,
			NFTNonce:        nonceStr,
		})
		return true
	}

	return
}

func (etp *esdtTransactionProcessor) extractNonceString(scr *data.ScResult) uint64 {
	scrDataSplit := etp.argumentParserExtended.split(string(scr.Data))
	if len(scrDataSplit) < 3 {
		return 0
	}

	if scrDataSplit[1] != okHexConst {
		return 0
	}

	return getNonce(scrDataSplit[2])
}

func getNonce(arg string) uint64 {
	nonceBytes, err := hex.DecodeString(arg)
	if err != nil {
		return 0
	}

	return big.NewInt(0).SetBytes(nonceBytes).Uint64()
}

func (etp *esdtTransactionProcessor) searchForESDTInScrs(alteredAccounts data.AlteredAccountsHandler, scrs []*data.ScResult) {
	for _, scr := range scrs {
		token, nonce, found := etp.processNftTransferData(scr.Data, alteredAccounts)
		if found {
			alteredAccounts.Add(scr.Sender, &data.AlteredAccount{
				IsNFTOperation:  true,
				TokenIdentifier: token,
				NFTNonce:        nonce,
			})
			continue
		}

		_ = etp.searchSCRWithNonceOfNFTAndPutInAlteredAddress(scr.Data, scr.EsdtTokenIdentifier, scr.OriginalTxHash, alteredAccounts, scrs)
	}
}

func (etp *esdtTransactionProcessor) searchForReceiverNFTTransferAndPutInAlteredAddress(
	txs map[string]*data.Transaction,
	alteredAccounts data.AlteredAccountsHandler,
) {
	for _, tx := range txs {
		_, _, _ = etp.processNftTransferData(tx.Data, alteredAccounts)
	}
}

func (etp *esdtTransactionProcessor) processNftTransferData(dataField []byte, alteredAccounts data.AlteredAccountsHandler) (token string, nonce uint64, found bool) {
	function, arguments, err := etp.argumentParserExtended.ParseData(string(dataField))
	if err != nil {
		return
	}

	if function != core.BuiltInFunctionESDTNFTTransfer {
		return
	}

	if len(arguments) < 4 {
		return
	}

	nonceUint := big.NewInt(0).SetBytes(arguments[1]).Uint64()

	receiverNFTTransfer := arguments[3]
	receiverShardID := etp.shardCoordinator.ComputeId(receiverNFTTransfer)
	if receiverShardID != etp.shardCoordinator.SelfId() {
		return
	}

	if len(receiverNFTTransfer) != etp.pubKeyConverter.Len() {
		return
	}

	encodedReceiverAddr := etp.pubKeyConverter.Encode(receiverNFTTransfer)
	alteredAccounts.Add(encodedReceiverAddr, &data.AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: string(arguments[0]),
		NFTNonce:        nonceUint,
	})

	return string(arguments[0]), nonceUint, true
}
