package integrationtests

import (
	"testing"

	indexerdata "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go-core/core"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

const (
	expectedTokenSemi                       = `{"name":"semi-token","ticker":"SEMI","token":"SEMI-abcd","issuer":"61646472","currentOwner":"61646472","type":"SemiFungibleESDT","timestamp":5040,"ownersHistory":[{"address":"61646472","timestamp":5040}]}`
	expectedTokenSemiAfterTransferOwnership = `{"name":"semi-token","ticker":"SEMI","token":"SEMI-abcd","issuer":"61646472","currentOwner":"6e65772d61646472657373","type":"SemiFungibleESDT","timestamp":5040,"ownersHistory":[{"address":"61646472","timestamp":5040},{"address":"6e65772d61646472657373","timestamp":10000}]}`
)

func TestIssueTokenAndTransferOwnership(t *testing.T) {
	t.Skip("integration test, should be run only when doing debugging")

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
		Logs: []indexer.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte("issueSemiFungible"),
							Topics:     [][]byte{[]byte("SEMI-abcd"), []byte("semi-token"), []byte("SEMI"), []byte(core.SemiFungibleESDT)},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{"SEMI-abcd"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, expectedTokenSemi, string(genericResponse.Docs[0].Source))

	// transfer ownership
	pool = &indexer.Pool{
		Logs: []indexer.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte("transferOwnership"),
							Topics:     [][]byte{[]byte("SEMI-abcd"), []byte("semi-token"), []byte("SEMI"), []byte(core.SemiFungibleESDT), []byte("new-address")},
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

	ids = []string{"SEMI-abcd"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, expectedTokenSemiAfterTransferOwnership, string(genericResponse.Docs[0].Source))
}
