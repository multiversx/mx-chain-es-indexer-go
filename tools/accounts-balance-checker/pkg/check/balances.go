package check

type balancesESDT map[string]map[string]string

func newBalancesESDT() balancesESDT {
	return make(map[string]map[string]string)
}

func (be balancesESDT) extractBalancesFromResponse(responseAccounts *ResponseAccounts) {
	for _, hit := range responseAccounts.Hits.Hits {
		tokenIdentifier := hit.Source.TokenIdentifier
		if hit.Source.TokenIdentifier == "" {
			tokenIdentifier = hit.Source.TokenName
		}

		be.add(hit.Source.Address, tokenIdentifier, hit.Source.Balance)
	}
}

func (be balancesESDT) add(address, tokenIdentifier, value string) {
	_, ok := be[address]
	if !ok {
		be[address] = map[string]string{}
	}

	be[address][tokenIdentifier] = value
}

func (be balancesESDT) getBalancesForAddress(address string) map[string]string {
	return be[address]
}

func (be balancesESDT) getBalance(address, tokenIdentifier string) string {
	return be[address][tokenIdentifier]
}
