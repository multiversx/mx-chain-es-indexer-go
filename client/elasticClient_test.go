package client

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/multiversx/mx-chain-es-indexer-go/client/logging"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	indexer "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
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

		byteValue, _ := io.ReadAll(jsonFile)
		_, _ = w.Write(byteValue)
	}

	esClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
		Logger:    &logging.CustomLogger{},
	})

	ids := []string{"id"}
	res := &data.ResponseTokens{}
	err := esClient.DoMultiGet(context.Background(), ids, "tokens", true, res)
	require.Nil(t, err)
	require.Len(t, res.Docs, 3)

	resMap := make(objectsMap)
	err = esClient.DoMultiGet(context.Background(), ids, "tokens", true, &resMap)
	require.Nil(t, err)

	_, ok := resMap["docs"]
	require.True(t, ok)
}

func TestElasticClient_GetWriteIndexMultipleIndicesBehind(t *testing.T) {
	handler := http.NotFound
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()

	handler = func(w http.ResponseWriter, r *http.Request) {
		jsonFile, err := os.Open("./testsData/response-get-alias.json")
		require.Nil(t, err)

		byteValue, _ := io.ReadAll(jsonFile)
		_, _ = w.Write(byteValue)
	}

	esClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
		Logger:    &logging.CustomLogger{},
	})
	res, set, err := esClient.getWriteIndex("blocks")
	require.Nil(t, err)
	require.True(t, set)
	require.Equal(t, "blocks-000004", res)
}

func TestElasticClient_GetWriteIndexOneIndex(t *testing.T) {
	handler := http.NotFound
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()

	handler = func(w http.ResponseWriter, r *http.Request) {
		jsonFile, err := os.Open("./testsData/response-get-alias-only-one-index.json")
		require.Nil(t, err)

		byteValue, _ := io.ReadAll(jsonFile)
		_, _ = w.Write(byteValue)
	}

	esClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
		Logger:    &logging.CustomLogger{},
	})
	res, set, err := esClient.getWriteIndex("delegators")
	require.Nil(t, err)
	require.False(t, set)
	require.Equal(t, "delegators-000001", res)
}

func TestElasticClient_CheckAndCreateTemplate_AlreadyExists(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	esClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
	})

	err := esClient.CheckAndCreateTemplate("test-template", &bytes.Buffer{})
	require.Nil(t, err)
}

func TestElasticClient_CheckAndCreateTemplate_DoesNotExist(t *testing.T) {
	t.Parallel()

	numRequests := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++
		if numRequests == 1 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`{"acknowledged":true}`))
	}))
	defer ts.Close()

	esClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
	})

	err := esClient.CheckAndCreateTemplate("test-template", &bytes.Buffer{})
	require.Nil(t, err)
	require.Equal(t, 2, numRequests)
}

func TestElasticClient_CheckAndCreatePolicy_AlreadyExists(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	esClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
	})

	err := esClient.CheckAndCreatePolicy("test-policy", &bytes.Buffer{})
	require.Nil(t, err)
}

func TestElasticClient_CheckAndCreatePolicy_DoesNotExist(t *testing.T) {
	t.Parallel()

	numRequests := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++
		if numRequests == 1 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`{"acknowledged":true}`))
	}))
	defer ts.Close()

	esClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
	})

	err := esClient.CheckAndCreatePolicy("test-policy", &bytes.Buffer{})
	require.Nil(t, err)
	require.Equal(t, 2, numRequests)
}

func TestElasticClient_SetWriteIndexTrue_AlreadySet(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonFile, err := os.Open("./testsData/response-get-alias.json")
		require.Nil(t, err)

		byteValue, _ := io.ReadAll(jsonFile)
		_, _ = w.Write(byteValue)
	}))
	defer ts.Close()

	esClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
		Logger:    &logging.CustomLogger{},
	})

	err := esClient.SetWriteIndexTrue("blocks", "blocks-000005")
	require.Nil(t, err)
}

func TestElasticClient_SetWriteIndexTrue_NotSet(t *testing.T) {
	t.Parallel()

	numRequests := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++
		if numRequests == 1 {
			jsonFile, err := os.Open("./testsData/response-get-alias-only-one-index.json")
			require.Nil(t, err)

			byteValue, _ := io.ReadAll(jsonFile)
			_, _ = w.Write(byteValue)
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	esClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
		Logger:    &logging.CustomLogger{},
	})

	err := esClient.SetWriteIndexTrue("delegators", "delegators-000002")
	require.Nil(t, err)
	require.Equal(t, 2, numRequests)
}

func TestElasticClient_CheckAndCreateIndex_AlreadyExists(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	esClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
	})

	err := esClient.CheckAndCreateIndex("test-index")
	require.Nil(t, err)
}

func TestElasticClient_CheckAndCreateIndex_DoesNotExist(t *testing.T) {
	t.Parallel()

	numRequests := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++
		if numRequests == 1 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`{"acknowledged":true}`))
	}))
	defer ts.Close()

	esClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
	})

	err := esClient.CheckAndCreateIndex("test-index")
	require.Nil(t, err)
	require.Equal(t, 2, numRequests)
}

func TestElasticClient_PutMappings(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"acknowledged":true}`))
	}))
	defer ts.Close()

	esClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
	})

	err := esClient.PutMappings("test-index", &bytes.Buffer{})
	require.Nil(t, err)
}

func TestElasticClient_CheckAndCreateAlias_AlreadyExists(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	esClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
	})

	err := esClient.CheckAndCreateAlias("test-alias", "test-index")
	require.Nil(t, err)
}

func TestElasticClient_CheckAndCreateAlias_DoesNotExist(t *testing.T) {
	t.Parallel()

	numRequests := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++
		if numRequests == 1 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`{"acknowledged":true}`))
	}))
	defer ts.Close()

	esClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
	})

	err := esClient.CheckAndCreateAlias("test-alias", "test-index")
	require.Nil(t, err)
	require.Equal(t, 2, numRequests)
}

func TestElasticClient_DoBulkRequest(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errors":false,"items":[]}`))
	}))
	defer ts.Close()

	esClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
	})

	err := esClient.DoBulkRequest(context.Background(), &bytes.Buffer{}, "test-index")
	require.Nil(t, err)
}

func TestElasticClient_DoQueryRemove(t *testing.T) {
	t.Parallel()

	numRequests := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++
		switch numRequests {
		case 1:
			_, _ = w.Write([]byte(`{"_shards":{"total":1,"successful":1,"failed":0}}`))
		case 2:
			jsonFile, err := os.Open("./testsData/response-get-alias-only-one-index.json")
			require.Nil(t, err)

			byteValue, _ := io.ReadAll(jsonFile)
			_, _ = w.Write(byteValue)
		case 3:
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	defer ts.Close()

	esClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
		Logger:    &logging.CustomLogger{},
	})

	err := esClient.DoQueryRemove(context.Background(), "delegators", &bytes.Buffer{})
	require.Nil(t, err)
	require.Equal(t, 3, numRequests)
}

func TestElasticClient_UpdateByQuery(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()

	esClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
	})

	err := esClient.UpdateByQuery(context.Background(), "test-index", &bytes.Buffer{})
	require.Nil(t, err)
}

func TestElasticClient_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var ec *elasticClient
	require.True(t, ec.IsInterfaceNil())

	ec = &elasticClient{}
	require.False(t, ec.IsInterfaceNil())
}
