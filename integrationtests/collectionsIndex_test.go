//go:build integrationtests

package integrationtests

import (
	"encoding/json"
	"math/big"
	"testing"

	indexerdata "github.com/ElrondNetwork/elastic-indexer-go/process/dataindexer"
	"github.com/ElrondNetwork/elrond-go-core/core"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestCollectionsIndexInsertAndDelete(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	// ################ ISSUE NON FUNGIBLE TOKEN ##########################
	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	body := &dataBlock.Body{}
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
		ShardID:   core.MetachainShardId,
	}

	address1 := "erd1v7e552pz9py4hv6raan0c4jflez3e6csdmzcgrncg0qrnk4tywvsqx0h5j"
	pool := &outport.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address1),
							Identifier: []byte("issueSemiFungible"),
							Topics:     [][]byte{[]byte("SSSS-dddd"), []byte("SEMI-semi"), []byte("SSS"), []byte(core.SemiFungibleESDT)},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool, nil, false, testNumOfShards)
	require.Nil(t, err)

	// ################ CREATE SEMI FUNGIBLE TOKEN 1 ##########################
	address2 := "erd1acjlnuhkd8773sqhmw85r0ur4lcyuqgm0n69h9ttxh0gwxtuuzxq4lckh6"
	address3 := "erd1emv0ffwkrq6wdszxrg2ktjxhtnqt8gsqp93v8pnl8qca5ewn4deshd8s5z"

	coreAlteredAccounts := map[string]*outport.AlteredAccount{
		address2: {
			Address: address2,
			Balance: "0",
			Tokens: []*outport.AccountTokenData{
				{
					Identifier: "SSSS-dddd",
					Balance:    "1000",
					Nonce:      2,
					Properties: "ok",
					MetaData: &outport.TokenMetaData{
						Creator: "creator",
					},
				},
			},
		},
		address3: {
			Address: address3,
			Balance: "0",
			Tokens: []*outport.AccountTokenData{
				{
					Identifier: "SSSS-dddd",
					Balance:    "1000",
					Nonce:      2,
					Properties: "ok",
					MetaData: &outport.TokenMetaData{
						Creator: "creator",
					},
				},
			},
		},
	}
	esProc, err = CreateElasticProcessor(esClient)
	require.Nil(t, err)

	header = &dataBlock.Header{
		Round:     51,
		TimeStamp: 5600,
		ShardID:   1,
	}

	esdtData := &esdt.ESDigitalToken{
		TokenMetaData: &esdt.MetaData{
			Creator: []byte("creator"),
		},
	}
	esdtDataBytes, _ := json.Marshal(esdtData)

	pool = &outport.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address2),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("SSSS-dddd"), big.NewInt(2).Bytes(), big.NewInt(1).Bytes(), esdtDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool, coreAlteredAccounts, false, testNumOfShards)
	require.Nil(t, err)
	ids := []string{address2}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.CollectionsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/collectionsIndex/collections-1.json"), string(genericResponse.Docs[0].Source))

	// ################ CREATE SEMI FUNGIBLE TOKEN 2 ##########################
	pool = &outport.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address2),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("SSSS-dddd"), big.NewInt(22).Bytes(), big.NewInt(1).Bytes(), esdtDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	coreAlteredAccounts[address2].Tokens[0].Nonce = 22
	coreAlteredAccounts[address2].Tokens[0].Nonce = 22

	err = esProc.SaveTransactions(body, header, pool, coreAlteredAccounts, false, testNumOfShards)
	require.Nil(t, err)
	ids = []string{address2}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.CollectionsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/collectionsIndex/collections-2.json"), string(genericResponse.Docs[0].Source))

	// ################ TRANSFER SEMI FUNGIBLE TOKEN 2 ##########################

	coreAlteredAccounts = map[string]*outport.AlteredAccount{
		address2: {
			Address: address2,
			Tokens: []*outport.AccountTokenData{
				{
					Identifier: "SSSS-dddd",
					Nonce:      22,
					Balance:    "0",
					Properties: "ok",
					MetaData: &outport.TokenMetaData{
						Creator: "creator",
					},
				},
			},
		},
	}

	esProc, err = CreateElasticProcessor(esClient)
	require.Nil(t, err)

	pool = &outport.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address2),
							Identifier: []byte(core.BuiltInFunctionESDTNFTTransfer),
							Topics:     [][]byte{[]byte("SSSS-dddd"), big.NewInt(22).Bytes(), big.NewInt(1).Bytes(), []byte("746573742d616464726573732d62616c616e63652d31")},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool, coreAlteredAccounts, false, testNumOfShards)
	require.Nil(t, err)
	ids = []string{address2}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.CollectionsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/collectionsIndex/collections-1.json"), string(genericResponse.Docs[0].Source))
}
