//go:build integrationtests

package integrationtests

import (
	"context"
	"encoding/hex"
	"testing"

	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
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

	body := &dataBlock.Body{MiniBlocks: miniBlocks}
	pool := &outport.TransactionPool{}
	ob := createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards)
	ob.BlockData.HeaderHash, _ = hex.DecodeString("3fede8a9a3c4f2ba6d7e6e01541813606cd61c4d3af2940f8e089827b5d94e50")
	err = esProc.SaveMiniblocks(ob)
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

	ob = createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards)
	ob.BlockData.HeaderHash, _ = hex.DecodeString("d7f1e8003a45c7adbd87bbbb269cb4af3d1f4aedd0c214973bfc096dd0f3b65e")
	err = esProc.SaveMiniblocks(ob)
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

	pool := &outport.TransactionPool{}
	body := &dataBlock.Body{MiniBlocks: miniBlocks}
	ob := createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards)
	ob.BlockData.HeaderHash, _ = hex.DecodeString("b36435faaa72390772da84f418348ce0d477c74432579519bf0ffea1dc4c36e9")
	ob.BlockData.TimestampMs = 54321000
	err = esProc.SaveMiniblocks(ob)
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

	ob = createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards)
	ob.BlockData.HeaderHash, _ = hex.DecodeString("b601381e1f41df2aa3da9f2b8eb169f14c86418229e30fc65f9e6b37b7f0d902")
	err = esProc.SaveMiniblocks(ob)
	require.Nil(t, err)
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.MiniblocksIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/miniblocks/cross-miniblock-on-source-second.json"), string(genericResponse.Docs[0].Source))
}
