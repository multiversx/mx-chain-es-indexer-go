package logsevents

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestIssueESDTProcessor(t *testing.T) {
	t.Parallel()

	issueESDTProc := newIssueESDTProcessor()

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(issueNonFungibleESDTFunc),
		Topics:     [][]byte{[]byte("MYTOKEN-abcd"), []byte("my-token"), []byte("MYTOKEN"), []byte(core.NonFungibleESDT)},
	}
	args := &argsProcessEvent{
		timestamp: 1234,
		event:     event,
	}

	res := issueESDTProc.processEvent(args)

	require.Equal(t, &data.TokenInfo{
		Token:     "MYTOKEN-abcd",
		Name:      "my-token",
		Ticker:    "MYTOKEN",
		Timestamp: time.Duration(1234),
		Type:      core.NonFungibleESDT,
	}, res.tokenInfo)
}
