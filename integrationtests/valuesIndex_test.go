//go:build integrationtests

package integrationtests

import (
	"context"
	"fmt"
	"testing"

	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	indexerData "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/factory"
	"github.com/stretchr/testify/require"
)

func TestCheckVersionIsIndexer(t *testing.T) {
	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	version := "v1.4.5"
	args := factory.ArgElasticProcessorFactory{
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		AddressPubkeyConverter:   pubKeyConverter,
		ValidatorPubkeyConverter: mock.NewPubkeyConverterMock(32),
		DBClient:                 esClient,
		Denomination:             18,
		Version:                  version,
		EnabledIndexes:           []string{indexerData.ValuesIndex},
	}

	_, err = factory.CreateElasticProcessor(args)
	require.Nil(t, err)

	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), []string{"indexer-version"}, indexerData.ValuesIndex, true, genericResponse)
	require.Nil(t, err)
	require.Equal(t, fmt.Sprintf(`{"key":"indexer-version","value":"%s"}`, version), string(genericResponse.Docs[0].Source))
}

func TestSaveTransactionsUpdatesShardRoundInValuesIndex(t *testing.T) {
	esClient, err := createESClient(esURL)
	require.NoError(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.NoError(t, err)

	for _, round := range []uint64{100, 101} {
		obh := createOutportBlockWithHeader(
			&dataBlock.Body{},
			&dataBlock.Header{ShardID: 1, Round: round},
			&outport.TransactionPool{},
			nil,
			testNumOfShards,
		)
		err = esProc.SaveTransactions(obh)
		require.NoError(t, err)

		genericResponse := &GenericResponse{}
		err = esClient.DoMultiGet(context.Background(), []string{"shard_1"}, indexerData.ValuesIndex, true, genericResponse)
		require.NoError(t, err)
		require.Len(t, genericResponse.Docs, 1)
		require.True(t, genericResponse.Docs[0].Found)
		require.Equal(t, fmt.Sprintf(`{"key":"shard_1","value":"%d"}`, round), string(genericResponse.Docs[0].Source))
	}
}
