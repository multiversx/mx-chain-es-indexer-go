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

func (etp *esdtTransactionProcessor) searchForReceiverNFTTransferAndPutInAlteredAddress(
	txs map[string]*data.Transaction,
	alteredAddresses map[string]*data.AlteredAccount,
) {
	for _, tx := range txs {
		etp.nftTransferData(tx.Data, alteredAddresses)
	}
}

func (etp *esdtTransactionProcessor) nftTransferData(dataField []byte, alteredAddresses map[string]*data.AlteredAccount) {
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

	encodedReceiverAddr := etp.pubKeyConverter.Encode(receiverNFTTransfer)
	// TODO multiple ESDT operations same address next PR
	alteredAddresses[encodedReceiverAddr] = &data.AlteredAccount{
		IsSender:        false,
		IsESDTOperation: false,
		IsNFTOperation:  true,
		TokenIdentifier: string(arguments[0]),
		NFTNonceString:  nonceStr,
	}
}
