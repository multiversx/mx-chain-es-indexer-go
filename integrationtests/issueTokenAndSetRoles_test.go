//go:build integrationtests

package integrationtests

import (
	"math/big"
	"testing"

	indexerdata "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go-core/core"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

const (
	expectedTokenAfterIssueTok  = `{"name":"semi-token","ticker":"SEMI","token":"TOK-abcd","issuer":"61646472","currentOwner":"61646472","type":"SemiFungibleESDT","timestamp":5040,"ownersHistory":[{"address":"61646472","timestamp":5040}]}`
	expectedTokenAfterSetRole   = `{"name":"semi-token","ticker":"SEMI","token":"TOK-abcd","issuer":"61646472","currentOwner":"61646472","type":"SemiFungibleESDT","timestamp":5040,"ownersHistory":[{"address":"61646472","timestamp":5040}],"roles":{"ESDTRoleNFTCreate":{"6d792d61646472657373":10000},"ESDTRoleNFTBurn":{"6d792d61646472657373":10000}}}`
	expectedTokenAfterUnsetRole = `{"name":"semi-token","ticker":"SEMI","token":"TOK-abcd","issuer":"61646472","currentOwner":"61646472","type":"SemiFungibleESDT","timestamp":5040,"ownersHistory":[{"address":"61646472","timestamp":5040}],"roles":{"ESDTRoleNFTCreate":{"6d792d61646472657373":10000},"ESDTRoleNFTBurn":{}}}`

	expectedTokenObjAfterSetRolesFirst    = `{"roles":{"ESDTRoleNFTCreate":{"6d792d61646472657373":10000},"ESDTRoleNFTBurn":{"6d792d61646472657373":10000}}}`
	expectedTokenObjAfterSetRolesAndIssue = `{"ticker":"SEMI","currentOwner":"61646472","roles":{"ESDTRoleNFTCreate":{"6d792d61646472657373":10000},"ESDTRoleNFTBurn":{"6d792d61646472657373":10000}},"name":"semi-token","type":"SemiFungibleESDT","issuer":"61646472","token":"TTT-abcd","timestamp":10000,"ownersHistory":[{"address":"61646472","timestamp":10000}]}`
)

func TestIssueTokenAndSetRole(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{
		SelfID: core.MetachainShardId,
	}

	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
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
							Topics:     [][]byte{[]byte("TOK-abcd"), []byte("semi-token"), []byte("SEMI"), []byte(core.SemiFungibleESDT)},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{"TOK-abcd"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, expectedTokenAfterIssueTok, string(genericResponse.Docs[0].Source))

	// SET ROLES
	pool = &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("my-address"),
							Identifier: []byte(core.BuiltInFunctionSetESDTRole),
							Topics:     [][]byte{[]byte("TOK-abcd"), big.NewInt(0).Bytes(), big.NewInt(0).Bytes(), []byte(core.ESDTRoleNFTCreate), []byte(core.ESDTRoleNFTBurn)},
						},
						nil,
					},
				},
			},
		},
	}

	header.TimeStamp = 10000
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids = []string{"TOK-abcd"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, expectedTokenAfterSetRole, string(genericResponse.Docs[0].Source))

	// UNSET ROLES

	pool = &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("my-address"),
							Identifier: []byte(core.BuiltInFunctionUnSetESDTRole),
							Topics:     [][]byte{[]byte("TOK-abcd"), big.NewInt(0).Bytes(), big.NewInt(0).Bytes(), []byte(core.ESDTRoleNFTBurn)},
						},
						nil,
					},
				},
			},
		},
	}

	header.TimeStamp = 10000
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids = []string{"TOK-abcd"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, expectedTokenAfterUnsetRole, string(genericResponse.Docs[0].Source))
}

func TestIssueSetRolesEventAndAfterTokenIssue(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}

	shardCoordinator := &mock.ShardCoordinatorMock{
		SelfID: core.MetachainShardId,
	}

	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	body := &dataBlock.Body{}
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}

	// SET ROLES
	pool := &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("my-address"),
							Identifier: []byte(core.BuiltInFunctionSetESDTRole),
							Topics:     [][]byte{[]byte("TTT-abcd"), big.NewInt(0).Bytes(), big.NewInt(0).Bytes(), []byte(core.ESDTRoleNFTCreate), []byte(core.ESDTRoleNFTBurn)},
						},
						nil,
					},
				},
			},
		},
	}

	header.TimeStamp = 10000
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{"TTT-abcd"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, expectedTokenObjAfterSetRolesFirst, string(genericResponse.Docs[0].Source))

	pool = &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte("issueSemiFungible"),
							Topics:     [][]byte{[]byte("TTT-abcd"), []byte("semi-token"), []byte("SEMI"), []byte(core.SemiFungibleESDT)},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids = []string{"TTT-abcd"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, expectedTokenObjAfterSetRolesAndIssue, string(genericResponse.Docs[0].Source))
}
