package integrationtests

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/ElrondNetwork/elastic-indexer-go/client"
	"github.com/ElrondNetwork/elastic-indexer-go/client/logging"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elastic-indexer-go/process/dataindexer"
	"github.com/ElrondNetwork/elastic-indexer-go/process/elasticproc"
	"github.com/ElrondNetwork/elastic-indexer-go/process/elasticproc/factory"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/elastic/go-elasticsearch/v7"
)

const esURL = "http://localhost:9200"

func setLogLevelDebug() {
	_ = logger.SetLogLevel("process:DEBUG")
}

func createESClient(url string) (elasticproc.DatabaseClientHandler, error) {
	return client.NewElasticClient(elasticsearch.Config{
		Addresses: []string{url},
		Logger:    &logging.CustomLogger{},
	})
}

// CreateElasticProcessor -
func CreateElasticProcessor(
	esClient elasticproc.DatabaseClientHandler,
	shardCoordinator dataindexer.ShardCoordinator,
) (dataindexer.ElasticProcessor, error) {
	args := factory.ArgElasticProcessorFactory{
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		AddressPubkeyConverter:   mock.NewPubkeyConverterMock(32),
		ValidatorPubkeyConverter: mock.NewPubkeyConverterMock(32),
		DBClient:                 esClient,
		ShardCoordinator:         shardCoordinator,
		EnabledIndexes: []string{dataindexer.TransactionsIndex, dataindexer.LogsIndex, dataindexer.AccountsESDTIndex, dataindexer.ScResultsIndex,
			dataindexer.ReceiptsIndex, dataindexer.BlockIndex, dataindexer.AccountsIndex, dataindexer.TokensIndex, dataindexer.TagsIndex, dataindexer.CollectionsIndex,
			dataindexer.OperationsIndex},
		Denomination: 18,
	}

	return factory.CreateElasticProcessor(args)
}

func readExpectedResult(path string) string {
	jsonFile, _ := os.Open(path)
	byteValue, _ := ioutil.ReadAll(jsonFile)

	return string(byteValue)
}

func getElementFromSlice(path string, index int) string {
	fileBytes := readExpectedResult(path)
	slice := make([]map[string]interface{}, 0)
	_ = json.Unmarshal([]byte(fileBytes), &slice)
	res, _ := json.Marshal(slice[index]["_source"])

	return string(res)
}
