package check

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
)

func (bc *balanceChecker) checkBalancesSC(addr string, balancesFromES map[string]string) {
	tokenBalanceProxy := make(map[string]string)

	var (
		wg    = &sync.WaitGroup{}
		done  = make(chan struct{}, bc.maxNumberOfParallelRequests)
		mutex = &sync.Mutex{}
	)

	for tokenIdentifier := range balancesFromES {
		done <- struct{}{}
		wg.Add(1)

		go func(identifier string) {
			defer func() {
				<-done
				wg.Done()
			}()

			balanceProxy, ok := bc.getBalanceFromProxy(computeEndpoint(identifier, addr))
			if !ok {
				return
			}

			mutex.Lock()
			tokenBalanceProxy[identifier] = balanceProxy
			mutex.Unlock()

		}(tokenIdentifier)
	}

	wg.Wait()

	tryAgain := bc.compareBalances(balancesFromES, tokenBalanceProxy, addr, true)
	if tryAgain {
		err := bc.getFromESAndCompare(addr, tokenBalanceProxy, len(balancesFromES))
		if err != nil {
			log.Warn("cannot compare second time", "address", addr, "error", err)
		}
	}
}

func (bc *balanceChecker) getBalanceFromProxy(endpoint string) (string, bool) {
	responseBalancesProxy := &BalancesESDTResponse{}
	err := bc.restClient.CallGetRestEndPoint(endpoint, responseBalancesProxy)
	if err != nil {
		log.Warn("cannot call get rest endpoint", "error", err)
		return "", false
	}
	if responseBalancesProxy.Error != "" {
		log.Warn("error in response from gateway", "error", responseBalancesProxy.Error)
		return "", false
	}

	if responseBalancesProxy.Data.TokenData.Balance == "0" {
		return "", false
	}

	return responseBalancesProxy.Data.TokenData.Balance, true
}

func computeEndpoint(tokenIdentifier string, addr string) string {
	split := strings.Split(tokenIdentifier, "-")

	endpoint := fmt.Sprintf(specificESDTEndpoint, addr, tokenIdentifier)
	if len(split) == 3 {
		token := split[0] + "-" + split[1]
		nonceStr := split[2]
		decodedNonce, _ := hex.DecodeString(nonceStr)
		nonceUint := big.NewInt(0).SetBytes(decodedNonce).Uint64()

		endpoint = fmt.Sprintf(specificNFTEndpoint, addr, token, nonceUint)
	}

	return endpoint
}
