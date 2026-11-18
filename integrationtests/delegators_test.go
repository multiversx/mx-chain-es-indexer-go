//go:build integrationtests

package integrationtests

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	indexerdata "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestDelegateUnDelegateAndWithdraw(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	body := &dataBlock.Body{
		MiniBlocks: []*dataBlock.MiniBlock{
			{},
		},
	}
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
		ShardID:   core.MetachainShardId,
	}

	address1 := "erd1v7e552pz9py4hv6raan0c4jflez3e6csdmzcgrncg0qrnk4tywvsqx0h5j"

	// delegate
	delegatedValue, _ := big.NewInt(0).SetString("200000000000000000000", 10)
	pool := &outport.TransactionPool{
		Logs: []*transaction.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Address: decodeAddress("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqhllllsajxzat"),
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address1),
							Identifier: []byte("delegate"),
							Topics:     [][]byte{delegatedValue.Bytes(), delegatedValue.Bytes(), big.NewInt(10).Bytes(), delegatedValue.Bytes()},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids := []string{"9v/pLAXxUZJ4Oy1U+x5al/Xg5sebh1dYCRTeZwg/u68="}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.DelegatorsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/delegators/delegator-after-delegate.json"), string(genericResponse.Docs[0].Source))

	// unDelegate 1
	unDelegatedValue, _ := big.NewInt(0).SetString("50000000000000000000", 10)
	totalDelegation, _ := big.NewInt(0).SetString("150000000000000000000", 10)
	pool = &outport.TransactionPool{
		Logs: []*transaction.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h2")),
				Log: &transaction.Log{
					Address: decodeAddress("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqhllllsajxzat"),
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address1),
							Identifier: []byte("unDelegate"),
							Topics:     [][]byte{unDelegatedValue.Bytes(), totalDelegation.Bytes(), big.NewInt(10).Bytes(), totalDelegation.Bytes(), []byte("1")},
						},
						nil,
					},
				},
			},
		},
	}

	header.TimeStamp = 5050
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.DelegatorsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/delegators/delegator-after-un-delegate-1.json"), string(genericResponse.Docs[0].Source))

	// index the same block again: the unDelegate entry should be kept only once
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.DelegatorsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/delegators/delegator-after-un-delegate-1.json"), string(genericResponse.Docs[0].Source))

	// unDelegate 2
	unDelegatedValue, _ = big.NewInt(0).SetString("25500000000000000000", 10)
	totalDelegation, _ = big.NewInt(0).SetString("124500000000000000000", 10)
	pool = &outport.TransactionPool{
		Logs: []*transaction.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h3")),
				Log: &transaction.Log{
					Address: decodeAddress("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqhllllsajxzat"),
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address1),
							Identifier: []byte("unDelegate"),
							Topics:     [][]byte{unDelegatedValue.Bytes(), totalDelegation.Bytes(), big.NewInt(10).Bytes(), totalDelegation.Bytes(), []byte("2")},
						},
						nil,
					},
				},
			},
		},
	}

	header.TimeStamp = 5060
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.DelegatorsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/delegators/delegator-after-un-delegate-2.json"), string(genericResponse.Docs[0].Source))
	time.Sleep(time.Second)

	// revert unDelegate 2
	header.TimeStamp = 5060
	err = esProc.RemoveTransactions(header, body, 5060000)
	require.Nil(t, err)

	time.Sleep(time.Second)
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.DelegatorsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/delegators/delegator-after-revert.json"), string(genericResponse.Docs[0].Source))

	// withdraw
	withdrawValue, _ := big.NewInt(0).SetString("725500000000000000000", 10)
	pool = &outport.TransactionPool{
		Logs: []*transaction.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h4")),
				Log: &transaction.Log{
					Address: decodeAddress("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqhllllsajxzat"),
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address1),
							Identifier: []byte("withdraw"),
							Topics:     [][]byte{withdrawValue.Bytes(), totalDelegation.Bytes(), big.NewInt(10).Bytes(), totalDelegation.Bytes(), []byte("false"), []byte("1"), []byte("2")},
						},
						nil,
					},
				},
			},
		},
	}

	header.TimeStamp = 5070
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.DelegatorsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/delegators/delegator-after-withdraw.json"), string(genericResponse.Docs[0].Source))
}

func TestUnDelegatesWithTheSameFundKeyAreKept(t *testing.T) {
	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	body := &dataBlock.Body{
		MiniBlocks: []*dataBlock.MiniBlock{
			{},
		},
	}
	header := &dataBlock.Header{
		Round:     60,
		TimeStamp: 6010,
		ShardID:   core.MetachainShardId,
	}

	address1 := "erd1v7e552pz9py4hv6raan0c4jflez3e6csdmzcgrncg0qrnk4tywvsqx0h5j"
	ids := []string{"9v/pLAXxUZJ4Oy1U+x5al/Xg5sebh1dYCRTeZwg/u68="}
	bigValue := func(value string) *big.Int {
		result, _ := big.NewInt(0).SetString(value, 10)
		return result
	}

	// the fund key is the same for both undelegations, but they are two distinct operations of the delegator
	fundKey := []byte("3")
	totalDelegation := bigValue("100000000000000000000")
	unDelegate := func(txHash string, value *big.Int) *outport.TransactionPool {
		return &outport.TransactionPool{
			Logs: []*transaction.LogData{
				{
					TxHash: hex.EncodeToString([]byte(txHash)),
					Log: &transaction.Log{
						Address: decodeAddress("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqhllllsajxzat"),
						Events: []*transaction.Event{
							{
								Address:    decodeAddress(address1),
								Identifier: []byte("unDelegate"),
								Topics:     [][]byte{value.Bytes(), totalDelegation.Bytes(), big.NewInt(10).Bytes(), totalDelegation.Bytes(), fundKey},
							},
						},
					},
				},
			},
		}
	}

	firstPool := unDelegate("u1", bigValue("50000000000000000000"))
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, firstPool, nil, testNumOfShards))
	require.Nil(t, err)

	header.TimeStamp = 6020
	secondPool := unDelegate("u2", bigValue("25000000000000000000"))
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, secondPool, nil, testNumOfShards))
	require.Nil(t, err)

	// the same block is indexed again: its undelegation is still kept only once
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, secondPool, nil, testNumOfShards))
	require.Nil(t, err)

	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.DelegatorsIndex, true, genericResponse)
	require.Nil(t, err)

	type unDelegateEntry struct {
		ID    string `json:"id"`
		Value string `json:"value"`
	}
	type delegatorDocument struct {
		UnDelegateInfo []unDelegateEntry `json:"unDelegateInfo"`
	}
	document := &delegatorDocument{}
	require.Nil(t, json.Unmarshal(genericResponse.Docs[0].Source, document))

	fundKeyHex := hex.EncodeToString(fundKey)
	values := make([]string, 0)
	for _, entry := range document.UnDelegateInfo {
		if entry.ID == fundKeyHex {
			values = append(values, entry.Value)
		}
	}
	require.Equal(t, []string{"50000000000000000000", "25000000000000000000"}, values)
}
