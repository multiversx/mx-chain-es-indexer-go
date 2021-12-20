package datafield

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("parse-tests")

var pubKeyConv, _ = pubkeyConverter.NewBech32PubkeyConverter(32, log)

var sender, _ = pubKeyConv.Decode("erd1kqdm94ef5dr9nz3208rrsdzkgwkz53saj4t5chx26cm4hlq8qz8qqd9207")
var receiver, _ = pubKeyConv.Decode("erd1kszzq4egxj5m3t22vt2s8vplmxmqrstghecmnk3tq9mn5fdy7pqqgvzkug")
var receiverSC, _ = pubKeyConv.Decode("erd1qqqqqqqqqqqqqpgqp699jngundfqw07d8jzkepucvpzush6k3wvqyc44rx")

func TestESDTNFTTransfer(t *testing.T) {
	t.Parallel()

	args := &ArgsOperationDataFieldParser{
		PubKeyConverter:  pubKeyConv,
		Marshalizer:      &mock.MarshalizerMock{},
		ShardCoordinator: &mock.ShardCoordinatorMock{},
	}

	parser, _ := NewOperationDataFieldParser(args)

	t.Run("NFTTransferNotOkNonHexArguments", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("ESDTNFTTransfer@@11316@01")
		res := parser.Parse(dataField, sender, receiver)
		require.Equal(t, &ResponseParseData{
			Operation: "transfer",
		}, res)
	})

	t.Run("NFTTransferNotEnoughArguments", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("ESDTNFTTransfer@@1131@01")
		res := parser.Parse(dataField, sender, receiver)
		require.Equal(t, &ResponseParseData{
			Operation: "ESDTNFTTransfer",
		}, res)
	})

	t.Run("NftTransferOk", func(t *testing.T) {
		t.Parallel()

		dataField := []byte("ESDTNFTTransfer@444541442d373966386431@1136@01@08011202000122bc0308b622120c556e646561642023343430361a2000000000000000000500a536e203953414ff92e0a2fdb9b9c0d987fac394242920e8072a2e516d5a39447237447051516b79336e51484a6a4e646b6a393570574c547542384273596a6f4e4c71326262587764324c68747470733a2f2f697066732e696f2f697066732f516d5a39447237447051516b79336e51484a6a4e646b6a393570574c547542384273596a6f4e4c713262625877642f313939302e706e67324d68747470733a2f2f697066732e696f2f697066732f516d5a39447237447051516b79336e51484a6a4e646b6a393570574c547542384273596a6f4e4c713262625877642f313939302e6a736f6e325368747470733a2f2f697066732e696f2f697066732f516d5a39447237447051516b79336e51484a6a4e646b6a393570574c547542384273596a6f4e4c713262625877642f636f6c6c656374696f6e2e6a736f6e3a62746167733a556e646561642c54726561737572652048756e742c456c726f6e643b6d657461646174613a516d5a39447237447051516b79336e51484a6a4e646b6a393570574c547542384273596a6f4e4c713262625877642f313939302e6a736f6e")
		res := parser.Parse(dataField, sender, receiver)
		require.Equal(t, &ResponseParseData{
			Operation:        "ESDTNFTTransfer",
			ESDTValues:       []string{"1"},
			Tokens:           []string{"DEAD-79f8d1-1136"},
			Receivers:        []string{"erd1kszzq4egxj5m3t22vt2s8vplmxmqrstghecmnk3tq9mn5fdy7pqqgvzkug"},
			ReceiversShardID: []uint32{0},
		}, res)
	})

	t.Run("NFTTransferWithSCCallOk", func(t *testing.T) {
		t.Parallel()

		dataField := []byte(`ESDTNFTTransfer@4c4b4641524d2d396431656138@1e47f1@018c88873c27e96447@000000000000000005001e2a1428dd1e3a5146b3960d9e0f4a50369904ee5483@636c61696d5265776172647350726f7879@0000000000000000050026751893d6789be9e5a99863ba9eeaa8088dd25f5483`)
		res := parser.Parse(dataField, sender, sender)
		require.Equal(t, &ResponseParseData{
			Operation:        "ESDTNFTTransfer",
			Function:         "claimRewardsProxy",
			ESDTValues:       []string{"28573236528289506375"},
			Tokens:           []string{"LKFARM-9d1ea8-1e47f1"},
			Receivers:        []string{"erd1qqqqqqqqqqqqqpgqrc4pg2xarca9z34njcxeur622qmfjp8w2jps89fxnl"},
			ReceiversShardID: []uint32{0},
		}, res)
	})
}
