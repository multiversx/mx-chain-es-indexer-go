package outputs

import (
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
)

// Output defines a destination for the data prepared by the indexer (for example Elasticsearch or RabbitMQ)
type Output interface {
	SaveHeader(outportBlockWithHeader *outport.OutportBlockWithHeader) error
	SaveMiniblocks(outportBlockWithHeader *outport.OutportBlockWithHeader) error
	SaveTransactions(block *data.PreparedBlock) error
	RevertBlock(revertedBlock *data.RevertedBlock) error
	SaveValidatorsRating(ratingData *outport.ValidatorsRating) error
	SaveRoundsInfo(rounds *outport.RoundsInfo) error
	SaveShardValidatorsPubKeys(validatorsPubKeys *outport.ValidatorsPubKeys) error
	SaveAccounts(accounts *outport.Accounts) error
	Close() error
	IsInterfaceNil() bool
}

// OutputWithPolicy holds an output together with its name and with the way an error from it is handled
type OutputWithPolicy struct {
	// Name is used in logs and in the returned errors
	Name string
	// Output is the destination of the data
	Output Output
	// ContinueOnError should be set to true if an error from this output should only be logged.
	// If it is false, the error is returned and the node will send the block again
	ContinueOnError bool
}
