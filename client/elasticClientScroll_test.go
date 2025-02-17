package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/multiversx/mx-chain-es-indexer-go/client/logging"
	"github.com/stretchr/testify/require"
)

func TestElasticClient_DoCountRequest(t *testing.T) {
	handler := http.NotFound
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()

	handler = func(w http.ResponseWriter, r *http.Request) {
		jsonFile, err := os.Open("./testsData/response-count-request.json")
		require.Nil(t, err)

		byteValue, _ := io.ReadAll(jsonFile)
		_, _ = w.Write(byteValue)
	}

	esClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
		Logger:    &logging.CustomLogger{},
	})

	count, err := esClient.DoCountRequest(context.Background(), "tokens", []byte(`{}`))
	require.Nil(t, err)
	require.Equal(t, uint64(112671), count)
}
