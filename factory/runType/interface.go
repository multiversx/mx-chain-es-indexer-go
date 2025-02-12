package runType

import (
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/transactions"
)

// RunTypeComponentsCreator is the interface for creating run type components
type RunTypeComponentsCreator interface {
	Create() (*runTypeComponents, error)
	IsInterfaceNil() bool
}

// ComponentHandler defines the actions common to all component handlers
type ComponentHandler interface {
	Create() error
	Close() error
	CheckSubcomponents() error
	String() string
}

// RunTypeComponentsHandler defines the run type components handler actions
type RunTypeComponentsHandler interface {
	ComponentHandler
	RunTypeComponentsHolder
}

// RunTypeComponentsHolder holds the run type components
type RunTypeComponentsHolder interface {
	TxHashExtractorCreator() transactions.TxHashExtractor
	RewardTxDataCreator() transactions.RewardTxDataHandler
	IndexTokensHandlerCreator() elasticproc.IndexTokensHandler
	Create() error
	Close() error
	CheckSubcomponents() error
	String() string
	IsInterfaceNil() bool
}
