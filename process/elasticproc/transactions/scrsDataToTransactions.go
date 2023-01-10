package transactions

import (
	"encoding/hex"
	"strings"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

const (
	minNumOfArgumentsNFTTransferORMultiTransfer = 4
)

type scrsDataToTransactions struct {
	retCodes []string
}

func newScrsDataToTransactions() *scrsDataToTransactions {
	return &scrsDataToTransactions{
		retCodes: []string{
			vmcommon.FunctionNotFound.String(),
			vmcommon.FunctionWrongSignature.String(),
			vmcommon.ContractNotFound.String(),
			vmcommon.UserError.String(),
			vmcommon.OutOfGas.String(),
			vmcommon.AccountCollision.String(),
			vmcommon.OutOfFunds.String(),
			vmcommon.CallStackOverFlow.String(),
			vmcommon.ContractInvalid.String(),
			vmcommon.ExecutionFailed.String(),
			vmcommon.UpgradeFailed.String(),
		},
	}
}

func (st *scrsDataToTransactions) attachSCRsToTransactionsAndReturnSCRsWithoutTx(txs map[string]*data.Transaction, scrs []*data.ScResult) []*data.ScResult {
	scrsWithoutTx := make([]*data.ScResult, 0)
	for _, scr := range scrs {
		decodedOriginalTxHash, err := hex.DecodeString(scr.OriginalTxHash)
		if err != nil {
			continue
		}

		tx, ok := txs[string(decodedOriginalTxHash)]
		if !ok {
			scrsWithoutTx = append(scrsWithoutTx, scr)
			continue
		}

		tx.SmartContractResults = append(tx.SmartContractResults, scr)
	}

	return scrsWithoutTx
}

func (st *scrsDataToTransactions) processTransactionsAfterSCRsWereAttached(transactions map[string]*data.Transaction) {
	for _, tx := range transactions {
		if len(tx.SmartContractResults) == 0 {
			continue
		}

		st.fillTxWithSCRsFields(tx)
	}
}

func (st *scrsDataToTransactions) fillTxWithSCRsFields(tx *data.Transaction) {
	tx.HasSCR = true

	if isRelayedTx(tx) {
		return
	}

	// ignore invalid transaction because status and gas fields were already set
	if tx.Status == transaction.TxStatusInvalid.String() {
		return
	}

	if hasSuccessfulSCRs(tx) {
		return
	}

	if hasCrossShardPendingTransfer(tx) {
		return
	}

	if st.hasSCRWithErrorCode(tx) {
		tx.Status = transaction.TxStatusFail.String()
	}
}

func (st *scrsDataToTransactions) hasSCRWithErrorCode(tx *data.Transaction) bool {
	for _, scr := range tx.SmartContractResults {
		for _, codeStr := range st.retCodes {
			if strings.Contains(string(scr.Data), hex.EncodeToString([]byte(codeStr))) ||
				scr.ReturnMessage == codeStr {
				return true
			}
		}
	}

	return false
}

func hasSuccessfulSCRs(tx *data.Transaction) bool {
	for _, scr := range tx.SmartContractResults {
		if isScResultSuccessful(scr.Data) {
			return true
		}
	}

	return false
}

func hasCrossShardPendingTransfer(tx *data.Transaction) bool {
	for _, scr := range tx.SmartContractResults {
		splitData := strings.Split(string(scr.Data), data.AtSeparator)
		if len(splitData) < 2 {
			continue
		}

		isMultiTransferOrNFTTransfer := splitData[0] == core.BuiltInFunctionESDTNFTTransfer || splitData[0] == core.BuiltInFunctionMultiESDTNFTTransfer
		if !isMultiTransferOrNFTTransfer {
			continue
		}

		if scr.SenderShard != scr.ReceiverShard {
			return true
		}
	}

	return false
}

func (st *scrsDataToTransactions) processSCRsWithoutTx(scrs []*data.ScResult) (map[string]string, map[string]*data.FeeData) {
	txHashStatus := make(map[string]string)
	txHashRefund := make(map[string]*data.FeeData)
	for _, scr := range scrs {
		if scr.InitialTxGasUsed != 0 {
			txHashRefund[scr.OriginalTxHash] = &data.FeeData{
				Fee:      scr.InitialTxFee,
				GasUsed:  scr.InitialTxGasUsed,
				Receiver: scr.Receiver,
			}
		}

		if !st.isESDTNFTTransferOrMultiTransferWithError(string(scr.Data)) {
			continue
		}

		txHashStatus[scr.OriginalTxHash] = transaction.TxStatusFail.String()
	}

	return txHashStatus, txHashRefund
}

func (st *scrsDataToTransactions) isESDTNFTTransferOrMultiTransferWithError(scrData string) bool {
	splitData := strings.Split(scrData, data.AtSeparator)
	isMultiTransferOrNFTTransfer := splitData[0] == core.BuiltInFunctionESDTNFTTransfer || splitData[0] == core.BuiltInFunctionMultiESDTNFTTransfer
	if !isMultiTransferOrNFTTransfer || len(splitData) < minNumOfArgumentsNFTTransferORMultiTransfer {
		return false
	}

	latestArgumentFromDataField := splitData[len(splitData)-1]
	for _, retCode := range st.retCodes {
		isWithError := latestArgumentFromDataField == hex.EncodeToString([]byte(retCode))
		if isWithError {
			return true
		}
	}

	return false
}
