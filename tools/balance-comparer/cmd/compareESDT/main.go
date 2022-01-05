package main

import "github.com/ElrondNetwork/elastic-indexer-go/tools/balance-comparer/pkg/comparer"

const (
	esURL      = ""
	gatewayURL = ""
)

func main() {
	balancesComparer, err := comparer.NewBalanceComparerESDT(esURL, gatewayURL)
	if err != nil {
		panic("cannot create comparer " + err.Error())
	}

	err = balancesComparer.CompareBalances()
	if err != nil {
		panic("cannot compare balances " + err.Error())
	}
}
