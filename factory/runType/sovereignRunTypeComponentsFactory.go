package runType

import (
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/factory"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/tokens"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/transactions"
)

type sovereignRunTypeComponentsFactory struct {
	mainChainElastic factory.ElasticConfig
	esdtPrefix       string
}

// NewSovereignRunTypeComponentsFactory will return a new instance of sovereign run type components factory
func NewSovereignRunTypeComponentsFactory(mainChainElastic factory.ElasticConfig, esdtPrefix string) *sovereignRunTypeComponentsFactory {
	return &sovereignRunTypeComponentsFactory{
		mainChainElastic: mainChainElastic,
		esdtPrefix:       esdtPrefix,
	}
}

// Create will create the run type components
func (srtcf *sovereignRunTypeComponentsFactory) Create() (*runTypeComponents, error) {
	sovIndexTokensHandler, err := tokens.NewSovereignIndexTokensHandler(srtcf.mainChainElastic, srtcf.esdtPrefix)
	if err != nil {
		return nil, err
	}

	return &runTypeComponents{
		txHashExtractor:    transactions.NewSovereignTxHashExtractor(),
		rewardTxData:       transactions.NewSovereignRewardTxData(),
		indexTokensHandler: sovIndexTokensHandler,
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (srtcf *sovereignRunTypeComponentsFactory) IsInterfaceNil() bool {
	return srtcf == nil
}
