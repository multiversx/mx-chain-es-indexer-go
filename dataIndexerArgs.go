package indexer

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
)

// ArgDataIndexer is struct that is used to store all components that are needed to create a indexer
type ArgDataIndexer struct {
	ShardCoordinator Coordinator
	Marshalizer      marshal.Marshalizer
	UseKibana        bool
	DataDispatcher   DispatcherHandler
	ElasticProcessor ElasticProcessor
}

// ArgElasticProcessor is struct that is used to store all components that are needed to an elastic indexer
type ArgElasticProcessor struct {
	IndexTemplates           map[string]*bytes.Buffer
	IndexPolicies            map[string]*bytes.Buffer
	Marshalizer              marshal.Marshalizer
	Hasher                   hashing.Hasher
	AddressPubkeyConverter   core.PubkeyConverter
	ValidatorPubkeyConverter core.PubkeyConverter
	UseKibana                bool
	DBClient                 DatabaseClientHandler
	EnabledIndexes           map[string]struct{}
	AccountsDB               AccountsAdapter
	Denomination             int
	TransactionFeeCalculator FeesProcessorHandler
	IsInImportDBMode         bool
	ShardCoordinator         Coordinator
}
