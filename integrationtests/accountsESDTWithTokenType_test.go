//go:build integrationtests

package integrationtests

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	indexerdata "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go-core/core"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

const (
	expectedTokenAfterIssue = `{"name":"SEMI-token","ticker":"SEM","token":"SEMI-abcd","issuer":"61646472","currentOwner":"61646472","type":"SemiFungibleESDT","timestamp":5040,"ownersHistory":[{"address":"61646472","timestamp":5040}]}`
	expectedAccountESDT     = `{"address":"6161616162626262","balance":"1000","balanceNum":1e-15,"token":"SEMI-abcd","identifier":"SEMI-abcd-02","tokenNonce":2,"properties":"6f6b","data":{"creator":"63726561746f72","nonEmptyURIs":false,"whiteListedStorage":false},"timestamp":5600,"type":"SemiFungibleESDT"}`

	expectedAccountsESDTWithoutType      = `{"address":"6161616162626262","balance":"1000","balanceNum":1e-15,"token":"TTTT-abcd","identifier":"TTTT-abcd-02","tokenNonce":2,"properties":"6f6b","data":{"creator":"63726561746f72","nonEmptyURIs":false,"whiteListedStorage":false},"timestamp":5600}`
	expectedSemiFungibleToken            = `{"name":"TTTT-token","ticker":"SEM","token":"TTTT-abcd","issuer":"61646472","currentOwner":"61646472","type":"SemiFungibleESDT","timestamp":5040,"ownersHistory":[{"address":"61646472","timestamp":5040}]}`
	expectedAccountsESDTWithType         = `{"address":"6161616162626262","balance":"1000","balanceNum":1.0E-15,"token":"TTTT-abcd","identifier":"TTTT-abcd-02","tokenNonce":2,"properties":"6f6b","data":{"creator":"63726561746f72","nonEmptyURIs":false,"whiteListedStorage":false},"timestamp":5600,"type":"SemiFungibleESDT"}`
	expectedSemiFungibleTokenAfterCreate = `{"identifier":"TTTT-abcd-02","token":"TTTT-abcd","nonce":2,"timestamp":5600,"data":{"creator":"63726561746f72","nonEmptyURIs":false,"whiteListedStorage":false},"type":"SemiFungibleESDT"}`
)

func TestIndexAccountESDTWithTokenType(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	feeComputer := &mock.EconomicsHandlerMock{}

	// ################ ISSUE NON FUNGIBLE TOKEN ##########################
	shardCoordinator := &mock.ShardCoordinatorMock{
		SelfID: core.MetachainShardId,
	}

	esProc, err := CreateElasticProcessor(esClient, shardCoordinator, feeComputer)
	require.Nil(t, err)

	body := &dataBlock.Body{}
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}

	pool := &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte("issueSemiFungible"),
							Topics:     [][]byte{[]byte("SEMI-abcd"), []byte("SEMI-token"), []byte("SEM"), []byte(core.SemiFungibleESDT)},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool, map[string]*indexer.AlteredAccount{})
	require.Nil(t, err)

	ids := []string{"SEMI-abcd"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, expectedTokenAfterIssue, string(genericResponse.Docs[0].Source))

	// ################ CREATE SEMI FUNGIBLE TOKEN ##########################
	shardCoordinator = &mock.ShardCoordinatorMock{
		SelfID: 0,
	}

	addr := "aaaabbbb"
	encodedAddr := hex.EncodeToString([]byte(addr))
	coreAlteredAccounts := map[string]*indexer.AlteredAccount{
		encodedAddr: {
			Address: encodedAddr,
			Balance: "1000",
			Tokens: []*indexer.AccountTokenData{
				{
					Identifier: "SEMI-abcd",
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
	esProc, err = CreateElasticProcessor(esClient, shardCoordinator, feeComputer)
	require.Nil(t, err)

	header = &dataBlock.Header{
		Round:     51,
		TimeStamp: 5600,
	}

	esdtData := &esdt.ESDigitalToken{
		TokenMetaData: &esdt.MetaData{
			Creator: []byte("creator"),
		},
	}
	esdtDataBytes, _ := json.Marshal(esdtData)

	pool = &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("aaaabbbb"),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("SEMI-abcd"), big.NewInt(2).Bytes(), big.NewInt(1).Bytes(), esdtDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool, coreAlteredAccounts)
	require.Nil(t, err)

	ids = []string{"6161616162626262-SEMI-abcd-02"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, expectedAccountESDT, string(genericResponse.Docs[0].Source))

}

func TestIndexAccountESDTWithTokenTypeShardFirstAndMetachainAfter(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	feeComputer := &mock.EconomicsHandlerMock{}

	// ################ CREATE SEMI FUNGIBLE TOKEN ##########################
	shardCoordinator := &mock.ShardCoordinatorMock{
		SelfID: 0,
	}

	body := &dataBlock.Body{}

	addr := "aaaabbbb"
	encodedAddr := hex.EncodeToString([]byte(addr))
	coreAlteredAccounts := map[string]*indexer.AlteredAccount{
		encodedAddr: {
			Address: encodedAddr,
			Balance: "1000",
			Tokens: []*indexer.AccountTokenData{
				{
					Identifier: "TTTT-abcd",
					Nonce:      2,
					Balance:    "1000",
					Properties: "ok",
					MetaData: &esdt.MetaData{
						Creator: []byte("creator"),
					},
				},
			},
		},
	}
	esProc, err := CreateElasticProcessor(esClient, shardCoordinator, feeComputer)
	require.Nil(t, err)

	header := &dataBlock.Header{
		Round:     51,
		TimeStamp: 5600,
	}

	esdtData := &esdt.ESDigitalToken{
		TokenMetaData: &esdt.MetaData{
			Creator: []byte("creator"),
		},
	}
	esdtDataBytes, _ := json.Marshal(esdtData)

	pool := &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("aaaabbbb"),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("TTTT-abcd"), big.NewInt(2).Bytes(), big.NewInt(1).Bytes(), esdtDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool, coreAlteredAccounts)
	require.Nil(t, err)

	ids := []string{"6161616162626262-TTTT-abcd-02"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, expectedAccountsESDTWithoutType, string(genericResponse.Docs[0].Source))

	time.Sleep(time.Second)

	// ################ ISSUE NON FUNGIBLE TOKEN ##########################
	shardCoordinator = &mock.ShardCoordinatorMock{
		SelfID: core.MetachainShardId,
	}
	header = &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}

	esProc, err = CreateElasticProcessor(esClient, shardCoordinator, feeComputer)
	require.Nil(t, err)

	pool = &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte("issueSemiFungible"),
							Topics:     [][]byte{[]byte("TTTT-abcd"), []byte("TTTT-token"), []byte("SEM"), []byte(core.SemiFungibleESDT)},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool, map[string]*indexer.AlteredAccount{})
	require.Nil(t, err)

	ids = []string{"TTTT-abcd"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, expectedSemiFungibleToken, string(genericResponse.Docs[0].Source))

	ids = []string{"6161616162626262-TTTT-abcd-02"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, expectedAccountsESDTWithType, string(genericResponse.Docs[0].Source))

	ids = []string{"TTTT-abcd-02"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, expectedSemiFungibleTokenAfterCreate, string(genericResponse.Docs[0].Source))
}
