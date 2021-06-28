package client

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/stretchr/testify/require"
)

func TestElasticClient_NewClientEmptyUrl(t *testing.T) {
	esClient, err := NewElasticClient(elasticsearch.Config{
		Addresses: []string{},
	})
	require.Nil(t, esClient)
	require.Equal(t, indexer.ErrNoElasticUrlProvided, err)
}

func TestElasticClient_NewClient(t *testing.T) {
	handler := http.NotFound
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()

	handler = func(w http.ResponseWriter, r *http.Request) {
		resp := ``
		_, _ = w.Write([]byte(resp))
	}

	esClient, err := NewElasticClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
	})
	require.Nil(t, err)
	require.NotNil(t, esClient)
}

func TestElasticClient_DoMultiGet(t *testing.T) {
	handler := http.NotFound
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()

	handler = func(w http.ResponseWriter, r *http.Request) {
		jsonFile, err := os.Open("./testsData/response-multi-get.json")
		require.Nil(t, err)

		byteValue, _ := ioutil.ReadAll(jsonFile)
		_, _ = w.Write(byteValue)
	}

	esClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
	})

	ids := []string{"id"}
	res := &data.ResponseTokens{}
	err := esClient.DoMultiGet(ids, "tokens", true, res)
	require.Nil(t, err)
	require.Len(t, res.Docs, 3)

	resMap := make(objectsMap)
	err = esClient.DoMultiGet(ids, "tokens", true, &resMap)
	require.Nil(t, err)

	_, ok := resMap["docs"]
	require.True(t, ok)
}
