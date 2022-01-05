package comparer

import (
	"encoding/json"
	"math"
	"strings"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/tools/balance-comparer/pkg/client"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/elastic/go-elasticsearch/v7"
)

var log = logger.GetOrCreate("index-modifier/pkg/alterindex")

type balanceESDTComparer struct {
	scrollClient  ScrollClient
	balanceGetter *balanceGetter
}

func backOff(i int) time.Duration {
	d := time.Duration(math.Exp2(float64(i))) * time.Second
	log.Info("elastic: retry backoff", "attempt", i, "sleep duration", d)
	return d
}

func createScrollESClient(esURL string) (ScrollClient, error) {
	cfg := elasticsearch.Config{
		Addresses:     []string{esURL},
		MaxRetries:    0,
		RetryBackoff:  backOff,
		RetryOnStatus: []int{429, 502, 503, 504},
	}
	return client.NewElasticClient(cfg)
}

func NewBalanceComparerESDT(esURL string, gatewayURL string) (*balanceESDTComparer, error) {
	esClient, err := createScrollESClient(esURL)
	if err != nil {
		return nil, err
	}

	bg, err := newBalanceGetter(gatewayURL)
	if err != nil {
		return nil, err
	}

	return &balanceESDTComparer{
		scrollClient:  esClient,
		balanceGetter: bg,
	}, nil
}

func (bce *balanceESDTComparer) CompareBalances() error {
	balancesFromES, err := bce.getAllBalancesFromES()
	if err != nil {
		return err
	}

	balancesFromGateway, err := bce.getBalancesFromGateway(balancesFromES)
	if err != nil {
		return err
	}
	balancesFromES.compare(balancesFromGateway)

	return nil
}

func (bce *balanceESDTComparer) getBalancesFromGateway(esBke *balancesKeepEsdt) (*balancesKeepEsdt, error) {
	bke := newBalancesKeepESDT()
	for _, addr := range esBke.getAllAddresses() {
		esdtBalances, err := bce.balanceGetter.getESDTBalances(addr)
		if err != nil {
			log.Error("cannot get balance", "address", addr, "error", err.Error())
		}

		bke.addMultiple(addr, esdtBalances)
		compareAddressBalances(addr, esBke.getBalances(addr), esdtBalances)
	}

	return bke, nil
}

func (bce *balanceESDTComparer) getAllBalancesFromES() (*balancesKeepEsdt, error) {
	bke := newBalancesKeepESDT()

	count := 0
	handlerFunc := func(bodyBytes []byte) error {
		resESDTs := &responseAccountsESDT{}
		err := json.Unmarshal(bodyBytes, resESDTs)
		if err != nil {
			return err
		}

		for _, res := range resESDTs.Hits.Hits {
			if strings.HasPrefix(res.Source.Address, "pending") {
				continue
			}

			id := res.Source.TokenIdentifier
			if id == "" {
				id = res.Source.TokenName
			}

			bke.add(res.Source.Address, id, res.Source.Balance)
		}

		count++
		log.Info("fetch esdt accounts", "count", count)

		return nil
	}

	err := bce.scrollClient.DoScrollRequestAllDocuments("accountsesdt", []byte(`{"query": {"match_all": {}}}`), handlerFunc)
	if err != nil {
		return nil, err
	}

	return bke, nil
}
