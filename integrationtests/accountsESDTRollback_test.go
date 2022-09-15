//go:build integrationtests

package integrationtests

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	indexerdata "github.com/ElrondNetwork/elastic-indexer-go/process/dataindexer"
	"github.com/ElrondNetwork/elrond-go-core/core"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestAccountsESDTDeleteOnRollback(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	shardCoordinator := &mock.ShardCoordinatorMock{
		SelfID: 1,
		ComputeIdCalled: func(address []byte) uint32 {
			return 1
		},
	}

	esdtToken := &esdt.ESDigitalToken{
		Value:      big.NewInt(1000),
		Properties: []byte("ok"),
		TokenMetaData: &esdt.MetaData{
			Creator: []byte("creator"),
		},
	}
	addr := hex.EncodeToString([]byte("aaaabbbb"))
	coreAlteredAccounts := map[string]*outport.AlteredAccount{
		addr: {
			Address: addr,
			Tokens: []*outport.AccountTokenData{
				{
					Identifier: "TOKEN-eeee",
					Nonce:      2,
					Balance:    "1000",
					MetaData: &esdt.MetaData{
						Creator: []byte("creator"),
					},
					Properties: "ok",
				},
			},
		},
	}

	esProc, err := CreateElasticProcessor(esClient, shardCoordinator)
	require.Nil(t, err)

	// CREATE SEMI-FUNGIBLE TOKEN
	esdtDataBytes, _ := json.Marshal(esdtToken)
	pool := &outport.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("aaaabbbb"),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("TOKEN-eeee"), big.NewInt(2).Bytes(), big.NewInt(1).Bytes(), esdtDataBytes},
						},
						nil,
					},
				},
			},
		},
	}

	body := &dataBlock.Body{}
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}

	err = esProc.SaveTransactions(body, header, pool, coreAlteredAccounts, false)
	require.Nil(t, err)

	ids := []string{"6161616162626262-TOKEN-eeee-02"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsESDTRollback/account-after-create.json"), string(genericResponse.Docs[0].Source))

	// DO ROLLBACK
	err = esProc.RemoveAccountsESDT(5040)
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.False(t, genericResponse.Docs[0].Found)
}
