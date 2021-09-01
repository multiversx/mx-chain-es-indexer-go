package logsevents

import (
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

type pendingBalancesProc struct {
	pendingBalances map[string]*data.AccountInfo
}

func newPendingBalancesProcessor() *pendingBalancesProc {
	return &pendingBalancesProc{
		pendingBalances: make(map[string]*data.AccountInfo),
	}
}

func (pbp *pendingBalancesProc) addInfo(receiver string, token string, tokenNonce uint64, value string) {
	hexEncodedNonce := converters.EncodeNonceToHex(tokenNonce)
	key := fmt.Sprintf("%s-%s-%s-%s", pendingBalanceIdentifier, receiver, token, hexEncodedNonce)

	_, found := pbp.pendingBalances[key]
	if !found {
		pbp.pendingBalances[key] = &data.AccountInfo{
			Address:         fmt.Sprintf("%s-%s", pendingBalanceIdentifier, receiver),
			Balance:         value,
			TokenName:       token,
			TokenIdentifier: converters.ComputeTokenIdentifier(token, tokenNonce),
			TokenNonce:      tokenNonce,
		}
		return
	}

	balanceBigMap, ok := big.NewInt(0).SetString(pbp.pendingBalances[key].Balance, 10)
	if !ok {
		pbp.pendingBalances[key].Balance = value
	}

	valueBig, ok := big.NewInt(0).SetString(value, 10)
	if !ok {
		return
	}

	sumBig := balanceBigMap.Add(balanceBigMap, valueBig)
	pbp.pendingBalances[key].Balance = sumBig.String()
}

func (pbp *pendingBalancesProc) getAll() map[string]*data.AccountInfo {
	return pbp.pendingBalances
}
