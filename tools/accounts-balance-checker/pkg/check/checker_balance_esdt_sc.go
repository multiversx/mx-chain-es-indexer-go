package check

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

func (bc *balanceChecker) checkBalancesSC(addr string, balancesFromES map[string]string) {
	tokenBalanceProxy := make(map[string]string)
	for tokenIdentifier := range balancesFromES {
		endpoint := computeEndpoint(tokenIdentifier, addr)

		responseBalancesProxy := &BalancesESDTResponse{}
		err := bc.restClient.CallGetRestEndPoint(endpoint, responseBalancesProxy)
		if err != nil {
			log.Warn("cannot call get rest endpoint", "error", err)
		}
		if responseBalancesProxy.Error != "" {
			log.Warn("error in response from gateway", "error", responseBalancesProxy.Error)
		}

		balanceProxy := responseBalancesProxy.Data.TokenData.Balance
		tokenBalanceProxy[tokenIdentifier] = balanceProxy
	}

	tryAgain := bc.compareBalances(balancesFromES, tokenBalanceProxy, addr, true)
	if tryAgain {
		err := bc.getFromESAndCompose(addr, tokenBalanceProxy, len(balancesFromES))
		if err != nil {
			log.Warn("cannot compare second time", "address", addr, "error", err)
		}
	}
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
