package validators

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ElrondNetwork/elastic-indexer-go/buff"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

// SerializeValidatorsPubKeys will serialize validators public keys
func (vp *validatorsProcessor) SerializeValidatorsPubKeys(validatorsPubKeys *data.ValidatorsPublicKeys) (*bytes.Buffer, error) {
	marshalizedValidatorPubKeys, err := json.Marshal(validatorsPubKeys)
	if err != nil {
		return nil, err
	}

	buffer := &bytes.Buffer{}
	buffer.Grow(len(marshalizedValidatorPubKeys))
	_, err = buffer.Write(marshalizedValidatorPubKeys)
	if err != nil {
		return nil, err
	}

	return buffer, nil
}

// SerializeValidatorsRating will serialize validators rating
func (vp *validatorsProcessor) SerializeValidatorsRating(
	index string,
	validatorsRatingInfo []*data.ValidatorRatingInfo,
) ([]*bytes.Buffer, error) {
	buffSlice := buff.NewBufferSlice()

	// inside elrond-go, the index is "shardID_epoch" so in order to keep backwards compatibility some adjustments have to be made.
	// shardID from index name has to be removed because it is sufficient to have document id = blsKey_epoch
	indexWithoutShardID := removeShardIDFromIndex(index)
	for _, valRatingInfo := range validatorsRatingInfo {
		id := fmt.Sprintf("%s_%s", valRatingInfo.PublicKey, indexWithoutShardID)
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, id, "\n"))

		serializedData, err := json.Marshal(valRatingInfo)
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

func removeShardIDFromIndex(index string) string {
	splitIndex := strings.Split(index, "_")
	if len(splitIndex) == 2 {
		return splitIndex[1]
	}

	return index
}
