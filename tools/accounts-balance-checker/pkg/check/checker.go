package check

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	indexer "github.com/multiversx/mx-chain-es-indexer-go/"
	indexerData "github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/tools/accounts-balance-checker/pkg/utils"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const (
	accountsIndex          = "accounts"
	addressBalanceEndpoint = "/address/%s/balance"
)

var log = logger.GetOrCreate("checker")

type balanceChecker struct {
	balanceToFloat              indexer.BalanceConverter
	pubKeyConverter             core.PubkeyConverter
	esClient                    ESClientHandler
	restClient                  RestClientHandler
	maxNumberOfParallelRequests int

	doRepair bool
}

// NewBalanceChecker will create a new instance of balanceChecker
func NewBalanceChecker(
	esClient ESClientHandler,
	restClient RestClientHandler,
	pubKeyConverter core.PubkeyConverter,
	balanceToFloat indexer.BalanceConverter,
	repair bool,
	maxNumberOfRequestsInParallel int,
) (*balanceChecker, error) {
	if check.IfNilReflect(esClient) {
		return nil, errors.New("nil elastic client")
	}
	if check.IfNilReflect(restClient) {
		return nil, errors.New("nil rest client")
	}
	if check.IfNil(pubKeyConverter) {
		return nil, errors.New("nil pub key converter")
	}
	if check.IfNilReflect(balanceToFloat) {
		return nil, errors.New("nil balance converter")
	}

	return &balanceChecker{
		esClient:                    esClient,
		restClient:                  restClient,
		pubKeyConverter:             pubKeyConverter,
		balanceToFloat:              balanceToFloat,
		doRepair:                    repair,
		maxNumberOfParallelRequests: maxNumberOfRequestsInParallel,
	}, nil
}

// CheckEGLDBalances will compare the EGLD balance from the Elasticsearch database with the results from gateway
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

	maxGoroutines := bc.maxNumberOfParallelRequests
	done := make(chan struct{}, maxGoroutines)
	for _, acct := range accountsRes.Hits.Hits {
		done <- struct{}{}
		go bc.checkBalance(acct.Source, done)
	}

	log.Info("comparing", "bulk count", countCheck)

	return nil
}

func (bc *balanceChecker) checkBalance(acct indexerData.AccountInfo, done chan struct{}) {
	defer func() {
		<-done
	}()

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
			log.Error("something went wrong", "address", acct.Address, "error", err)
			return
		}
		if newBalance != gatewayBalance {
			timestampLast, _ := bc.getLasTimeWhenBalanceWasChanged("", acct.Address)
			timestampString := formatTimestamp(int64(timestampLast))

			err = bc.fixWrongBalance(acct.Address, "", uint64(timestampLast), gatewayBalance, accountsIndex)
			if err != nil {
				log.Warn("cannot update balance from es", "addr", acct.Address, "data", timestampString)
			}

			log.Warn("balance mismatch",
				"address", acct.Address,
				"balance ES", newBalance,
				"balance proxy", gatewayBalance,
				"data", timestampString,
			)
			return
		}
	}
}

func (bc *balanceChecker) getAccountBalance(address string) (string, error) {
	endpoint := fmt.Sprintf(addressBalanceEndpoint, address)

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
