package logsevents

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestDelegatorsProcessor_ProcessEvent(t *testing.T) {
	t.Parallel()

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(delegateFunc),
		Topics:     [][]byte{big.NewInt(1000).Bytes(), big.NewInt(1000000000).Bytes(), big.NewInt(10).Bytes(), big.NewInt(1000000000).Bytes()},
	}
	args := &argsProcessEvent{
		timestamp:  1234,
		event:      event,
		logAddress: []byte("contract"),
	}

	balanceConverter, _ := converters.NewBalanceConverter(10)
	delegatorsProcessor := newDelegatorsProcessor(&mock.PubkeyConverterMock{}, balanceConverter)

	res := delegatorsProcessor.processEvent(args)
	require.True(t, res.processed)
	require.Equal(t, &data.Delegator{
		Address:        "61646472",
		Contract:       "636f6e7472616374",
		ActiveStakeNum: 0.1,
		ActiveStake:    "1000000000",
	}, res.delegator)
}
