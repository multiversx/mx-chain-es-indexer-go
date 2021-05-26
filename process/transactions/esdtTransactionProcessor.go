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
		encodedHash := hex.EncodeToString([]byte(txHash))
		etp.searchSCRWithNonceOfNFTAndPutInAlteredAddress(tx.Data, tx.EsdtTokenIdentifier, encodedHash, alteredAddresses, scrs)
	}
}

func (etp *esdtTransactionProcessor) searchSCRWithNonceOfNFTAndPutInAlteredAddress(
	dataField []byte,
	tokenIdentifier string,
	txHash string,
	alteredAddresses map[string]*data.AlteredAccount,
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
		if nonceStr == emptyString {
			continue
		}

		isNFT := etp.isNFTTx(dataField)
		isESDTNotNFT := !isNFT && etp.isESDTTx(dataField)
		alteredAddresses[scr.Sender] = &data.AlteredAccount{
			IsESDTOperation: isESDTNotNFT,
			IsNFTOperation:  isNFT,
			TokenIdentifier: tokenIdentifier,
			NFTNonceString:  nonceStr,
		}
		return true
	}

	return
}

func (etp *esdtTransactionProcessor) extractNonceString(scr *data.ScResult) string {
	scrDataSplit := etp.argumentParserExtended.split(string(scr.Data))
	if len(scrDataSplit) < 3 {
		return emptyString
	}

	if scrDataSplit[1] != okHexConst {
		return emptyString
	}

	return getNonceString(scrDataSplit[2])
}

func getNonceString(arg string) string {
	nonceBytes, err := hex.DecodeString(arg)
	if err != nil {
		return emptyString
	}

	return big.NewInt(0).SetBytes(nonceBytes).String()
}

func (etp *esdtTransactionProcessor) searchForESDTInScrs(alteredAddresses map[string]*data.AlteredAccount, scrs []*data.ScResult) {
	for _, scr := range scrs {
		token, nonce, found := etp.nftTransferData(scr.Data, alteredAddresses)
		if found {
			alteredAddresses[scr.Sender] = &data.AlteredAccount{
				IsNFTOperation:  true,
				TokenIdentifier: token,
				NFTNonceString:  nonce,
			}
			continue
		}

		found = etp.searchSCRWithNonceOfNFTAndPutInAlteredAddress(scr.Data, scr.EsdtTokenIdentifier, scr.OriginalTxHash, alteredAddresses, scrs)
		if found {
			continue
		}

	}
}

func (etp *esdtTransactionProcessor) searchForReceiverNFTTransferAndPutInAlteredAddress(
	txs map[string]*data.Transaction,
	alteredAddresses map[string]*data.AlteredAccount,
) {
	for _, tx := range txs {
		_, _, _ = etp.nftTransferData(tx.Data, alteredAddresses)
	}
}

func (etp *esdtTransactionProcessor) nftTransferData(dataField []byte, alteredAddresses map[string]*data.AlteredAccount) (token, nonce string, found bool) {
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

	nonceStr := big.NewInt(0).SetBytes(arguments[1]).String()

	receiverNFTTransfer := arguments[3]
	receiverShardID := etp.shardCoordinator.ComputeId(receiverNFTTransfer)
	if receiverShardID != etp.shardCoordinator.SelfId() {
		return
	}

	if len(receiverNFTTransfer) != etp.pubKeyConverter.Len() {
		return
	}

	encodedReceiverAddr := etp.pubKeyConverter.Encode(receiverNFTTransfer)
	// TODO multiple ESDT operations same address next PR
	alteredAddresses[encodedReceiverAddr] = &data.AlteredAccount{
		IsNFTOperation:  true,
		TokenIdentifier: string(arguments[0]),
		NFTNonceString:  nonceStr,
	}

	return string(arguments[0]), nonceStr, true
}
