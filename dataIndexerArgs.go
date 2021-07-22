package indexer

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
)

// ArgDataIndexer is struct that is used to store all components that are needed to create a indexer
type ArgDataIndexer struct {
	ShardCoordinator sharding.Coordinator
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
	AccountsDB               state.AccountsAdapter
	Denomination             int
	TransactionFeeCalculator process.TransactionFeeCalculator
	IsInImportDBMode         bool
	ShardCoordinator         sharding.Coordinator
}
