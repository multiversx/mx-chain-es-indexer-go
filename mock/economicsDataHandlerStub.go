package mock

import (
	"math/big"

	coreData "github.com/ElrondNetwork/elrond-go-core/data"
)

// EconomicsHandlerStub -
type EconomicsHandlerStub struct {
	ComputeGasLimitCalled                        func(tx coreData.TransactionWithFeeHandler) uint64
	MinGasLimitCalled                            func() uint64
	ComputeGasUsedAndFeeBasedOnRefundValueCalled func(tx coreData.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int)
	ComputeTxFeeBasedOnGasUsedCalled             func(tx coreData.TransactionWithFeeHandler, gasUsed uint64) *big.Int
	ComputeMoveBalanceGasUsedCalled              func(tx coreData.TransactionWithFeeHandler) uint64
}

// MinGasLimit -
func (e *EconomicsHandlerStub) MinGasLimit() uint64 {
	if e.MinGasLimitCalled != nil {
		return e.MinGasLimitCalled()
	}
	return 0
}

// ComputeGasLimit -
func (e *EconomicsHandlerStub) ComputeGasLimit(tx coreData.TransactionWithFeeHandler) uint64 {
	if e.ComputeGasLimitCalled != nil {
		return e.ComputeGasLimitCalled(tx)
	}
	return 0
}

// ComputeGasUsedAndFeeBasedOnRefundValue -
func (e *EconomicsHandlerStub) ComputeGasUsedAndFeeBasedOnRefundValue(tx coreData.TransactionWithFeeHandler, refundValue *big.Int) (uint64, *big.Int) {
	if e.ComputeGasUsedAndFeeBasedOnRefundValueCalled != nil {
		return e.ComputeGasUsedAndFeeBasedOnRefundValueCalled(tx, refundValue)
	}

	return 0, nil
}

// ComputeTxFeeBasedOnGasUsed -
func (e *EconomicsHandlerStub) ComputeTxFeeBasedOnGasUsed(tx coreData.TransactionWithFeeHandler, gasUsed uint64) *big.Int {
	if e.ComputeTxFeeBasedOnGasUsedCalled != nil {
		return e.ComputeTxFeeBasedOnGasUsedCalled(tx, gasUsed)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (e *EconomicsHandlerStub) IsInterfaceNil() bool {
	return e == nil
}
