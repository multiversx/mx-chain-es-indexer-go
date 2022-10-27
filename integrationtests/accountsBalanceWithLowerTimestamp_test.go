//go:build integrationtests

package integrationtests

import (
	"encoding/hex"
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

func TestIndexAccountsBalance(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	// ################ UPDATE ACCOUNT-ESDT BALANCE ##########################
	body := &dataBlock.Body{}

	esdtToken := &esdt.ESDigitalToken{
		Value: big.NewInt(1000),
	}

	addr := "aaaabbbb"
	addr2 := "eeeebbbb"
	encodedAddr := hex.EncodeToString([]byte(addr))
	encodedAddr2 := hex.EncodeToString([]byte(addr2))

	alteredAccount := &outport.AlteredAccount{
		Address: encodedAddr,
		Balance: "0",
		Tokens: []*outport.AccountTokenData{
			{
				Identifier: "TTTT-abcd",
				Balance:    "1000",
				Nonce:      0,
			},
		},
	}

	coreAlteredAccounts := map[string]*outport.AlteredAccount{
		encodedAddr:  alteredAccount,
		encodedAddr2: alteredAccount,
	}

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	header := &dataBlock.Header{
		Round:     51,
		TimeStamp: 5600,
		ShardID:   2,
	}

	pool := &outport.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("eeeebbbb"),
							Identifier: []byte(core.BuiltInFunctionESDTTransfer),
							Topics:     [][]byte{[]byte("TTTT-abcd"), nil, big.NewInt(1).Bytes()},
						},
						nil,
					},
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool, coreAlteredAccounts, false, testNumOfShards)
	require.Nil(t, err)

	ids := []string{"6161616162626262"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceWithLowerTimestamp/account-balance-first-update.json"), string(genericResponse.Docs[0].Source))

	ids = []string{"6161616162626262-TTTT-abcd-00"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceWithLowerTimestamp/account-balance-esdt-first-update.json"), string(genericResponse.Docs[0].Source))

	//////////////////// INDEX BALANCE LOWER TIMESTAMP ///////////////////////////////////

	header = &dataBlock.Header{
		Round:     51,
		TimeStamp: 5000,
		ShardID:   2,
	}

	err = esProc.SaveTransactions(body, header, pool, map[string]*outport.AlteredAccount{}, false, testNumOfShards)
	require.Nil(t, err)

	ids = []string{"6161616162626262"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceWithLowerTimestamp/account-balance-first-update.json"), string(genericResponse.Docs[0].Source))

	ids = []string{"6161616162626262-TTTT-abcd-00"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceWithLowerTimestamp/account-balance-esdt-first-update.json"), string(genericResponse.Docs[0].Source))

	//////////////////// INDEX BALANCE HIGHER TIMESTAMP ///////////////////////////////////
	header = &dataBlock.Header{
		Round:     51,
		TimeStamp: 6000,
		ShardID:   2,
	}

	coreAlteredAccounts[encodedAddr].Balance = "2000"
	coreAlteredAccounts[encodedAddr].BalanceChanged = true
	coreAlteredAccounts[encodedAddr].IsSender = true
	pool = &outport.Pool{
		Txs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			"h1": outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
				SndAddr: []byte(addr),
			}, 0, big.NewInt(0)),
		},
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("eeeebbbb"),
							Identifier: []byte(core.BuiltInFunctionESDTTransfer),
							Topics:     [][]byte{[]byte("TTTT-abcd"), nil, big.NewInt(1).Bytes()},
						},
						nil,
					},
				},
			},
		},
	}
	body = &dataBlock.Body{
		MiniBlocks: []*dataBlock.MiniBlock{
			{
				Type:          dataBlock.TxBlock,
				TxHashes:      [][]byte{[]byte("h1")},
				SenderShardID: 2,
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool, coreAlteredAccounts, false, testNumOfShards)
	require.Nil(t, err)

	ids = []string{"6161616162626262"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceWithLowerTimestamp/account-balance-second-update.json"), string(genericResponse.Docs[0].Source))

	ids = []string{"6161616162626262-TTTT-abcd-00"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceWithLowerTimestamp/account-balance-esdt-second-update.json"), string(genericResponse.Docs[0].Source))

	//////////////////////// DELETE ESDT BALANCE LOWER TIMESTAMP ////////////////

	esdtToken.Value = big.NewInt(0)
	encodedAddr = hex.EncodeToString([]byte(addr))
	esProc, err = CreateElasticProcessor(esClient)
	require.Nil(t, err)

	header = &dataBlock.Header{
		Round:     51,
		TimeStamp: 6001,
		ShardID:   2,
	}

	coreAlteredAccounts[encodedAddr].Balance = "2000"
	coreAlteredAccounts[encodedAddr].Tokens[0].Balance = "0"
	coreAlteredAccounts[encodedAddr].BalanceChanged = false
	coreAlteredAccounts[encodedAddr].IsSender = false

	pool.Txs = make(map[string]coreData.TransactionHandlerWithGasUsedAndFee)
	err = esProc.SaveTransactions(body, header, pool, coreAlteredAccounts, false, testNumOfShards)
	require.Nil(t, err)

	ids = []string{"6161616162626262"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceWithLowerTimestamp/account-balance-second-update.json"), string(genericResponse.Docs[0].Source))

	ids = []string{"6161616162626262-TTTT-abcd-00"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.False(t, genericResponse.Docs[0].Found)
}
