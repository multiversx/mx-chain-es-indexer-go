package rest

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/tools/accounts-balance-checker/pkg/utils"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

const maxNumOfRetries = 10

var log = logger.GetOrCreate("restClient")

type restClient struct {
	httpClient *http.Client
	url        string
}

// NewRestClient will create a new instance of restClient
func NewRestClient(url string) (*restClient, error) {
	c := http.DefaultClient

	return &restClient{
		httpClient: c,
		url:        url,
	}, nil
}

// CallGetRestEndPoint calls an external end point (sends a get request)
func (rc *restClient) CallGetRestEndPoint(
	pathEndpoint string,
	value interface{},
) error {
	defer utils.LogExecutionTime(log, time.Now(), "rc.CallGetRestEndPoint "+pathEndpoint)

	req, err := http.NewRequest(http.MethodGet, rc.url+pathEndpoint, nil)
	if err != nil {
		return err
	}

	userAgent := "Accounts manager>"
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	var count = 0
TryAgain:
	resp, err := rc.httpClient.Do(req)
	if err != nil && count < maxNumOfRetries {
		log.Warn("rc.httpClient.Do", "error", err)
		count++
		sleep(count)
		goto TryAgain
	}
	if err != nil {
		return fmt.Errorf("too many retries, error: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusRequestTimeout {
		_ = resp.Body.Close()
		count++
		sleep(count)
		goto TryAgain
	}

	defer func() {
		errNotCritical := resp.Body.Close()
		if errNotCritical != nil {
			log.Warn("restClient.CallGetRestEndPoint: close body", "error", errNotCritical.Error())
		}
	}()

	err = json.NewDecoder(resp.Body).Decode(value)
	if err != nil {
		return err
	}

	return nil
}

func sleep(count int) {
	delay := time.Duration(math.Exp2(float64(count))) * time.Second
	time.Sleep(delay)
}
