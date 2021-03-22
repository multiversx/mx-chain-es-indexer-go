package validators

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

// SerializeValidatorsPubKeys will serialize validators public keys
func (vp *validatorsProcessor) SerializeValidatorsPubKeys(validatorsPubKeys *data.ValidatorsPublicKeys) (*bytes.Buffer, error) {
	marshalizedValidatorPubKeys, err := json.Marshal(validatorsPubKeys)
	if err != nil {
		log.Warn("validatorsProcessor.SerializeValidatorPubKeys cannot marshal", "error", err)

		return nil, err
	}

	buff := &bytes.Buffer{}
	buff.Grow(len(marshalizedValidatorPubKeys))
	_, err = buff.Write(marshalizedValidatorPubKeys)
	if err != nil {
		log.Warn("validatorsProcessor.SerializeValidatorPubKeys cannot save validators public keys", "error", err)

		return nil, err
	}

	return buff, nil
}

// SerializeValidatorsRating will serialize validators rating
func (vp *validatorsProcessor) SerializeValidatorsRating(
	index string,
	validatorsRatingInfo []*data.ValidatorRatingInfo,
) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()

	for _, valRatingInfo := range validatorsRatingInfo {
		id := fmt.Sprintf("%s_%s", valRatingInfo.PublicKey, index)
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, id, "\n"))

		serializedData, err := json.Marshal(valRatingInfo)
		if err != nil {
			log.Warn("validatorsProcessor.SerializeValidatorsRating", "error", err, "publickey", valRatingInfo.PublicKey)
			continue
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			log.Warn("validatorsProcessor.SerializeValidatorsRating",
				"cannot put data in buffer", "error", err, "publickey", valRatingInfo.PublicKey)
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}
