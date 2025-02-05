package runType

import (
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/tokens"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/transactions"
)

type runTypeComponentsFactory struct{}

// NewRunTypeComponentsFactory will return a new instance of run type components factory
func NewRunTypeComponentsFactory() *runTypeComponentsFactory {
	return &runTypeComponentsFactory{}
}

// Create will create the run type components
func (rtcf *runTypeComponentsFactory) Create() (*runTypeComponents, error) {
	return &runTypeComponents{
		txHashExtractor:    transactions.NewTxHashExtractor(),
		rewardTxData:       transactions.NewRewardTxData(),
		indexTokensHandler: tokens.NewDisabledIndexTokensHandler(),
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rtcf *runTypeComponentsFactory) IsInterfaceNil() bool {
	return rtcf == nil
}
