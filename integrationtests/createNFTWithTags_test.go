//go:build integrationtests

package integrationtests

import (
	"encoding/json"
	"math/big"
	"testing"

	indexerdata "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go-core/core"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
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

	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{
		SelfID: 0,
	}

	header := &dataBlock.Header{
		Round:     51,
		TimeStamp: 5600,
	}

	addr := "aaaabbbb"
	mockAccount := &mock.UserAccountStub{
		RetrieveValueFromDataTrieTrackerCalled: func(key []byte) ([]byte, error) {
			return json.Marshal(esdtToken)
		},
		AddressBytesCalled: func() []byte {
			return []byte(addr)
		},
	}
	accounts := &mock.AccountsStub{
		LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
			return mockAccount, nil
		},
	}
	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	esdtDataBytes, _ := json.Marshal(esdtToken)

	// CREATE A FIRST NFT WITH THE TAGS
	pool := &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("aaaabbbb"),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("DESK-abcd"), big.NewInt(1).Bytes(), big.NewInt(1).Bytes(), esdtDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	body := &dataBlock.Body{}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{"6161616162626262-DESK-abcd-01"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/createNFTWithTags/accounts-esdt-address-balance.json"), string(genericResponse.Docs[0].Source))

	ids = []string{"bXVzaWM=", "aGVsbG8=", "Z2FsbGVyeQ==", "ZG8=", "YXJ0", "c29tZXRoaW5n"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TagsIndex, true, genericResponse)
	require.Nil(t, err)

	for idx, id := range ids {
		expectedDoc := getElementFromSlice("./testdata/createNFTWithTags/tags1.json", idx)
		for _, doc := range genericResponse.Docs {
			if doc.ID == id {
				require.JSONEq(t, expectedDoc, string(doc.Source))
			}
		}
	}

	// CREATE A SECOND NFT WITH THE SAME TAGS
	pool = &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("aaaabbbb"),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("DESK-abcd"), big.NewInt(2).Bytes(), big.NewInt(1).Bytes(), esdtDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	body = &dataBlock.Body{}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TagsIndex, true, genericResponse)
	require.Nil(t, err)

	for idx, id := range ids {
		expectedDoc := getElementFromSlice("./testdata/createNFTWithTags/tags2.json", idx)
		for _, doc := range genericResponse.Docs {
			if doc.ID == id {
				require.JSONEq(t, expectedDoc, string(doc.Source))
			}
		}
	}

}
