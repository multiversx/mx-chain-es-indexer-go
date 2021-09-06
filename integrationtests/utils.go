package integrationtests

import (
	"encoding/json"
	"fmt"
	"testing"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elastic-indexer-go/process"
	"github.com/ElrondNetwork/elastic-indexer-go/process/factory"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/require"
)

func setLogLevelDebug() {
	err := logger.SetLogLevel("indexer:DEBUG")
	if err != nil {
		fmt.Printf("cannot set log level: error %s", err.Error())
	}
}

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
		EnabledIndexes:           []string{indexer.TransactionsIndex, indexer.LogsIndex, indexer.AccountsESDTIndex, indexer.ScResultsIndex, indexer.ReceiptsIndex, indexer.BlockIndex, indexer.AccountsIndex},
		Denomination:             18,
		IsInImportDBMode:         false,
	}

	return factory.CreateElasticProcessor(args)
}

func compareTxs(t *testing.T, expected string, actual string) {
	expectedTx := &data.Transaction{}
	err := json.Unmarshal([]byte(expected), expectedTx)
	require.Nil(t, err)

	actualTx := &data.Transaction{}
	err = json.Unmarshal([]byte(actual), actualTx)
	require.Nil(t, err)

	require.Equal(t, expectedTx, actualTx)
}
