package integrationtests

import (
	"context"
	"testing"

	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	indexerdata "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestIndexMiniBlocksOnSourceAndDestination(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)
	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	// index on the source shard
	header := &dataBlock.Header{
		ShardID:   1,
		TimeStamp: 1234,
	}
	miniBlocks := []*dataBlock.MiniBlock{
		{
			SenderShardID:   1,
			ReceiverShardID: 2,
		},
	}
	err = esProc.SaveMiniblocks(header, miniBlocks)
	require.Nil(t, err)
	mbHash := "11a1bb4065e16a2e93b2b5ac5957b7b69f1cfba7579b170b24f30dab2d3162e0"
	ids := []string{mbHash}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.MiniblocksIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/miniblocks/cross-miniblock-on-source.json"), string(genericResponse.Docs[0].Source))

	// index on the destination shard
	mbhr := &dataBlock.MiniBlockHeaderReserved{
		ExecutionType: dataBlock.ProcessingType(1),
	}

	marshaller := &marshal.GogoProtoMarshalizer{}
	mbhrBytes, _ := marshaller.Marshal(mbhr)
	header = &dataBlock.Header{
		ShardID:   2,
		TimeStamp: 1234,
		MiniBlockHeaders: []dataBlock.MiniBlockHeader{
			{
				Reserved: mbhrBytes,
			},
		},
	}

	err = esProc.SaveMiniblocks(header, miniBlocks)
	require.Nil(t, err)

	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.MiniblocksIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/miniblocks/cross-miniblock-on-destination.json"), string(genericResponse.Docs[0].Source))
}

func TestIndexMiniBlockFirstOnDestinationAndAfterSource(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)
	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	// index on destination
	mbhr := &dataBlock.MiniBlockHeaderReserved{
		ExecutionType: dataBlock.ProcessingType(2),
	}

	marshaller := &marshal.GogoProtoMarshalizer{}
	mbhrBytes, _ := marshaller.Marshal(mbhr)
	header := &dataBlock.Header{
		ShardID:   0,
		TimeStamp: 54321,
		MiniBlockHeaders: []dataBlock.MiniBlockHeader{
			{
				Reserved: mbhrBytes,
			},
		},
	}
	miniBlocks := []*dataBlock.MiniBlock{
		{
			SenderShardID:   2,
			ReceiverShardID: 0,
		},
	}

	err = esProc.SaveMiniblocks(header, miniBlocks)
	require.Nil(t, err)
	genericResponse := &GenericResponse{}

	ids := []string{"2f3ee0ff3b6426916df3b123a10f425b7e2027e2ae8d231229d27b12aa522ade"}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.MiniblocksIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/miniblocks/cross-miniblock-on-destination-first.json"), string(genericResponse.Docs[0].Source))

	// index on source
	mbhr = &dataBlock.MiniBlockHeaderReserved{
		ExecutionType: dataBlock.ProcessingType(0),
	}
	mbhrBytes, _ = marshaller.Marshal(mbhr)
	header.ShardID = 2
	header.MiniBlockHeaders = []dataBlock.MiniBlockHeader{
		{
			Reserved: mbhrBytes,
		},
	}
	err = esProc.SaveMiniblocks(header, miniBlocks)
	require.Nil(t, err)
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.MiniblocksIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/miniblocks/cross-miniblock-on-source-second.json"), string(genericResponse.Docs[0].Source))
}
