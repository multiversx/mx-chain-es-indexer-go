package mock

import (
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
)

const (
	minGasLimit      = uint64(50000)
	gasPerDataByte   = uint64(1500)
	gasPriceModifier = float64(0.01)
)

// EconomicsHandlerMock -
type EconomicsHandlerMock struct {
}

// MinGasLimit -
func (e *EconomicsHandlerMock) MinGasLimit() uint64 {
	return minGasLimit
}

// ComputeGasLimit -
func (e *EconomicsHandlerMock) ComputeGasLimit(tx coreData.TransactionWithFeeHandler) uint64 {
	gasLimit := minGasLimit

	dataLen := uint64(len(tx.GetData()))
	gasLimit += dataLen * gasPerDataByte

	return gasLimit
}

// ComputeGasUsedAndFeeBasedOnRefundValue -
func (e *EconomicsHandlerMock) ComputeGasUsedAndFeeBasedOnRefundValue(tx coreData.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int) {
	if refundValue.Cmp(big.NewInt(0)) == 0 {
		txFee := e.ComputeTxFee(tx)
		return tx.GetGasLimit(), txFee
	}

	txFee := e.ComputeTxFee(tx)

	txFee = big.NewInt(0).Sub(txFee, refundValue)

	moveBalanceGasUnits := e.ComputeGasLimit(tx)
	moveBalanceFee := e.computeMoveBalanceFee(tx)

	scOpFee := big.NewInt(0).Sub(txFee, moveBalanceFee)
	gasPriceForProcessing := big.NewInt(0).SetUint64(e.GasPriceForProcessing(tx))
	scOpGasUnits := big.NewInt(0).Div(scOpFee, gasPriceForProcessing)

	gasUsed := moveBalanceGasUnits + scOpGasUnits.Uint64()

	return gasUsed, txFee
}

// ComputeTxFeeBasedOnGasUsed -
func (e *EconomicsHandlerMock) ComputeTxFeeBasedOnGasUsed(tx coreData.TransactionWithFeeHandler, gasUsed uint64) *big.Int {
	moveBalanceGasLimit := e.ComputeGasLimit(tx)
	moveBalanceFee := e.computeMoveBalanceFee(tx)
	if gasUsed <= moveBalanceGasLimit {
		return moveBalanceFee
	}

	computeFeeForProcessing := e.ComputeFeeForProcessing(tx, gasUsed-moveBalanceGasLimit)
	txFee := big.NewInt(0).Add(moveBalanceFee, computeFeeForProcessing)

	return txFee
}

func (e *EconomicsHandlerMock) computeMoveBalanceFee(tx coreData.TransactionWithFeeHandler) *big.Int {
	return core.SafeMul(tx.GetGasPrice(), e.ComputeGasLimit(tx))
}

// IsInterfaceNil returns true if there is no value under the interface
func (e *EconomicsHandlerMock) IsInterfaceNil() bool {
	return e == nil
}

// ComputeFeeForProcessing will compute the fee using the gas price modifier, the gas to use and the actual gas price
func (e *EconomicsHandlerMock) ComputeFeeForProcessing(tx coreData.TransactionWithFeeHandler, gasToUse uint64) *big.Int {
	gasPrice := e.GasPriceForProcessing(tx)
	return core.SafeMul(gasPrice, gasToUse)
}

// GasPriceForProcessing computes the price for the gas in addition to balance movement and data
func (e *EconomicsHandlerMock) GasPriceForProcessing(tx coreData.TransactionWithFeeHandler) uint64 {
	return uint64(float64(tx.GetGasPrice()) * gasPriceModifier)
}

func (e *EconomicsHandlerMock) ComputeTxFee(tx coreData.TransactionWithFeeHandler) *big.Int {
	gasLimitForMoveBalance, difference := e.SplitTxGasInCategories(tx)
	moveBalanceFee := core.SafeMul(tx.GetGasPrice(), gasLimitForMoveBalance)
	if tx.GetGasLimit() <= gasLimitForMoveBalance {
		return moveBalanceFee
	}

	extraFee := e.ComputeFeeForProcessing(tx, difference)
	moveBalanceFee.Add(moveBalanceFee, extraFee)
	return moveBalanceFee
}

// SplitTxGasInCategories returns the gas split per categories
func (e *EconomicsHandlerMock) SplitTxGasInCategories(tx coreData.TransactionWithFeeHandler) (gasLimitMove, gasLimitProcess uint64) {
	var err error
	gasLimitMove = e.ComputeGasLimit(tx)
	gasLimitProcess, err = core.SafeSubUint64(tx.GetGasLimit(), gasLimitMove)
	if err != nil {
		fmt.Println("SplitTxGasInCategories - insufficient gas for move",
			"providedGas", tx.GetGasLimit(),
			"computedMinimumRequired", gasLimitMove,
			"dataLen", len(tx.GetData()),
		)
	}

	return
}