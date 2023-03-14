package workItems

import (
	"github.com/multiversx/mx-chain-core-go/data/outport"
)

type itemRounds struct {
	indexer saveRounds
	rounds  *outport.RoundsInfo
}

// NewItemRounds will create a new instance of itemRounds
func NewItemRounds(indexer saveRounds, rounds *outport.RoundsInfo) WorkItemHandler {
	return &itemRounds{
		indexer: indexer,
		rounds:  rounds,
	}
}

// Save will save in elasticsearch database information about rounds
func (wir *itemRounds) Save() error {
	err := wir.indexer.SaveRoundsInfo(wir.rounds)
	if err != nil {
		log.Warn("itemRounds.Save", "could not index rounds info", err.Error())
		return err
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (wir *itemRounds) IsInterfaceNil() bool {
	return wir == nil
}
