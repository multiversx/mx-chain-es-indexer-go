package logsevents

import (
	"math/big"
	"testing"

	elasticIndexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go/core"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestNewLogsAndEventsProcessor(t *testing.T) {
	t.Parallel()

	_, err := NewLogsAndEventsProcessor(nil, &mock.PubkeyConverterMock{}, &mock.MarshalizerMock{})
	require.Equal(t, elasticIndexer.ErrNilShardCoordinator, err)

	_, err = NewLogsAndEventsProcessor(&mock.ShardCoordinatorMock{}, nil, &mock.MarshalizerMock{})
	require.Equal(t, elasticIndexer.ErrNilPubkeyConverter, err)

	_, err = NewLogsAndEventsProcessor(&mock.ShardCoordinatorMock{}, &mock.PubkeyConverterMock{}, nil)
	require.Equal(t, elasticIndexer.ErrNilMarshalizer, err)

	proc, err := NewLogsAndEventsProcessor(&mock.ShardCoordinatorMock{}, &mock.PubkeyConverterMock{}, &mock.MarshalizerMock{})
	require.NotNil(t, proc)
	require.Nil(t, err)
}

func TestLogsAndEventsProcessor_PrepareLogsForDB(t *testing.T) {
	t.Parallel()

	logsAndEvents := map[string]nodeData.LogHandler{
		"wrong": nil,

		"txHash": &transaction.Log{
			Address: []byte("address"),
			Events: []*transaction.Event{
				{
					Address:    []byte("addr"),
					Identifier: []byte(core.BuiltInFunctionESDTNFTTransfer),
					Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).SetUint64(1).Bytes(), []byte("receiver")},
				},
			},
		},
	}

	proc, _ := NewLogsAndEventsProcessor(&mock.ShardCoordinatorMock{}, mock.NewPubkeyConverterMock(32), &mock.MarshalizerMock{})

	logsDB := proc.PrepareLogsForDB(logsAndEvents)
	require.Equal(t, &data.Logs{
		ID:      "747848617368",
		Address: "61646472657373",
		Events: []*data.Event{
			{
				Address:    "61646472",
				Identifier: core.BuiltInFunctionESDTNFTTransfer,
				Topics:     [][]byte{[]byte("my-token"), big.NewInt(0).SetUint64(1).Bytes(), []byte("receiver")},
			},
		},
	}, logsDB[0])
}
