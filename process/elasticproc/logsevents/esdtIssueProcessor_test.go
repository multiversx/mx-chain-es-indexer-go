package logsevents

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIssueESDTProcessor(t *testing.T) {
	t.Parallel()

	esdtIssueProc := newESDTIssueProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(issueNonFungibleESDTFunc),
		Topics:     [][]byte{[]byte("MYTOKEN-abcd"), []byte("my-token"), []byte("MYTOKEN"), []byte(core.NonFungibleESDT)},
	}
	args := &argsProcessEvent{
		timestamp:   1234,
		event:       event,
		selfShardID: core.MetachainShardId,
		timestampMs: 1234000,
	}

	res := esdtIssueProc.processEvent(args)

	require.Equal(t, &data.TokenInfo{
		Token:        "MYTOKEN-abcd",
		Name:         "my-token",
		Ticker:       "MYTOKEN",
		Timestamp:    1234,
		TimestampMs:  1234000,
		Type:         core.NonFungibleESDT,
		Issuer:       "61646472",
		CurrentOwner: "61646472",
		OwnersHistory: []*data.OwnerData{
			{
				Address:     "61646472",
				Timestamp:   1234,
				TimestampMs: 1234000,
			},
		},
		Properties: &data.TokenProperties{},
	}, res.tokenInfo)
}

func TestIssueESDTProcessor_TransferOwnership(t *testing.T) {
	t.Parallel()

	esdtIssueProc := newESDTIssueProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(transferOwnershipFunc),
		Topics:     [][]byte{[]byte("MYTOKEN-abcd"), []byte("my-token"), []byte("MYTOKEN"), []byte(core.NonFungibleESDT), []byte("newOwner")},
	}
	args := &argsProcessEvent{
		timestamp:   1234,
		timestampMs: 1234000,
		event:       event,
		selfShardID: core.MetachainShardId,
	}

	res := esdtIssueProc.processEvent(args)

	require.Equal(t, &data.TokenInfo{
		Token:        "MYTOKEN-abcd",
		Name:         "my-token",
		Ticker:       "MYTOKEN",
		Timestamp:    1234,
		TimestampMs:  1234000,
		Type:         core.NonFungibleESDT,
		Issuer:       "61646472",
		CurrentOwner: "6e65774f776e6572",
		OwnersHistory: []*data.OwnerData{
			{
				Address:     "6e65774f776e6572",
				Timestamp:   1234,
				TimestampMs: 1234000,
			},
		},
		TransferOwnership: true,
		Properties:        &data.TokenProperties{},
	}, res.tokenInfo)
}

func TestIssueESDTProcessor_EventWithShardID0ShouldBeIgnored(t *testing.T) {
	t.Parallel()

	esdtIssueProc := newESDTIssueProcessor(&mock.PubkeyConverterMock{})

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(transferOwnershipFunc),
		Topics:     [][]byte{[]byte("MYTOKEN-abcd"), []byte("my-token"), []byte("MYTOKEN"), []byte(core.NonFungibleESDT), []byte("newOwner")},
	}
	args := &argsProcessEvent{
		timestamp:   1234,
		timestampMs: 1234000,
		event:       event,
		selfShardID: 0,
	}

	res := esdtIssueProc.processEvent(args)
	require.False(t, res.processed)
}
