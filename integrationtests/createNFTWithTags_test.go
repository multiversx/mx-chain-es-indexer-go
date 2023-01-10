//go:build integrationtests

package integrationtests

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	indexerdata "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestCreateNFTWithTags(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esdtToken := &esdt.ESDigitalToken{
		Value:      big.NewInt(1000),
		Properties: []byte("ok"),
		TokenMetaData: &esdt.MetaData{
			Creator:    []byte("creator"),
			Attributes: []byte("tags:hello,something,do,music,art,gallery;metadata:QmZ2QqaGq4bqsEzs5JLTjRmmvR2GAR4qXJZBN8ibfDdaud"),
		},
	}

	header := &dataBlock.Header{
		Round:     51,
		TimeStamp: 5600,
		ShardID:   2,
	}

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	esdtDataBytes, _ := json.Marshal(esdtToken)

	// CREATE A FIRST NFT WITH THE TAGS
	address1 := "erd1v7e552pz9py4hv6raan0c4jflez3e6csdmzcgrncg0qrnk4tywvsqx0h5j"
	pool := &outport.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address1),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("DESK-abcd"), big.NewInt(1).Bytes(), big.NewInt(1).Bytes(), esdtDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	coreAlteredAccounts := map[string]*outport.AlteredAccount{
		address1: {
			Address: address1,
			Balance: "0",
			Tokens: []*outport.AccountTokenData{
				{
					AdditionalData: &outport.AdditionalAccountTokenData{
						IsNFTCreate: true,
					},
					Identifier: "DESK-abcd",
					Nonce:      1,
					Balance:    "1000",
					Properties: "ok",
					MetaData: &outport.TokenMetaData{
						Creator:    "creator",
						Attributes: []byte("tags:hello,something,do,music,art,gallery;metadata:QmZ2QqaGq4bqsEzs5JLTjRmmvR2GAR4qXJZBN8ibfDdaud"),
					},
				},
			},
		},
	}

	body := &dataBlock.Body{}
	err = esProc.SaveTransactions(body, header, pool, coreAlteredAccounts, false, testNumOfShards)
	require.Nil(t, err)

	ids := []string{fmt.Sprintf("%s-DESK-abcd-01", address1)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/createNFTWithTags/accounts-esdt-address-balance.json"), string(genericResponse.Docs[0].Source))

	ids = []string{"bXVzaWM=", "aGVsbG8=", "Z2FsbGVyeQ==", "ZG8=", "YXJ0", "c29tZXRoaW5n"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TagsIndex, true, genericResponse)
	require.Nil(t, err)

	tagsChecked := 0
	for idx, id := range ids {
		expectedDoc := getElementFromSlice("./testdata/createNFTWithTags/tags1.json", idx)
		for _, doc := range genericResponse.Docs {
			if doc.ID == id {
				require.JSONEq(t, expectedDoc, string(doc.Source))
				tagsChecked++
			}
		}
	}
	require.Equal(t, len(ids), tagsChecked)

	// CREATE A SECOND NFT WITH THE SAME TAGS
	pool = &outport.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address1),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("DESK-abcd"), big.NewInt(2).Bytes(), big.NewInt(1).Bytes(), esdtDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	coreAlteredAccounts[address1].Tokens[0].Nonce = 2
	body = &dataBlock.Body{}
	err = esProc.SaveTransactions(body, header, pool, coreAlteredAccounts, false, testNumOfShards)
	require.Nil(t, err)

	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TagsIndex, true, genericResponse)
	require.Nil(t, err)

	tagsChecked = 0
	for idx, id := range ids {
		expectedDoc := getElementFromSlice("./testdata/createNFTWithTags/tags2.json", idx)
		for _, doc := range genericResponse.Docs {
			if doc.ID == id {
				require.JSONEq(t, expectedDoc, string(doc.Source))
				tagsChecked++
			}
		}
	}
	require.Equal(t, len(ids), tagsChecked)

	// CREATE A 3RD NFT WITH THE SPECIAL TAGS
	hexEncodedAttributes := "746167733a5c5c5c5c5c5c5c5c5c5c5c5c5c5c5c5c5c5c5c2c3c3c3c3e3e3e2626262626262626262626262626262c272727273b6d657461646174613a516d533757525566464464516458654c513637516942394a33663746654d69343554526d6f79415741563568345a"
	attributes, _ := hex.DecodeString(hexEncodedAttributes)

	coreAlteredAccounts[address1].Tokens[0].Nonce = 3
	coreAlteredAccounts[address1].Tokens[0].MetaData.Attributes = attributes

	esProc, err = CreateElasticProcessor(esClient)
	require.Nil(t, err)

	pool = &outport.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address1),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("DESK-abcd"), big.NewInt(3).Bytes(), big.NewInt(1).Bytes(), esdtDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	body = &dataBlock.Body{}
	err = esProc.SaveTransactions(body, header, pool, coreAlteredAccounts, false, testNumOfShards)
	require.Nil(t, err)

	ids = append(ids, "XFxcXFxcXFxcXFxcXFxcXFxcXA==", "JycnJw==", "PDw8Pj4+JiYmJiYmJiYmJiYmJiYm")
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TagsIndex, true, genericResponse)
	require.Nil(t, err)

	tagsChecked = 0
	for idx, id := range ids {
		expectedDoc := getElementFromSlice("./testdata/createNFTWithTags/tags3.json", idx)
		for _, doc := range genericResponse.Docs {
			if doc.ID == id {
				require.JSONEq(t, expectedDoc, string(doc.Source))
				tagsChecked++
			}
		}
	}
	require.Equal(t, len(ids), tagsChecked)
}
