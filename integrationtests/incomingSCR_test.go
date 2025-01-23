//go:build integrationtests

package integrationtests

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/stretchr/testify/require"

	indexerData "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/factory"
)

func TestSovereignTransactionWithScCallSuccess(t *testing.T) {
	setLogLevelDebug()

	mainChainEs := factory.MainChainElastic{
		Url: testnetEsURL,
	}

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateSovereignElasticProcessor(esClient, mainChainEs)
	require.Nil(t, err)

	txHash := []byte("txHash")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}
	body := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   4294967293,
				ReceiverShardID: 0,
				TxHashes:        [][]byte{txHash},
			},
		},
	}

	token1Identifier := "AGE-be2571"
	token2Identifier := "BGD16-c47f46"
	data := []byte(core.BuiltInFunctionMultiESDTNFTTransfer +
		"@02" +
		"@" + hex.EncodeToString([]byte(token1Identifier)) +
		"@" +
		"@" + hex.EncodeToString(big.NewInt(123).Bytes()) +
		"@" + hex.EncodeToString([]byte(token2Identifier)) +
		"@" +
		"@" + hex.EncodeToString(big.NewInt(333).Bytes()))

	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), []string{token1Identifier, token2Identifier}, indexerData.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	for _, token := range genericResponse.Docs {
		require.False(t, token.Found)
	}

	pool := &outport.TransactionPool{
		SmartContractResults: map[string]*outport.SCRInfo{
			hex.EncodeToString(txHash): {SmartContractResult: &smartContractResult.SmartContractResult{
				Nonce:          11,
				Value:          big.NewInt(0),
				GasLimit:       0,
				SndAddr:        decodeAddress("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"),
				RcvAddr:        decodeAddress("erd1kzrfl2tztgzjpeedwec37c8npcr0a2ulzh9lhmj7xufyg23zcxuqxcqz0s"),
				Data:           data,
				OriginalTxHash: nil,
			}, FeeInfo: &outport.FeeInfo{}},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, 1))
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerData.ScResultsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t,
		readExpectedResult("./testdata/incomingSCR/incoming-scr.json"),
		string(genericResponse.Docs[0].Source),
	)

	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), []string{token1Identifier, token2Identifier}, indexerData.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	for _, token := range genericResponse.Docs {
		require.True(t, token.Found)
	}
}
