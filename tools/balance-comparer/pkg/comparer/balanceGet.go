package comparer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type balanceGetter struct {
	client   httpClient
	proxyURL string
}

func newBalanceGetter(gatewayURL string) (*balanceGetter, error) {
	client := http.DefaultClient

	return &balanceGetter{
		client:   client,
		proxyURL: gatewayURL,
	}, nil
}

func (bg *balanceGetter) getESDTBalances(address string) (map[string]string, error) {
	endpoint := fmt.Sprintf("address/%s/esdt", address)
	bodyBybytes, err := bg.getHTTP(context.Background(), endpoint)
	if err != nil {
		return nil, err
	}

	balancesESDT := &balancesESDTResponse{}
	err = json.Unmarshal(bodyBybytes, balancesESDT)
	if err != nil {
		return nil, err
	}
	if balancesESDT.Error != "" {
		return nil, errors.New(balancesESDT.Error)
	}

	balances := map[string]string{}
	for token, res := range balancesESDT.Data.ESDTS {
		balances[token] = res.Balance
	}

	return balances, nil
}

func (bg *balanceGetter) getHTTP(ctx context.Context, endpoint string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", bg.proxyURL, endpoint)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	response, err := bg.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
