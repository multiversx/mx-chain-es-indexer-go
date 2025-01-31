package transactions

import (
	"encoding/hex"
	"math/big"

	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
)

const (
	minNumOfArgumentsNFTTransferORMultiTransfer = 4
)

type scrsDataToTransactions struct {
	balanceConverter dataindexer.BalanceConverter
}

func newScrsDataToTransactions(balanceConverter dataindexer.BalanceConverter) *scrsDataToTransactions {
	return &scrsDataToTransactions{
		balanceConverter: balanceConverter,
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
		tx.HasSCR = true
	}
}

func (st *scrsDataToTransactions) processSCRsWithoutTx(scrs []*data.ScResult) map[string]*data.FeeData {
	txHashRefund := make(map[string]*data.FeeData)
	for _, scr := range scrs {
		if (scr.InitialTxGasUsed == 0 || scr.OriginalTxHash == "") && scr.GasRefunded == 0 {
			continue
		}

		var feeNum float64
		var err error
		initialTxFeeBig, ok := big.NewInt(0).SetString(scr.InitialTxFee, 10)
		if ok {
			feeNum, err = st.balanceConverter.ConvertBigValueToFloat(initialTxFeeBig)
		}
		if err != nil {
			log.Warn("scrsDataToTransactions.processSCRsWithoutTx: cannot compute fee as num",
				"initial Tx fee", initialTxFeeBig, "error", err)
		}

		txHashRefund[scr.OriginalTxHash] = &data.FeeData{
			FeeNum:      feeNum,
			Fee:         scr.InitialTxFee,
			GasUsed:     scr.InitialTxGasUsed,
			Receiver:    scr.Receiver,
			GasRefunded: scr.GasRefunded,
		}
	}

	return txHashRefund
}
