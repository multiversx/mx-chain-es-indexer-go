//go:build integrationtests

package integrationtests

import (
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	indexerdata "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestIndexAccountESDTWithTokenType(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	// ################ ISSUE NON FUNGIBLE TOKEN ##########################
	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	body := &dataBlock.Body{}
	header := &dataBlock.Header{
		Round:     50,
		ShardID:   core.MetachainShardId,
		TimeStamp: 5040,
	}

	address := "erd1sqy2ywvswp09ef7qwjhv8zwr9kzz3xas6y2ye5nuryaz0wcnfzzsnq0am3"
	pool := &outport.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte("issueSemiFungible"),
							Topics:     [][]byte{[]byte("SEMI-abcd"), []byte("SEMI-token"), []byte("SEM"), []byte(core.SemiFungibleESDT)},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool, map[string]*outport.AlteredAccount{}, false, testNumOfShards)
	require.Nil(t, err)

	ids := []string{"SEMI-abcd"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsESDTWithTokenType/token-after-issue.json"), string(genericResponse.Docs[0].Source))

	// ################ CREATE SEMI FUNGIBLE TOKEN ##########################
	coreAlteredAccounts := map[string]*outport.AlteredAccount{
		address: {
			Address: address,
			Balance: "1000",
			Tokens: []*outport.AccountTokenData{
				{
					Identifier: "SEMI-abcd",
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
		ShardID:   2,
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
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("SEMI-abcd"), big.NewInt(2).Bytes(), big.NewInt(1).Bytes(), esdtDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool, coreAlteredAccounts, false, testNumOfShards)
	require.Nil(t, err)

	ids = []string{fmt.Sprintf("%s-SEMI-abcd-02", address)}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsESDTWithTokenType/account-esdt.json"), string(genericResponse.Docs[0].Source))

}

func TestIndexAccountESDTWithTokenTypeShardFirstAndMetachainAfter(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	// ################ CREATE SEMI FUNGIBLE TOKEN #########################
	body := &dataBlock.Body{}

	address := "erd1l29zsl2dqq988kvr2y0xlfv9ydgnvhzkatfd8ccalpag265pje8qn8lslf"
	coreAlteredAccounts := map[string]*outport.AlteredAccount{
		address: {
			Address: address,
			Balance: "1000",
			Tokens: []*outport.AccountTokenData{
				{
					Identifier: "TTTT-abcd",
					Nonce:      2,
					Balance:    "1000",
					Properties: "ok",
					MetaData: &outport.TokenMetaData{
						Creator: "erd1l29zsl2dqq988kvr2y0xlfv9ydgnvhzkatfd8ccalpag265pje8qn8lslf",
					},
				},
			},
		},
	}
	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	header := &dataBlock.Header{
		Round:     51,
		TimeStamp: 5600,
		ShardID:   2,
	}

	esdtData := &esdt.ESDigitalToken{
		TokenMetaData: &esdt.MetaData{
			Creator: decodeAddress(address),
		},
	}
	esdtDataBytes, _ := json.Marshal(esdtData)

	pool := &outport.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("TTTT-abcd"), big.NewInt(2).Bytes(), big.NewInt(1).Bytes(), esdtDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool, coreAlteredAccounts, false, testNumOfShards)
	require.Nil(t, err)

	ids := []string{fmt.Sprintf("%s-TTTT-abcd-02", address)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsESDTWithTokenType/account-esdt-without-type.json"), string(genericResponse.Docs[0].Source))

	time.Sleep(time.Second)

	// ################ ISSUE NON FUNGIBLE TOKEN ##########################
	header = &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
		ShardID:   core.MetachainShardId,
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
							Address:    decodeAddress(address),
							Identifier: []byte("issueSemiFungible"),
							Topics:     [][]byte{[]byte("TTTT-abcd"), []byte("TTTT-token"), []byte("SEM"), []byte(core.SemiFungibleESDT)},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool, map[string]*outport.AlteredAccount{}, false, testNumOfShards)
	require.Nil(t, err)

	ids = []string{"TTTT-abcd"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsESDTWithTokenType/semi-fungible-token.json"), string(genericResponse.Docs[0].Source))

	ids = []string{fmt.Sprintf("%s-TTTT-abcd-02", address)}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsESDTWithTokenType/account-esdt-with-type.json"), string(genericResponse.Docs[0].Source))

	ids = []string{"TTTT-abcd-02"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsESDTWithTokenType/semi-fungible-token-after-create.json"), string(genericResponse.Docs[0].Source))
}
