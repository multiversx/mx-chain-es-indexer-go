//go:build integrationtests

package integrationtests

import (
	"context"
	"encoding/hex"
	"testing"

	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	indexerdata "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestIndexExecutionResults(t *testing.T) {
	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	header := &dataBlock.HeaderV3{
		ExecutionResults: []*dataBlock.ExecutionResult{
			{
				BaseExecutionResult: &dataBlock.BaseExecutionResult{
					HeaderHash:  []byte("h1"),
					HeaderNonce: 1,
					HeaderRound: 2,
					HeaderEpoch: 3,
				},
			},
			{
				BaseExecutionResult: &dataBlock.BaseExecutionResult{
					HeaderHash:  []byte("h2"),
					HeaderNonce: 2,
					HeaderRound: 3,
					HeaderEpoch: 3,
				},
			},
		},
	}
	pool := &outport.TransactionPool{}
	body := &dataBlock.Body{}
	ob := createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards)
	ob.BlockData.HeaderHash = []byte("hh1")
	ob.HeaderGasConsumption = &outport.HeaderGasConsumption{}
	err = esProc.SaveHeader(ob)
	require.Nil(t, err)

	ids := []string{hex.EncodeToString([]byte("h1")), hex.EncodeToString([]byte("h2"))}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.ExecutionResultsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t, readExpectedResult("./testdata/executionResults/execution-result-1.json"), string(genericResponse.Docs[0].Source))
	require.JSONEq(t, readExpectedResult("./testdata/executionResults/execution-result-2.json"), string(genericResponse.Docs[1].Source))
}
