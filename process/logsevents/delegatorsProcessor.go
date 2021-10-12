package logsevents

import (
	"math/big"

	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
)

const (
	minNumTopicsDelegators = 4

	delegateFunc          = "delegate"
	unDelegateFunc        = "unDelegate"
	withdrawFunc          = "withdraw"
	reDelegateRewardsFunc = "reDelegateRewards"
)

type delegatorsProc struct {
	balanceConverter     converters.BalanceConverter
	pubkeyConverter      core.PubkeyConverter
	delegatorsOperations map[string]struct{}
}

func newDelegatorsProcessor(
	pubkeyConverter core.PubkeyConverter,
	balanceConverter converters.BalanceConverter,
) *delegatorsProc {
	return &delegatorsProc{
		delegatorsOperations: map[string]struct{}{
			delegateFunc:          {},
			unDelegateFunc:        {},
			withdrawFunc:          {},
			reDelegateRewardsFunc: {},
		},
		pubkeyConverter:  pubkeyConverter,
		balanceConverter: balanceConverter,
	}
}

func (dp *delegatorsProc) processEvent(args *argsProcessEvent) argOutputProcessEvent {
	eventIdentifierStr := string(args.event.GetIdentifier())
	_, ok := dp.delegatorsOperations[eventIdentifierStr]
	if !ok {
		return argOutputProcessEvent{}
	}

	topics := args.event.GetTopics()
	if len(topics) < minNumTopicsDelegators {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	// for delegate/ unDelegate/ withdraw/ reDelegateRewards
	// topics slice contains:
	// topics[0] = delegated value / unDelegated value / withdraw value/ reDelegated value
	// topics[1] = active stake
	// topics[2] = num contract users
	// topics[3] = total contract active stake
	activeStake := big.NewInt(0).SetBytes(topics[1])

	delegator := &data.Delegator{
		Address:        dp.pubkeyConverter.Encode(args.event.GetAddress()),
		Contract:       dp.pubkeyConverter.Encode(args.logAddress),
		ActiveStake:    activeStake.String(),
		ActiveStakeNum: dp.balanceConverter.ComputeBalanceAsFloat(activeStake),
	}

	return argOutputProcessEvent{
		delegator: delegator,
		processed: true,
	}
}
