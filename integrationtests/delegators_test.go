package integrationtests

import (
	"math/big"
	"testing"

	indexerdata "github.com/ElrondNetwork/elastic-indexer-go/process/dataindexer"
	"github.com/ElrondNetwork/elrond-go-core/core"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestDelegateUnDelegateAndWithdraw(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	body := &dataBlock.Body{}
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
		ShardID:   core.MetachainShardId,
	}

	address1 := "erd1v7e552pz9py4hv6raan0c4jflez3e6csdmzcgrncg0qrnk4tywvsqx0h5j"

	// delegate
	delegatedValue, _ := big.NewInt(0).SetString("200000000000000000000", 10)
	pool := &outport.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
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

	err = esProc.SaveTransactions(body, header, pool, nil, false, testNumOfShards)
	require.Nil(t, err)

	ids := []string{"9v/pLAXxUZJ4Oy1U+x5al/Xg5sebh1dYCRTeZwg/u68="}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.DelegatorsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/delegators/delegator-after-delegate.json"), string(genericResponse.Docs[0].Source))

	// unDelegate 1
	unDelegatedValue, _ := big.NewInt(0).SetString("50000000000000000000", 10)
	totalDelegation, _ := big.NewInt(0).SetString("150000000000000000000", 10)
	pool = &outport.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h2",
				LogHandler: &transaction.Log{
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
	err = esProc.SaveTransactions(body, header, pool, nil, false, testNumOfShards)
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexerdata.DelegatorsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/delegators/delegator-after-un-delegate-1.json"), string(genericResponse.Docs[0].Source))

	// unDelegate 2
	unDelegatedValue, _ = big.NewInt(0).SetString("25500000000000000000", 10)
	totalDelegation, _ = big.NewInt(0).SetString("124500000000000000000", 10)
	pool = &outport.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h3",
				LogHandler: &transaction.Log{
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
	err = esProc.SaveTransactions(body, header, pool, nil, false, testNumOfShards)
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexerdata.DelegatorsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/delegators/delegator-after-un-delegate-2.json"), string(genericResponse.Docs[0].Source))

	// withdraw
	withdrawValue, _ := big.NewInt(0).SetString("725500000000000000000", 10)
	pool = &outport.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h4",
				LogHandler: &transaction.Log{
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
	err = esProc.SaveTransactions(body, header, pool, nil, false, testNumOfShards)
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexerdata.DelegatorsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/delegators/delegator-after-withdraw.json"), string(genericResponse.Docs[0].Source))
}
