package comparer

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

type balancesESDTResponse struct {
	Data struct {
		ESDTS map[string]struct {
			Balance string `json:"balance"`
		} `json:"esdts"`
	} `json:"data"`
	Error string `json:"error"`
}

type responseAccountsESDT struct {
	Hits struct {
		Hits []struct {
			ID     string            `json:"_id"`
			Source *data.AccountInfo `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type balancesKeepEsdt struct {
	balances map[string]map[string]string
}

func newBalancesKeepESDT() *balancesKeepEsdt {
	return &balancesKeepEsdt{
		balances: map[string]map[string]string{},
	}
}

func (bke *balancesKeepEsdt) add(address, token, balance string) {
	_, ok := bke.balances[address]
	if !ok {
		bke.balances[address] = map[string]string{}
	}

	bke.balances[address][token] = balance
}

func (bke *balancesKeepEsdt) addMultiple(address string, balances map[string]string) {
	bke.balances[address] = balances
}

func (bke *balancesKeepEsdt) getAllAddresses() []string {
	addresses := make([]string, 0, len(bke.balances))
	for addr := range bke.balances {
		addresses = append(addresses, addr)
	}

	return addresses
}

func (bke *balancesKeepEsdt) getBalances(addr string) map[string]string {
	return bke.balances[addr]
}

func (bke *balancesKeepEsdt) compare(altBke *balancesKeepEsdt) {
	for addr, balancesESDT := range bke.balances {
		altBalances, ok := altBke.balances[addr]
		if !ok {
			log.Warn("cannot find balances", "address", addr)
		}

		compareAddressBalances(addr, balancesESDT, altBalances)
	}
}

func compareAddressBalances(sourceAddr string, sourceBalance map[string]string, dstBalances map[string]string) {
	if len(sourceBalance) == 0 {
		log.Warn("no balances on source", "address", sourceAddr)
	}

	for token, value := range sourceBalance {
		altBalance, okB := dstBalances[token]
		if !okB {
			log.Warn("cannot find balances", "address", sourceAddr, "token", token)
		}

		if altBalance != value {
			log.Warn("different balances",
				"address", sourceAddr, "token", token,
				"balance source", value,
				"balance destination", altBalance,
			)
		}
	}
}
