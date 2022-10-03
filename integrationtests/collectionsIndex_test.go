//go:build integrationtests

package integrationtests

import (
	"encoding/hex"
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

	pool := &outport.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
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
	addr := "aaaabbbbcccccccc"
	addrHex := hex.EncodeToString([]byte(addr))

	addrForLog := "aaaabbbb"
	addrForLogHex := hex.EncodeToString([]byte(addrForLog))

	coreAlteredAccounts := map[string]*outport.AlteredAccount{
		addrHex: {
			Address: addrHex,
			Balance: "0",
			Tokens: []*outport.AccountTokenData{
				{
					Identifier: "SSSS-dddd",
					Balance:    "1000",
					Nonce:      2,
					Properties: "ok",
					MetaData: &esdt.MetaData{
						Creator: []byte("creator"),
					},
				},
			},
		},
		addrForLogHex: {
			Address: addrForLogHex,
			Balance: "0",
			Tokens: []*outport.AccountTokenData{
				{
					Identifier: "SSSS-dddd",
					Balance:    "1000",
					Nonce:      2,
					Properties: "ok",
					MetaData: &esdt.MetaData{
						Creator: []byte("creator"),
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
							Address:    []byte(addr),
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
	ids := []string{"61616161626262626363636363636363"}
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
							Address:    []byte(addr),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("SSSS-dddd"), big.NewInt(22).Bytes(), big.NewInt(1).Bytes(), esdtDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	coreAlteredAccounts[addrHex].Tokens[0].Nonce = 22
	coreAlteredAccounts[addrForLogHex].Tokens[0].Nonce = 22

	err = esProc.SaveTransactions(body, header, pool, coreAlteredAccounts, false, testNumOfShards)
	require.Nil(t, err)
	ids = []string{"61616161626262626363636363636363"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.CollectionsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/collectionsIndex/collections-2.json"), string(genericResponse.Docs[0].Source))

	// ################ TRANSFER SEMI FUNGIBLE TOKEN 2 ##########################

	addr = "aaaabbbbcccccccc"
	addrHex = hex.EncodeToString([]byte(addr))
	coreAlteredAccounts = map[string]*outport.AlteredAccount{
		addrHex: {
			Address: addrHex,
			Tokens: []*outport.AccountTokenData{
				{
					Identifier: "NFT-abcdef",
					Balance:    "0",
					Properties: "ok",
					MetaData: &esdt.MetaData{
						Creator: []byte("creator"),
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
							Address:    []byte(addr),
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
	ids = []string{"61616161626262626363636363636363"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.CollectionsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/collectionsIndex/collections-1.json"), string(genericResponse.Docs[0].Source))
}
