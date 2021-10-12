package converters

import "math/big"

// BalanceConverter defines what a balance converter should be able to do
type BalanceConverter interface {
	ComputeBalanceAsFloat(balance *big.Int) float64
	ComputeESDTBalanceAsFloat(balance *big.Int) float64
	IsInterfaceNil() bool
}
