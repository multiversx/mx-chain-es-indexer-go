package validators

import (
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/process"
)

func (vp *validatorsProcessor) ValidatorsRatingToPostgres(
	postgresClient process.PostgresClientHandler,
	index string,
	validatorsRatingInfo []*data.ValidatorRatingInfo,
) error {
	indexWithoutShardID := removeShardIDFromIndex(index)

	for _, valRatingInfo := range validatorsRatingInfo {
		id := fmt.Sprintf("%s_%s", valRatingInfo.PublicKey, indexWithoutShardID)

		err := postgresClient.InsertValidatorsRating(id, valRatingInfo)
		if err != nil {
			return err
		}

	}

	return nil
}
