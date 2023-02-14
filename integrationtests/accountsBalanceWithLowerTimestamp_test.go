//go:build integrationtests

package integrationtests

import (
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

func TestIndexAccountsBalance(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	// ################ UPDATE ACCOUNT-ESDT BALANCE ##########################
	body := &dataBlock.Body{}

	esdtToken := &esdt.ESDigitalToken{
		Value: big.NewInt(1000),
	}

	addr := "erd17umc0uvel62ng30k5uprqcxh3ue33hq608njejaqljuqzqlxtzuqeuzlcv"
	addr2 := "erd1m2pyjudsqt8gn0tnsstht35gfqcfx8ku5utz07mf2r6pq3sfxjzszhcx6w"

	alteredAccount := &outport.AlteredAccount{
		Address: addr,
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
		addr:  alteredAccount,
		addr2: alteredAccount,
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

	ids := []string{addr}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceWithLowerTimestamp/account-balance-first-update.json"), string(genericResponse.Docs[0].Source))

	ids = []string{fmt.Sprintf("%s-TTTT-abcd-00", addr)}
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

	ids = []string{addr}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceWithLowerTimestamp/account-balance-first-update.json"), string(genericResponse.Docs[0].Source))

	ids = []string{fmt.Sprintf("%s-TTTT-abcd-00", addr)}
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

	coreAlteredAccounts[addr].Balance = "2000"
	coreAlteredAccounts[addr].AdditionalData = &outport.AdditionalAccountData{
		IsSender:       true,
		BalanceChanged: true,
	}
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
							Address:    decodeAddress(addr2),
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

	ids = []string{addr}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceWithLowerTimestamp/account-balance-second-update.json"), string(genericResponse.Docs[0].Source))

	ids = []string{fmt.Sprintf("%s-TTTT-abcd-00", addr)}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceWithLowerTimestamp/account-balance-esdt-second-update.json"), string(genericResponse.Docs[0].Source))

	//////////////////////// DELETE ESDT BALANCE LOWER TIMESTAMP ////////////////

	esdtToken.Value = big.NewInt(0)
	esProc, err = CreateElasticProcessor(esClient)
	require.Nil(t, err)

	header = &dataBlock.Header{
		Round:     51,
		TimeStamp: 6001,
		ShardID:   2,
	}

	coreAlteredAccounts[addr].Balance = "2000"
	coreAlteredAccounts[addr].Tokens[0].Balance = "0"
	coreAlteredAccounts[addr].AdditionalData = &outport.AdditionalAccountData{
		IsSender:       false,
		BalanceChanged: false,
	}

	pool.Txs = make(map[string]coreData.TransactionHandlerWithGasUsedAndFee)
	err = esProc.SaveTransactions(body, header, pool, coreAlteredAccounts, false, testNumOfShards)
	require.Nil(t, err)

	ids = []string{addr}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceWithLowerTimestamp/account-balance-second-update.json"), string(genericResponse.Docs[0].Source))

	ids = []string{fmt.Sprintf("%s-TTTT-abcd-00", addr)}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.False(t, genericResponse.Docs[0].Found)
}
