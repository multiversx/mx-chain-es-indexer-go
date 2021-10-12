package logsevents

const (
	delegateFunc          = "delegate"
	unDelegateFunc        = "unDelegate"
	withdrawFunc          = "withdraw"
	reDelegateRewardsFunc = "reDelegateRewards"
)

type delegatorsProc struct {
	delegatorsOperations map[string]struct{}
}

func newDelegatorsProcessor() *delegatorsProc {
	return &delegatorsProc{
		delegatorsOperations: map[string]struct{}{
			delegateFunc:          {},
			unDelegateFunc:        {},
			withdrawFunc:          {},
			reDelegateRewardsFunc: {},
		},
	}
}

func (dp *delegatorsProc) processEvent(args *argsProcessEvent) argOutputProcessEvent {
	eventIdentifierStr := string(args.event.GetIdentifier())
	_, ok := dp.delegatorsOperations[eventIdentifierStr]
	if !ok {
		return argOutputProcessEvent{}
	}

	return argOutputProcessEvent{}
}
