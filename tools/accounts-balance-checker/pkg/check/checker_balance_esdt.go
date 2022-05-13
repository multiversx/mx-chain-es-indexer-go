package check

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/tools/accounts-balance-checker/pkg/utils"
	"github.com/ElrondNetwork/elrond-go-core/core"
)

const (
	accountsesdtIndex = "accountsesdt"

	allTokensEndpoint    = "/address/%s/esdt"
	specificESDTEndpoint = allTokensEndpoint + "/%s"
	specificNFTEndpoint  = "/address/%s/nft/%s/nonce/%d"
)

var countTotalCompared uint64 = 0

func (bc *balanceChecker) CheckESDTBalances() error {
	balancesFromEs, err := bc.getAllESDTAccountsFromFile()
	if err != nil {
		return err
	}

	log.Info("total accounts with ESDT tokens ", "count", len(balancesFromEs))

	maxGoroutines := maxNumberOfRequestsInParallel
	done := make(chan struct{}, maxGoroutines)
	for addr, tokenBalanceMap := range balancesFromEs {
		done <- struct{}{}
		atomic.AddUint64(&countTotalCompared, 1)
		go bc.compareBalancesFromES(addr, tokenBalanceMap, done)
	}

	return nil
}

func (bc *balanceChecker) compareBalancesFromES(addr string, tokenBalanceMap map[string]string, done chan struct{}) {
	defer func() {
		<-done
	}()

	decoded, errD := bc.pubKeyConverter.Decode(addr)
	if errD != nil {
		log.Warn("cannot decode address", "address", addr, "error", errD)
		return
	}

	if core.IsSmartContractAddress(decoded) {
		// TODO treat sc
		return
	}

	balancesFromProxy, errP := bc.getBalancesFromProxy(addr)
	if errP != nil {
		log.Warn("cannot get balances from proxy", "address", addr, "error", errP)
	}

	tryAgain := bc.compareBalances(tokenBalanceMap, balancesFromProxy, addr, true)
	if tryAgain {
		err := bc.getFromESAndCompose(addr, balancesFromProxy)
		if err != nil {
			log.Warn("cannot compare second time", "address", addr, "error", err)
		}
		return
	}
}

func (bc *balanceChecker) getFromESAndCompose(address string, balancesFromProxy map[string]string) error {
	log.Info("second compare", "address", address, "total compared till now", atomic.LoadUint64(&countTotalCompared))

	encoded, _ := encodeQuery(getBalancesByAddress(address))
	accountsResponse := &ResponseAccounts{}
	err := bc.esClient.DoGetRequest(&encoded, accountsesdtIndex, accountsResponse, 9999)
	if err != nil {
		return err
	}

	balancesES := newBalancesESDT()
	balancesES.extractBalancesFromResponse(accountsResponse)

	_ = bc.compareBalances(balancesES.getBalancesForAddress(address), balancesFromProxy, address, false)

	return nil
}

func (bc *balanceChecker) compareBalances(balancesFromES, balancesFromProxy map[string]string, address string, firstCompare bool) (tryAgain bool) {
	copyBalancesProxy := make(map[string]string)
	for k, v := range balancesFromProxy {
		copyBalancesProxy[k] = v
	}

	for tokenIdentifier, balanceES := range balancesFromES {
		balanceProxy, ok := copyBalancesProxy[tokenIdentifier]
		if !ok && firstCompare {
			return true
		}

		if !ok {
			log.Warn("extra balance in ES", "address", address, "token identifier", tokenIdentifier)
			continue
		}

		delete(copyBalancesProxy, tokenIdentifier)

		if balanceES != balanceProxy && firstCompare {
			return true
		}

		if balanceES != balanceProxy {
			log.Warn("different balance", "address", address, "token identifier", tokenIdentifier,
				"balance from ES", balanceES, "balance from proxy", balanceProxy,
			)
			continue
		}
	}

	if len(copyBalancesProxy) > 0 && firstCompare {
		return true
	}

	for tokenIdentifier, balance := range copyBalancesProxy {
		log.Warn("missing balance from ES", "address", address,
			"token identifier", tokenIdentifier, "balance", balance,
		)
	}

	return false
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
	for tokenIdentifier, tokenData := range responseBalancesProxy.Data.ESDTS {
		balances[tokenIdentifier] = tokenData.Balance
	}

	return balances, nil
}

func (bc *balanceChecker) getAllESDTAccounts() (balancesESDT, error) {
	defer utils.LogExecutionTime(log, time.Now(), "get all accounts with ESDT tokens from ES")

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
		accountsesdtIndex,
		[]byte(matchAllQuery),
		handlerFunc,
	)
	if err != nil {
		return nil, err
	}

	return balances, nil
}

// TODO delete this after testing is done
func (bc *balanceChecker) getAllESDTAccountsFromFile() (balancesESDT, error) {
	balances := newBalancesESDT()
	jsonFile, _ := os.Open("./accounts-esdt.json")
	byteValue, _ := ioutil.ReadAll(jsonFile)
	_ = json.Unmarshal(byteValue, &balances)

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
