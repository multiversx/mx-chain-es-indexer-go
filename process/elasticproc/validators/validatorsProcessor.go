package validators

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

type validatorsProcessor struct {
	bulkSizeMaxSize          int
	validatorPubkeyConverter core.PubkeyConverter
}

// NewValidatorsProcessor will create a new instance of validatorsProcessor
func NewValidatorsProcessor(validatorPubkeyConverter core.PubkeyConverter, bulkSizeMaxSize int) (*validatorsProcessor, error) {
	if check.IfNil(validatorPubkeyConverter) {
		return nil, dataindexer.ErrNilPubkeyConverter
	}

	return &validatorsProcessor{
		bulkSizeMaxSize:          bulkSizeMaxSize,
		validatorPubkeyConverter: validatorPubkeyConverter,
	}, nil
}

// PrepareValidatorsPublicKeys will prepare validators public keys
func (vp *validatorsProcessor) PrepareValidatorsPublicKeys(shardValidatorsPubKeys [][]byte) *data.ValidatorsPublicKeys {
	validatorsPubKeys := &data.ValidatorsPublicKeys{
		PublicKeys: make([]string, 0),
	}

	for _, validatorPk := range shardValidatorsPubKeys {
		strValidatorPk := vp.validatorPubkeyConverter.Encode(validatorPk)

		validatorsPubKeys.PublicKeys = append(validatorsPubKeys.PublicKeys, strValidatorPk)
	}

	return validatorsPubKeys
}
