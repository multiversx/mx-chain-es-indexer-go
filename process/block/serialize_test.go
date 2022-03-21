package block

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go-core/core"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/stretchr/testify/require"
)

func TestBlockProcessor_SerializeBlockNilElasticBlockErrors(t *testing.T) {
	t.Parallel()

	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	err := bp.SerializeBlock(nil, nil, "")
	require.True(t, errors.Is(err, indexer.ErrNilElasticBlock))
}

func TestBlockProcessor_SerializeBlock(t *testing.T) {
	t.Parallel()

	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	buffSlice := data.NewBufferSlice(data.DefaultBulkSizeThreshold)
	err := bp.SerializeBlock(&data.Block{Nonce: 1}, buffSlice, "blocks")
	require.Nil(t, err)
	require.Equal(t, `{ "index" : { "_index":"blocks", "_id" : "" } }
{"nonce":1,"round":0,"epoch":0,"miniBlocksHashes":null,"notarizedBlocksHashes":null,"proposer":0,"validators":null,"pubKeyBitmap":"","size":0,"sizeTxs":0,"timestamp":0,"stateRootHash":"","prevHash":"","shardId":0,"txCount":0,"notarizedTxsCount":0,"accumulatedFees":"","developerFees":"","epochStartBlock":false,"searchOrder":0,"gasProvided":0,"gasRefunded":0,"gasPenalized":0,"maxGasLimit":0}
`, buffSlice.Buffers()[0].String())
}

func TestBlockProcessor_SerializeEpochInfoDataErrors(t *testing.T) {
	t.Parallel()

	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	err := bp.SerializeEpochInfoData(nil, nil, "")
	require.Equal(t, indexer.ErrNilHeaderHandler, err)

	err = bp.SerializeEpochInfoData(&dataBlock.Header{}, nil, "")
	require.True(t, errors.Is(err, indexer.ErrHeaderTypeAssertion))
}

func TestBlockProcessor_SerializeEpochInfoData(t *testing.T) {
	t.Parallel()

	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	buffSlice := data.NewBufferSlice(data.DefaultBulkSizeThreshold)
	err := bp.SerializeEpochInfoData(&dataBlock.MetaBlock{
		AccumulatedFeesInEpoch: big.NewInt(1),
		DevFeesInEpoch:         big.NewInt(2),
	}, buffSlice, "epochinfo")
	require.Nil(t, err)
	require.Equal(t, `{ "index" : { "_index":"epochinfo", "_id" : "0" } }
{"accumulatedFees":"1","developerFees":"2"}
`, buffSlice.Buffers()[0].String())
}

func TestBlockProcessor_SerializeBlockEpochStartMeta(t *testing.T) {
	t.Parallel()

	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	buffSlice := data.NewBufferSlice(data.DefaultBulkSizeThreshold)
	err := bp.SerializeBlock(&data.Block{
		Nonce:                 1,
		Round:                 2,
		Epoch:                 3,
		MiniBlocksHashes:      []string{"mb1Hash", "mbHash2"},
		NotarizedBlocksHashes: []string{"notarized1"},
		Proposer:              5,
		Validators:            []uint64{0, 1, 2, 3, 4, 5},
		TxCount:               100,
		NotarizedTxsCount:     120,
		PubKeyBitmap:          "00000110",
		Timestamp:             123456,
		StateRootHash:         "stateHash",
		PrevHash:              "prevHash",
		AccumulatedFees:       "1000",
		DeveloperFees:         "50",
		Hash:                  "11cb2a3a28522a11ae646a93aa4d50f87194cead7d6edeb333d502349407b61d",
		Size:                  345,
		ShardID:               core.MetachainShardId,
		EpochStartBlock:       true,
		SearchOrder:           0x3f2,
		EpochStartInfo: &data.EpochStartInfo{
			TotalSupply:                      "100",
			TotalToDistribute:                "55",
			TotalNewlyMinted:                 "20",
			RewardsPerBlock:                  "15",
			RewardsForProtocolSustainability: "2",
			NodePrice:                        "10",
			PrevEpochStartRound:              222,
			PrevEpochStartHash:               "7072657645706f6368",
		},
	}, buffSlice, "blocks")
	require.Nil(t, err)
	require.Equal(t, `{ "index" : { "_index":"blocks", "_id" : "11cb2a3a28522a11ae646a93aa4d50f87194cead7d6edeb333d502349407b61d" } }
{"nonce":1,"round":2,"epoch":3,"miniBlocksHashes":["mb1Hash","mbHash2"],"notarizedBlocksHashes":["notarized1"],"proposer":5,"validators":[0,1,2,3,4,5],"pubKeyBitmap":"00000110","size":345,"sizeTxs":0,"timestamp":123456,"stateRootHash":"stateHash","prevHash":"prevHash","shardId":4294967295,"txCount":100,"notarizedTxsCount":120,"accumulatedFees":"1000","developerFees":"50","epochStartBlock":true,"searchOrder":1010,"epochStartInfo":{"totalSupply":"100","totalToDistribute":"55","totalNewlyMinted":"20","rewardsPerBlock":"15","rewardsForProtocolSustainability":"2","nodePrice":"10","prevEpochStartRound":222,"prevEpochStartHash":"7072657645706f6368"},"gasProvided":0,"gasRefunded":0,"gasPenalized":0,"maxGasLimit":0}
`, buffSlice.Buffers()[0].String())
}
