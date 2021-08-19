package logsevents

import (
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/stretchr/testify/require"
)

func TestLogsAndEventsProcessor_SerializeLogs(t *testing.T) {
	t.Parallel()

	logs := []*data.Logs{
		{
			ID:        "747848617368",
			Address:   "61646472657373",
			Timestamp: time.Duration(1234),
			Events: []*data.Event{
				{
					Address:    "61646472",
					Identifier: core.BuiltInFunctionESDTNFTTransfer,
					Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).SetUint64(1).Bytes(), []byte("receiver")},
					Data:       []byte("data"),
				},
			},
		},
	}

	res, err := (&logsAndEventsProcessor{}).SerializeLogs(logs)
	require.Nil(t, err)

	expectedRes := `{ "index" : { "_id" : "747848617368" } }
{"address":"61646472657373","events":[{"address":"61646472","identifier":"ESDTNFTTransfer","topics":["bXktdG9rZW4=","AQ==","cmVjZWl2ZXI="],"data":"ZGF0YQ=="}],"timestamp":1234}
`
	require.Equal(t, expectedRes, res[0].String())
}

func TestLogsAndEventsProcessor_SerializeSCDeploys(t *testing.T) {
	t.Parallel()

	scDeploys := map[string]*data.ScDeployInfo{
		"scAddr": {
			Creator:   "creator",
			Timestamp: 123,
			TxHash:    "hash",
		},
	}

	res, err := (&logsAndEventsProcessor{}).SerializeSCDeploys(scDeploys)
	require.Nil(t, err)

	expectedRes := `{ "update" : { "_id" : "scAddr", "_type" : "_doc" } }
{"script": {"source": "if (!ctx._source.containsKey('upgrades')) { ctx._source.upgrades = [ params.elem ]; } else {  ctx._source.upgrades.add(params.elem); }","lang": "painless","params": {"elem": {"upgradeTxHash":"hash","upgrader":"creator","timestamp":123}}},"upsert": {"deployTxHash":"hash","deployer":"creator","timestamp":123,"upgrades":[]}}
`
	require.Equal(t, expectedRes, res[0].String())
}
