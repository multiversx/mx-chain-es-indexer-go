package transactions

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/parsers"
	"github.com/ElrondNetwork/elrond-go/process"
)

type esdtTransactionProcessor struct {
	esdtOperations    map[string]struct{}
	esdtNFTOperations map[string]struct{}
	argumentParser    process.CallArgumentsParser
}

func newEsdtTransactionHandler() *esdtTransactionProcessor {
	esdtTxProc := &esdtTransactionProcessor{
		argumentParser: parsers.NewCallArgsParser(),
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
		"wipeSingleNFT":                        {},
	}
}

func (etp *esdtTransactionProcessor) getFunctionName(txData []byte) string {
	function, _, err := etp.argumentParser.ParseData(string(txData))
	if err != nil {
		return ""
	}

	return function
}

func (etp *esdtTransactionProcessor) isESDTTx(txData []byte) bool {
	functionName := etp.getFunctionName(txData)
	if functionName == "" {
		return false
	}

	_, esdtFunc := etp.esdtOperations[functionName]
	_, nftFunc := etp.esdtNFTOperations[functionName]

	return esdtFunc || nftFunc
}

func (etp *esdtTransactionProcessor) isNFTTx(txData []byte) bool {
	functionName := etp.getFunctionName(txData)
	if functionName == "" {
		return false
	}

	_, ok := etp.esdtNFTOperations[functionName]
	return ok
}

func (etp *esdtTransactionProcessor) getTokenIdentifier(txData []byte) string {
	_, arguments, err := etp.argumentParser.ParseData(string(txData))
	if err != nil {
		return ""
	}

	if len(arguments) >= 1 {
		return string(arguments[0])
	}

	return ""
}

func (etp *esdtTransactionProcessor) getNFTTxInfo(txData []byte) (tokenIdentifier, nonce string) {
	function, arguments, err := etp.argumentParser.ParseData(string(txData))
	if err != nil {
		return
	}

	if len(arguments) < 1 {
		return
	}

	_, nftFunc := etp.esdtNFTOperations[function]
	if !nftFunc {
		return
	}

	tokenIdentifier = string(arguments[0])

	if len(arguments) < 2 {
		return
	}

	switch function {
	case core.BuiltInFunctionESDTNFTTransfer, "wipeSingleNFT", core.BuiltInFunctionESDTNFTAddQuantity:
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
		if !strings.HasPrefix(string(tx.Data), core.BuiltInFunctionESDTNFTCreate) {
			continue
		}

		encodedHash := hex.EncodeToString([]byte(txHash))
		searchSCRWithNonceOfNFT(encodedHash, alteredAddresses, scrs)
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

		searchSCRWithNonceOfNFT(scr.OriginalTxHash, alteredAddresses, scrs)
	}
}

func searchSCRWithNonceOfNFT(txHash string, alteredAddresses map[string]*data.AlteredAccount, scrs []*data.ScResult) {
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

		altered.NFTNonceString = extractNonceString(scr)
	}
}

func extractNonceString(scr *data.ScResult) string {
	scrDataSplit := bytes.Split(scr.Data, []byte("@"))
	if len(scrDataSplit) < 3 {
		return ""
	}

	if !bytes.Equal(scrDataSplit[1], []byte("6f6b")) {
		return ""
	}

	nonceBytes, err := hex.DecodeString(string(scrDataSplit[2]))
	if err != nil {
		return ""
	}

	return big.NewInt(0).SetBytes(nonceBytes).String()
}
