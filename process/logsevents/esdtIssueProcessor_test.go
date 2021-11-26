package logsevents

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
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
		timestamp: 1234,
		event:     event,
	}

	res := esdtIssueProc.processEvent(args)

	require.Equal(t, &data.TokenInfo{
		Token:        "MYTOKEN-abcd",
		Name:         "my-token",
		Ticker:       "MYTOKEN",
		Timestamp:    time.Duration(1234),
		Type:         core.NonFungibleESDT,
		Issuer:       "61646472",
		CurrentOwner: "61646472",
		OwnersHistory: []*data.OwnerData{
			{
				Address:   "61646472",
				Timestamp: time.Duration(1234),
			},
		},
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
		timestamp: 1234,
		event:     event,
	}

	res := esdtIssueProc.processEvent(args)

	require.Equal(t, &data.TokenInfo{
		Token:        "MYTOKEN-abcd",
		Name:         "my-token",
		Ticker:       "MYTOKEN",
		Timestamp:    time.Duration(1234),
		Type:         core.NonFungibleESDT,
		Issuer:       "61646472",
		CurrentOwner: "6e65774f776e6572",
		OwnersHistory: []*data.OwnerData{
			{
				Address:   "6e65774f776e6572",
				Timestamp: time.Duration(1234),
			},
		},
		TransferOwnership: true,
	}, res.tokenInfo)
}
