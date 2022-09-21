//go:build integrationtests

package integrationtests

import (
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
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestIndexAccountESDTWithTokenType(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}

	// ################ ISSUE NON FUNGIBLE TOKEN ##########################
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
							Topics:     [][]byte{[]byte("SEMI-abcd"), []byte("SEMI-token"), []byte("SEM"), []byte(core.SemiFungibleESDT)},
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
	require.JSONEq(t, readExpectedResult("./testdata/accountsESDTWithTokenType/token-after-issue.json"), string(genericResponse.Docs[0].Source))

	// ################ CREATE SEMI FUNGIBLE TOKEN ##########################
	shardCoordinator = &mock.ShardCoordinatorMock{
		SelfID: 0,
	}

	esdtToken := &esdt.ESDigitalToken{
		Value:      big.NewInt(1000),
		Properties: []byte("ok"),
		TokenMetaData: &esdt.MetaData{
			Creator: []byte("creator"),
		},
	}

	addr := "aaaabbbb"
	mockAccount := &mock.UserAccountStub{
		RetrieveValueCalled: func(key []byte) ([]byte, error) {
			return json.Marshal(esdtToken)
		},
		AddressBytesCalled: func() []byte {
			return []byte(addr)
		},
	}
	accounts = &mock.AccountsStub{
		LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
			return mockAccount, nil
		},
	}
	esProc, err = CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
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

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids = []string{"6161616162626262-SEMI-abcd-02"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsESDTWithTokenType/account-esdt.json"), string(genericResponse.Docs[0].Source))

}

func TestIndexAccountESDTWithTokenTypeShardFirstAndMetachainAfter(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}

	// ################ CREATE SEMI FUNGIBLE TOKEN ##########################
	shardCoordinator := &mock.ShardCoordinatorMock{
		SelfID: 0,
	}

	body := &dataBlock.Body{}

	esdtToken := &esdt.ESDigitalToken{
		Value:      big.NewInt(1000),
		Properties: []byte("ok"),
		TokenMetaData: &esdt.MetaData{
			Creator: []byte("creator"),
		},
	}

	addr := "aaaabbbb"
	mockAccount := &mock.UserAccountStub{
		RetrieveValueCalled: func(key []byte) ([]byte, error) {
			return json.Marshal(esdtToken)
		},
		AddressBytesCalled: func() []byte {
			return []byte(addr)
		},
	}
	accounts = &mock.AccountsStub{
		LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
			return mockAccount, nil
		},
	}
	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
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

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{"6161616162626262-TTTT-abcd-02"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsESDTWithTokenType/account-esdt-without-type.json"), string(genericResponse.Docs[0].Source))

	time.Sleep(time.Second)

	// ################ ISSUE NON FUNGIBLE TOKEN ##########################
	shardCoordinator = &mock.ShardCoordinatorMock{
		SelfID: core.MetachainShardId,
	}
	header = &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}

	esProc, err = CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
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

	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids = []string{"TTTT-abcd"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsESDTWithTokenType/semi-fungible-token.json"), string(genericResponse.Docs[0].Source))

	ids = []string{"6161616162626262-TTTT-abcd-02"}
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
