package logsevents

import (
	"math/big"
	"strconv"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
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
		timestamp:   1234,
		event:       event,
		logAddress:  []byte("contract"),
		selfShardID: core.MetachainShardId,
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

func TestDelegatorProcessor_WithdrawWithDelete(t *testing.T) {
	t.Parallel()

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(withdrawFunc),
		Topics:     [][]byte{big.NewInt(1000).Bytes(), big.NewInt(0).Bytes(), big.NewInt(10).Bytes(), big.NewInt(1000000000).Bytes(), []byte(strconv.FormatBool(true))},
	}
	args := &argsProcessEvent{
		timestamp:   1234,
		event:       event,
		logAddress:  []byte("contract"),
		selfShardID: core.MetachainShardId,
	}

	balanceConverter, _ := converters.NewBalanceConverter(10)
	delegatorsProcessor := newDelegatorsProcessor(&mock.PubkeyConverterMock{}, balanceConverter)

	res := delegatorsProcessor.processEvent(args)
	require.True(t, res.processed)
	require.Equal(t, &data.Delegator{
		Address:        "61646472",
		Contract:       "636f6e7472616374",
		ActiveStakeNum: 0,
		ActiveStake:    "0",
		ShouldDelete:   true,
	}, res.delegator)
}

func TestDelegatorProcessor_ClaimRewardsWithDelete(t *testing.T) {
	t.Parallel()

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(claimRewardsFunc),
		Topics:     [][]byte{big.NewInt(1000).Bytes(), []byte(strconv.FormatBool(true))},
	}
	args := &argsProcessEvent{
		timestamp:   1234,
		event:       event,
		logAddress:  []byte("contract"),
		selfShardID: core.MetachainShardId,
	}

	balanceConverter, _ := converters.NewBalanceConverter(10)
	delegatorsProcessor := newDelegatorsProcessor(&mock.PubkeyConverterMock{}, balanceConverter)

	res := delegatorsProcessor.processEvent(args)
	require.True(t, res.processed)
	require.Equal(t, &data.Delegator{
		Address:      "61646472",
		Contract:     "636f6e7472616374",
		ShouldDelete: true,
	}, res.delegator)
}

func TestDelegatorProcessor_ClaimRewardsNoDelete(t *testing.T) {
	t.Parallel()

	event := &transaction.Event{
		Address:    []byte("addr"),
		Identifier: []byte(claimRewardsFunc),
		Topics:     [][]byte{big.NewInt(1000).Bytes(), []byte(strconv.FormatBool(false))},
	}
	args := &argsProcessEvent{
		timestamp:   1234,
		event:       event,
		logAddress:  []byte("contract"),
		selfShardID: core.MetachainShardId,
	}

	balanceConverter, _ := converters.NewBalanceConverter(10)
	delegatorsProcessor := newDelegatorsProcessor(&mock.PubkeyConverterMock{}, balanceConverter)

	res := delegatorsProcessor.processEvent(args)
	require.True(t, res.processed)
	require.Nil(t, res.delegator)
}
