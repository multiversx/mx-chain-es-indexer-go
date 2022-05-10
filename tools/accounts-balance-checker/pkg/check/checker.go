package check

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	indexerData "github.com/ElrondNetwork/elastic-indexer-go/data"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

const (
	accountEndpoint               = "/address/%s/balance"
	maxNumberOfRequestsInParallel = 40
)

var log = logger.GetOrCreate("checker")

type balanceChecker struct {
	esClient   ESClientHandler
	restClient RestClientHandler
}

func NewBalanceChecker(
	esClient ESClientHandler,
	restClient RestClientHandler,
) (*balanceChecker, error) {
	return &balanceChecker{
		esClient:   esClient,
		restClient: restClient,
	}, nil
}

func (bc *balanceChecker) CheckEGLDBalances() error {
	return bc.esClient.DoScrollRequestAllDocuments(
		"accounts",
		[]byte(`{ "query": { "match_all": { } } }`),
		bc.handlerFuncScroll,
	)
}

var countCheck = 0

func (bc *balanceChecker) handlerFuncScroll(responseBytes []byte) error {
	accountsRes := &ResponseAccounts{}
	err := json.Unmarshal(responseBytes, accountsRes)
	if err != nil {
		return err
	}
	countCheck++

	defer logExecutionTime(time.Now(), fmt.Sprintf("checked bulk of accounts %d", countCheck))

	maxGoroutines := maxNumberOfRequestsInParallel
	done := make(chan struct{}, maxGoroutines)
	for _, acct := range accountsRes.Hits.Hits {
		done <- struct{}{}
		go bc.checkBalance(acct.Source, done)
	}

	return nil
}

func (bc *balanceChecker) checkBalance(acct indexerData.AccountInfo, done chan struct{}) {
	defer func() {
		<-done
	}()

	if acct.Balance == "0" {
		return
	}
	gatewayBalance, errGetBalance := bc.getAccountBalance(acct.Address)
	if errGetBalance != nil {
		log.Error("cannot get balance for address",
			"address", acct.Address,
			"error", errGetBalance)
		return
	}

	if gatewayBalance != acct.Balance {
		log.Warn("balance mismatch",
			"address", acct.Address,
			"balance ES", acct.Balance,
			"balance proxy", gatewayBalance,
		)
	}
}

func (bc *balanceChecker) getAccountBalance(address string) (string, error) {
	endpoint := fmt.Sprintf(accountEndpoint, address)

	accountResponse := &AccountResponse{}
	err := bc.restClient.CallGetRestEndPoint(endpoint, accountResponse)
	if err != nil {
		return "", err
	}
	if accountResponse.Error != "" {
		return "", errors.New(accountResponse.Error)
	}

	return accountResponse.Data.Balance, nil
}

func logExecutionTime(start time.Time, message string) {
	log.Info(message, "duration in seconds", time.Since(start).Seconds())
}
