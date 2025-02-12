package runType

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"

	elasticIndexer "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/transactions"
)

const runTypeComponentsName = "managedRunTypeComponents"

var _ ComponentHandler = (*managedRunTypeComponents)(nil)
var _ RunTypeComponentsHandler = (*managedRunTypeComponents)(nil)
var _ RunTypeComponentsHolder = (*managedRunTypeComponents)(nil)

type managedRunTypeComponents struct {
	*runTypeComponents
	factory                  RunTypeComponentsCreator
	mutRunTypeCoreComponents sync.RWMutex
}

// NewManagedRunTypeComponents returns a news instance of managed runType core components
func NewManagedRunTypeComponents(rtc RunTypeComponentsCreator) (*managedRunTypeComponents, error) {
	if rtc == nil {
		return nil, errNilRunTypeComponents
	}

	return &managedRunTypeComponents{
		runTypeComponents: nil,
		factory:           rtc,
	}, nil
}

// Create will create the managed components
func (mrtc *managedRunTypeComponents) Create() error {
	rtc, err := mrtc.factory.Create()
	if err != nil {
		return err
	}

	mrtc.mutRunTypeCoreComponents.Lock()
	mrtc.runTypeComponents = rtc
	mrtc.mutRunTypeCoreComponents.Unlock()

	return nil
}

// Close will close all underlying subcomponents
func (mrtc *managedRunTypeComponents) Close() error {
	mrtc.mutRunTypeCoreComponents.Lock()
	defer mrtc.mutRunTypeCoreComponents.Unlock()

	if check.IfNil(mrtc.runTypeComponents) {
		return nil
	}

	err := mrtc.runTypeComponents.Close()
	if err != nil {
		return err
	}
	mrtc.runTypeComponents = nil

	return nil
}

// CheckSubcomponents verifies all subcomponents
func (mrtc *managedRunTypeComponents) CheckSubcomponents() error {
	mrtc.mutRunTypeCoreComponents.RLock()
	defer mrtc.mutRunTypeCoreComponents.RUnlock()

	if check.IfNil(mrtc.runTypeComponents) {
		return errNilRunTypeComponents
	}
	if check.IfNil(mrtc.txHashExtractor) {
		return transactions.ErrNilTxHashExtractor
	}
	if check.IfNil(mrtc.rewardTxData) {
		return transactions.ErrNilRewardTxDataHandler
	}
	if check.IfNil(mrtc.indexTokensHandler) {
		return elasticIndexer.ErrNilIndexTokensHandler
	}
	return nil
}

// TxHashExtractorCreator returns tx hash extractor
func (mrtc *managedRunTypeComponents) TxHashExtractorCreator() transactions.TxHashExtractor {
	mrtc.mutRunTypeCoreComponents.Lock()
	defer mrtc.mutRunTypeCoreComponents.Unlock()

	if check.IfNil(mrtc.runTypeComponents) {
		return nil
	}

	return mrtc.runTypeComponents.txHashExtractor
}

// RewardTxDataCreator return reward tx handler
func (mrtc *managedRunTypeComponents) RewardTxDataCreator() transactions.RewardTxDataHandler {
	mrtc.mutRunTypeCoreComponents.Lock()
	defer mrtc.mutRunTypeCoreComponents.Unlock()

	if check.IfNil(mrtc.runTypeComponents) {
		return nil
	}

	return mrtc.runTypeComponents.rewardTxData
}

// IndexTokensHandlerCreator returns the index tokens handler
func (mrtc *managedRunTypeComponents) IndexTokensHandlerCreator() elasticproc.IndexTokensHandler {
	mrtc.mutRunTypeCoreComponents.Lock()
	defer mrtc.mutRunTypeCoreComponents.Unlock()

	if check.IfNil(mrtc.runTypeComponents) {
		return nil
	}

	return mrtc.runTypeComponents.indexTokensHandler
}

// IsInterfaceNil returns true if the interface is nil
func (mrtc *managedRunTypeComponents) IsInterfaceNil() bool {
	return mrtc == nil
}

// String returns the name of the component
func (mrtc *managedRunTypeComponents) String() string {
	return runTypeComponentsName
}
