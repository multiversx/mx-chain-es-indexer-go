package logsevents

import (
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
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

func TestSerializeTokens(t *testing.T) {
	t.Parallel()

	tok1 := &data.TokenInfo{
		Name:      "TokenName",
		Ticker:    "TKN",
		Token:     "TKN-01234",
		Timestamp: 50000,
		Issuer:    "erd123",
		Type:      core.SemiFungibleESDT,
	}
	tok2 := &data.TokenInfo{
		Name:      "Token2",
		Ticker:    "TKN2",
		Token:     "TKN2-51234",
		Issuer:    "erd1231213123",
		Timestamp: 60000,
		Type:      core.NonFungibleESDT,
	}
	tokens := []*data.TokenInfo{tok1, tok2}

	res, err := (&logsAndEventsProcessor{}).SerializeTokens(tokens)
	require.Nil(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "index" : { "_id" : "TKN-01234" } }
{"name":"TokenName","ticker":"TKN","token":"TKN-01234","issuer":"erd123","type":"SemiFungibleESDT","timestamp":50000}
{ "index" : { "_id" : "TKN2-51234" } }
{"name":"Token2","ticker":"TKN2","token":"TKN2-51234","issuer":"erd1231213123","type":"NonFungibleESDT","timestamp":60000}
`
	require.Equal(t, expectedRes, res[0].String())
}

func TestLogsAndEventsProcessor_SerializeDelegators(t *testing.T) {
	t.Parallel()

	delegator1 := &data.Delegator{
		Address:        "addr1",
		Contract:       "contract1",
		ActiveStake:    "100000000000000",
		ActiveStakeNum: 0.1,
	}

	delegators := map[string]*data.Delegator{
		"key1": delegator1,
	}

	logsProc := &logsAndEventsProcessor{
		hahser: &mock.HasherMock{},
	}

	res, err := logsProc.SerializeDelegators(delegators)
	require.Nil(t, err)

	expectedRes := `{ "index" : { "_id" : "/GeogJjDjtpxnceK9t6+BVBYWuuJHbjmsWK0/1BlH9c=" } }
{"address":"addr1","contract":"contract1","activeStake":"100000000000000","activeStakeNum":0.1}
`
	require.Equal(t, expectedRes, res[0].String())
}
