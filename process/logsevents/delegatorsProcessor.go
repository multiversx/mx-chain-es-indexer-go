package logsevents

import (
	"math/big"
	"strconv"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
)

const (
	minNumTopicsDelegators = 4
	delegateFunc           = "delegate"
	unDelegateFunc         = "unDelegate"
	withdrawFunc           = "withdraw"
	reDelegateRewardsFunc  = "reDelegateRewards"
	claimRewardsFunc       = "claimRewards"
)

type delegatorsProc struct {
	balanceConverter     indexer.BalanceConverter
	pubkeyConverter      core.PubkeyConverter
	delegatorsOperations map[string]struct{}
}

func newDelegatorsProcessor(
	pubkeyConverter core.PubkeyConverter,
	balanceConverter indexer.BalanceConverter,
) *delegatorsProc {
	return &delegatorsProc{
		delegatorsOperations: map[string]struct{}{
			delegateFunc:          {},
			unDelegateFunc:        {},
			withdrawFunc:          {},
			reDelegateRewardsFunc: {},
			claimRewardsFunc:      {},
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

	if eventIdentifierStr == claimRewardsFunc {
		return argOutputProcessEvent{
			delegator: dp.getDelegatorFromClaimRewardsEvent(args),
			processed: true,
		}
	}

	topics := args.event.GetTopics()
	if len(topics) < minNumTopicsDelegators {
		return argOutputProcessEvent{
			processed: true,
		}
	}

	// for delegate / unDelegate / withdraw / reDelegateRewards
	// topics slice contains:
	// topics[0] = delegated value / unDelegated value / withdraw value / reDelegated value
	// topics[1] = active stake
	// topics[2] = num contract users
	// topics[3] = total contract active stake
	// topics[4] = true - if delegator was deleted in case of withdrawal OR contract address in case of delegate from async
	activeStake := big.NewInt(0).SetBytes(topics[1])

	contractAddr := dp.pubkeyConverter.Encode(args.logAddress)
	if len(topics) >= minNumTopicsDelegators+1 && eventIdentifierStr == delegateFunc {
		contractAddr = dp.pubkeyConverter.Encode(topics[4])
	}

	delegator := &data.Delegator{
		Address:        dp.pubkeyConverter.Encode(args.event.GetAddress()),
		Contract:       contractAddr,
		ActiveStake:    activeStake.String(),
		ActiveStakeNum: dp.balanceConverter.ComputeBalanceAsFloat(activeStake),
	}

	if eventIdentifierStr == withdrawFunc && len(topics) >= minNumTopicsDelegators+1 {
		delegator.ShouldDelete = bytesToBool(topics[4])
	}

	return argOutputProcessEvent{
		delegator: delegator,
		processed: true,
	}
}

func (dp *delegatorsProc) getDelegatorFromClaimRewardsEvent(args *argsProcessEvent) *data.Delegator {
	topics := args.event.GetTopics()
	// for claimRewards
	// topics slice contains:
	// topics[0] -- claimed rewards
	// topics[1] -- true = if delegator was deleted

	if len(topics) < 2 {
		return nil
	}

	shouldDelete := bytesToBool(topics[1])
	if !shouldDelete {
		return nil
	}

	return &data.Delegator{
		Address:      dp.pubkeyConverter.Encode(args.event.GetAddress()),
		Contract:     dp.pubkeyConverter.Encode(args.logAddress),
		ShouldDelete: shouldDelete,
	}
}

func bytesToBool(boolBytes []byte) bool {
	b, err := strconv.ParseBool(string(boolBytes))
	if err != nil {
		log.Warn("delegatorsProc.bytesToBool", "error", err.Error())
	}

	return b
}
