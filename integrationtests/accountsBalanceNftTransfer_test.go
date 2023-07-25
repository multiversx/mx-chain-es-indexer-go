//go:build integrationtests

package integrationtests

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	indexerdata "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func createOutportBlockWithHeader(
	body *dataBlock.Body,
	header coreData.HeaderHandler,
	pool *outport.TransactionPool,
	coreAlteredAccounts map[string]*alteredAccount.AlteredAccount,
	numOfShards uint32,
) *outport.OutportBlockWithHeader {
	return &outport.OutportBlockWithHeader{
		OutportBlock: &outport.OutportBlock{
			BlockData: &outport.BlockData{
				Body: body,
			},
			TransactionPool: pool,
			AlteredAccounts: coreAlteredAccounts,
			NumberOfShards:  numOfShards,
		},
		Header: header,
	}
}

func TestAccountBalanceNFTTransfer(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	// ################ CREATE NFT ##########################
	body := &dataBlock.Body{}

	addr := "erd1wdylghcn2uu393t703vufwa3ycdqfachgqyanha2xm2aqmsa5kfqg8qgrl"

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	header := &dataBlock.Header{
		Round:     51,
		TimeStamp: 5600,
		ShardID:   1,
	}

	pool := &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(addr),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("NFT-abcdef"), big.NewInt(7440483).Bytes(), big.NewInt(1).Bytes()},
						},
						nil,
					},
				},
			},
		},
	}

	coreAlteredAccounts := map[string]*alteredAccount.AlteredAccount{
		addr: {
			Address: addr,
			Tokens: []*alteredAccount.AccountTokenData{
				{
					Identifier: "NFT-abcdef",
					Nonce:      7440483,
					Balance:    "1000",
				},
			},
		},
	}

	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, coreAlteredAccounts, testNumOfShards))
	require.Nil(t, err)

	ids := []string{fmt.Sprintf("%s-NFT-abcdef-718863", addr)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceNftTransfer/balance-nft-after-create.json"), string(genericResponse.Docs[0].Source))

	// ################ TRANSFER NFT ##########################

	addrReceiver := "erd1caejdhq28fc03wddsf2lqs90jlwqlzesxjlyd0k2zeekxckpp6qsxty5x2"
	header = &dataBlock.Header{
		Round:     51,
		TimeStamp: 5600,
		ShardID:   1,
	}

	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("test-address-balance-1"),
							Identifier: []byte(core.BuiltInFunctionESDTNFTTransfer),
							Topics:     [][]byte{[]byte("NFT-abcdef"), big.NewInt(7440483).Bytes(), big.NewInt(1).Bytes(), decodeAddress(addrReceiver)},
						},
						nil,
					},
				},
			},
		},
	}

	esProc, err = CreateElasticProcessor(esClient)
	require.Nil(t, err)

	coreAlteredAccounts = map[string]*alteredAccount.AlteredAccount{
		addr: {
			Address: addr,
			Tokens: []*alteredAccount.AccountTokenData{
				{
					Identifier: "NFT-abcdef",
					Nonce:      7440483,
					Balance:    "0",
				},
			},
		},
		addrReceiver: {
			Address: addrReceiver,
			Tokens: []*alteredAccount.AccountTokenData{
				{
					Identifier: "NFT-abcdef",
					Nonce:      7440483,
					Balance:    "1000",
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, coreAlteredAccounts, testNumOfShards))
	require.Nil(t, err)

	ids = []string{fmt.Sprintf("%s-NFT-abcdef-718863", addr)}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.False(t, genericResponse.Docs[0].Found)

	ids = []string{fmt.Sprintf("%s-NFT-abcdef-718863", addrReceiver)}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceNftTransfer/balance-nft-after-transfer.json"), string(genericResponse.Docs[0].Source))
}
