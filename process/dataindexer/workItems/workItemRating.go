package workItems

import (
	"github.com/multiversx/mx-chain-core-go/data/outport"
)

type itemRating struct {
	indexer    saveRatingIndexer
	ratingData *outport.ValidatorsRating
}

// NewItemRating will create a new instance of itemRating
func NewItemRating(indexer saveRatingIndexer, ratingData *outport.ValidatorsRating) WorkItemHandler {
	return &itemRating{
		indexer:    indexer,
		ratingData: ratingData,
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (wir *itemRating) IsInterfaceNil() bool {
	return wir == nil
}

// Save will save validators rating in elasticsearch database
func (wir *itemRating) Save() error {
	err := wir.indexer.SaveValidatorsRating(wir.ratingData)
	if err != nil {
		log.Warn("itemRating.Save", "could not index validators rating", err.Error())
		return err
	}

	return nil
}
