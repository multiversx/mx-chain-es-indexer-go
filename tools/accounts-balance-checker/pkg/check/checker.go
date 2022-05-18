package check

import (
	"encoding/json"
	"errors"
	"fmt"
	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/tools/accounts-balance-checker/pkg/utils"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"time"

	indexerData "github.com/ElrondNetwork/elastic-indexer-go/data"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

const (
	accountsIndex                 = "accounts"
	accountEndpoint               = "/address/%s/balance"
	maxNumberOfRequestsInParallel = 40
)

var log = logger.GetOrCreate("checker")

type balanceChecker struct {
	balanceToFloat  indexer.BalanceConverter
	pubKeyConverter core.PubkeyConverter
	esClient        ESClientHandler
	restClient      RestClientHandler

	doRepair bool
}

func NewBalanceChecker(
	esClient ESClientHandler,
	restClient RestClientHandler,
	pubKeyConverter core.PubkeyConverter,
	balanceToFloat indexer.BalanceConverter,
	repair bool,
) (*balanceChecker, error) {
	return &balanceChecker{
		esClient:        esClient,
		restClient:      restClient,
		pubKeyConverter: pubKeyConverter,
		balanceToFloat:  balanceToFloat,
		doRepair:        repair,
	}, nil
}

func (bc *balanceChecker) CheckEGLDBalances() error {
	return bc.esClient.DoScrollRequestAllDocuments(
		accountsIndex,
		[]byte(matchAllQuery),
		bc.handlerFuncScrollAccountEGLD,
	)
}

var countCheck = 0

func (bc *balanceChecker) handlerFuncScrollAccountEGLD(responseBytes []byte) error {
	accountsRes := &ResponseAccounts{}
	err := json.Unmarshal(responseBytes, accountsRes)
	if err != nil {
		return err
	}
	countCheck++

	defer utils.LogExecutionTime(log, time.Now(), fmt.Sprintf("checked bulk of accounts %d", countCheck))

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
		newBalance, err := bc.getBalanceFromES(acct.Address)
		if err != nil {
			log.Error("something when wrong", "address", acct.Address, "error", err)
			return
		}
		if newBalance != gatewayBalance {
			log.Warn("balance mismatch",
				"address", acct.Address,
				"balance ES", newBalance,
				"balance proxy", gatewayBalance,
			)
			return
		}
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

func (bc *balanceChecker) getBalanceFromES(address string) (string, error) {
	encoded, _ := encodeQuery(getDocumentsByIDsQuery([]string{address}, true))
	accountsResponse := &ResponseAccounts{}
	err := bc.esClient.DoGetRequest(&encoded, accountsIndex, accountsResponse, 1)
	if err != nil {
		return "", err
	}

	if len(accountsResponse.Hits.Hits) == 0 {
		return "", fmt.Errorf("cannot find accounts with address: %s", address)
	}

	return accountsResponse.Hits.Hits[0].Source.Balance, nil
}
