package validators

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
)

// SerializeValidatorsRating will serialize validators rating
func (vp *validatorsProcessor) SerializeValidatorsRating(ratingData *outport.ValidatorsRating) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice(vp.bulkSizeMaxSize)

	for _, ratingInfo := range ratingData.ValidatorsRatingInfo {
		id := fmt.Sprintf("%s_%d", ratingInfo.PublicKey, ratingData.Epoch)
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, id, "\n"))

		validatorRatingInfo := &data.ValidatorRatingInfo{
			PublicKey: ratingInfo.PublicKey,
			Rating:    ratingInfo.Rating,
		}
		serializedData, err := json.Marshal(validatorRatingInfo)
		if err != nil {
			continue
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}
