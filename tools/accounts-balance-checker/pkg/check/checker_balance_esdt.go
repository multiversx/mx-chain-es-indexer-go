package check

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"time"
)

const (
	allTokensEndpoint    = "/address/%s/esdt"
	specificESDTEndpoint = allTokensEndpoint + "/%s"
	specificNFTEndpoint  = "/address/%s/nft/%s/nonce/%d"
)

func (bc *balanceChecker) CheckESDTBalances() error {
	balancesFromEs, err := bc.getAllESDTAccounts()
	if err != nil {
		return err
	}

	log.Info("total accounts with ESDT tokens ", "count", len(balancesFromEs))

	for addr, tokenBalanceMap := range balancesFromEs {
		decoded, errD := bc.pubKeyConverter.Decode(addr)
		if errD != nil {
			log.Warn("cannot decode address", "address", addr, "error", err)
			continue
		}

		if core.IsSmartContractAddress(decoded) {
			// TODO treat sc
			continue
		}

		balancesFromProxy, errP := bc.getBalancesFromProxy(addr)
		if errP != nil {
			log.Warn("cannot get balances from proxy", "address", addr, "error", err)
		}

		bc.compareBalances(tokenBalanceMap, balancesFromProxy, addr)
	}

	return nil
}

func (bc *balanceChecker) compareBalances(balancesFromES, balancesFromProxy map[string]string, address string) {
	for tokenIdentifier, balanceES := range balancesFromES {
		balanceProxy, ok := balancesFromProxy[tokenIdentifier]
		if !ok {
			log.Warn("extra balance in ES", "address", address, "token identifier", tokenIdentifier)
			continue
		}

		delete(balancesFromProxy, tokenIdentifier)

		if balanceES != balanceProxy {
			log.Warn("different balance", "address", address, "token identifier", tokenIdentifier,
				"balance from ES", balanceES, "balance from proxy", balanceProxy,
			)
			continue
		}
	}

	for tokenIdentifier, balance := range balancesFromProxy {
		log.Warn("missing balance from ES", "address", address,
			"token identifier", tokenIdentifier, "balance", balance,
		)
	}
}

func (bc *balanceChecker) getBalancesFromProxy(address string) (map[string]string, error) {
	responseBalancesProxy := &BalancesESDTResponse{}
	err := bc.restClient.CallGetRestEndPoint(fmt.Sprintf(allTokensEndpoint, address), responseBalancesProxy)
	if err != nil {
		return nil, err
	}
	if responseBalancesProxy.Error != "" {
		return nil, errors.New(responseBalancesProxy.Error)
	}

	balances := make(map[string]string)
	for tokenIdentifier, tokenData := range responseBalancesProxy.Data {
		balances[tokenIdentifier] = tokenData.Balance
	}

	return balances, nil
}

func (bc *balanceChecker) getAllESDTAccounts() (balancesESDT, error) {
	defer logExecutionTime(time.Now(), "balanceChecker.getAllESDTAccounts")

	balances := newBalancesESDT()

	countAccountsESDT := 0
	handlerFunc := func(responseBytes []byte) error {
		accountsRes := &ResponseAccounts{}
		err := json.Unmarshal(responseBytes, accountsRes)
		if err != nil {
			return err
		}

		balances.extractBalancesFromResponse(accountsRes)

		countAccountsESDT++
		log.Info("read accounts balance from es", "count", countAccountsESDT)

		return nil
	}

	err := bc.esClient.DoScrollRequestAllDocuments(
		accountsIndex,
		[]byte(matchAllQuery),
		handlerFunc,
	)
	if err != nil {
		return nil, err
	}

	return balances, nil
}

func (bc *balanceChecker) handlerFuncScrollAccountESDT(responseBytes []byte) error {
	accountsRes := &ResponseAccounts{}
	err := json.Unmarshal(responseBytes, accountsRes)
	if err != nil {
		return err
	}

	return nil
}
