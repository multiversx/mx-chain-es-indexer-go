//go:build integrationtests

package integrationtests

import (
	"encoding/json"
	"math/big"
	"testing"

	indexerdata "github.com/ElrondNetwork/elastic-indexer-go/data"
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
	expectedTokenAfterUpdate = `{"identifier":"NFT-abcd-0e","token":"NFT-abcd","nonce":14,"timestamp":5040,"data":{"uris":["dXJp","dXJp","dXJpMQ==","dXJpMg=="],"nonEmptyURIs":true,"whiteListedStorage":false}}`
)

func TestNFTUpdateMetadata(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esdtCreateData := &esdt.ESDigitalToken{
		TokenMetaData: &esdt.MetaData{
			URIs: [][]byte{[]byte("uri"), []byte("uri")},
		},
	}
	marshalizedCreate, _ := json.Marshal(esdtCreateData)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}
	body := &dataBlock.Body{}

	// CREATE NFT data
	pool := &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(1).Bytes(), marshalizedCreate},
						},
						nil,
					},
				},
				TxHash: "h1",
			},
		},
	}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	// Update NFT data
	pool = &indexer.Pool{
		Logs: []*coreData.LogData{
			{
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionESDTNFTAddURI),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("caller"), []byte("uri1"), []byte("uri2")},
						},
						nil,
					},
				},
				TxHash: "h1",
			},
		},
	}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{"NFT-abcd-0e"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.Equal(t, expectedTokenAfterUpdate, string(genericResponse.Docs[0].Source))
}
