package integrationtests

import (
	"encoding/json"
	"testing"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/client"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elastic-indexer-go/process"
	"github.com/ElrondNetwork/elastic-indexer-go/process/factory"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/stretchr/testify/require"
)

func setLogLevelDebug() {
	_ = logger.SetLogLevel("indexer:DEBUG")
}

func createESClient(url string) (process.DatabaseClientHandler, error) {
	return client.NewElasticClient(elasticsearch.Config{
		Addresses: []string{url},
	})
}

// CreateElasticProcessor -
func CreateElasticProcessor(
	esClient process.DatabaseClientHandler,
	accountsDB indexer.AccountsAdapter,
	shardCoordinator indexer.ShardCoordinator,
	feeProcessor indexer.FeesProcessorHandler,
) (indexer.ElasticProcessor, error) {
	args := factory.ArgElasticProcessorFactory{
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		AddressPubkeyConverter:   mock.NewPubkeyConverterMock(32),
		ValidatorPubkeyConverter: mock.NewPubkeyConverterMock(32),
		DBClient:                 esClient,
		AccountsDB:               accountsDB,
		ShardCoordinator:         shardCoordinator,
		TransactionFeeCalculator: feeProcessor,
		EnabledIndexes:           []string{indexer.TransactionsIndex, indexer.LogsIndex, indexer.AccountsESDTIndex, indexer.ScResultsIndex, indexer.ReceiptsIndex, indexer.BlockIndex, indexer.AccountsIndex, indexer.TokensIndex, indexer.OperationsIndex},
		Denomination:             18,
		IsInImportDBMode:         false,
	}

	return factory.CreateElasticProcessor(args)
}

func compareTxs(t *testing.T, expected []byte, actual []byte) {
	expectedTx := &data.Transaction{}
	err := json.Unmarshal(expected, expectedTx)
	require.Nil(t, err)

	actualTx := &data.Transaction{}
	err = json.Unmarshal(actual, actualTx)
	require.Nil(t, err)

	require.Equal(t, expectedTx, actualTx)
}
