package integrationtests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-es-indexer-go/client"
	"github.com/multiversx/mx-chain-es-indexer-go/client/logging"
	"github.com/multiversx/mx-chain-es-indexer-go/config"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataprocessor/factory"
	"github.com/multiversx/mx-chain-es-indexer-go/process/outputs/elastic"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var (
	// nolint
	log                = logger.GetOrCreate("integration-tests")
	pubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, addressPrefix)
)

// nolint
func setLogLevelDebug() {
	_ = logger.SetLogLevel("process:DEBUG")
}

// nolint
func createESClient(url string) (elastic.DatabaseClientHandler, error) {
	return client.NewElasticClient(elasticsearch.Config{
		Addresses: []string{url},
		Logger:    &logging.CustomLogger{},
	})
}

// nolint
func decodeAddress(address string) []byte {
	decoded, err := pubKeyConverter.Decode(address)
	log.LogIfError(err, "address", address)

	return decoded
}

// CreateElasticProcessor creates the data processor that has the Elasticsearch output as the only output
func CreateElasticProcessor(
	esClient elastic.DatabaseClientHandler,
) (dataindexer.DataProcessor, error) {
	return factory.CreateDataProcessor(createDataProcessorArgs(esClient))
}

// CreateElasticOutput creates the Elasticsearch output, it should be used by the tests that need only a step of the revert
func CreateElasticOutput(
	esClient elastic.DatabaseClientHandler,
) (elastic.OutputHandler, error) {
	return factory.CreateElasticOutput(createDataProcessorArgs(esClient))
}

func createDataProcessorArgs(esClient elastic.DatabaseClientHandler) factory.ArgDataProcessorFactory {
	return factory.ArgDataProcessorFactory{
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		AddressPubkeyConverter:   pubKeyConverter,
		ValidatorPubkeyConverter: mock.NewPubkeyConverterMock(32),
		DBClient:                 esClient,
		EnabledIndexes: []string{dataindexer.TransactionsIndex, dataindexer.LogsIndex, dataindexer.AccountsESDTIndex, dataindexer.ScResultsIndex,
			dataindexer.ReceiptsIndex, dataindexer.BlockIndex, dataindexer.AccountsIndex, dataindexer.TokensIndex, dataindexer.TagsIndex, dataindexer.EventsIndex,
			dataindexer.OperationsIndex, dataindexer.DelegatorsIndex, dataindexer.ESDTsIndex, dataindexer.SCDeploysIndex, dataindexer.MiniblocksIndex, dataindexer.ValuesIndex, dataindexer.ExecutionResultsIndex},
		Denomination: 18,
		EnableEpochsConfig: config.EnableEpochsConfig{
			RelayedTransactionsV1V2DisableEpoch: 1,
		},
		NumWritesInParallel: 1,
	}
}

// nolint
func readExpectedResult(path string) string {
	jsonFile, _ := os.Open(path)
	byteValue, _ := io.ReadAll(jsonFile)

	return string(byteValue)
}

// nolint
func getElementFromSlice(path string, index int) string {
	fileBytes := readExpectedResult(path)
	slice := make([]map[string]interface{}, 0)
	_ = json.Unmarshal([]byte(fileBytes), &slice)
	res, _ := json.Marshal(slice[index]["_source"])

	return string(res)
}

// nolint
func getIndexMappings(index string) (string, error) {
	u, _ := url.Parse(esURL)
	u.Path = path.Join(u.Path, index, "_mappings")
	res, err := http.Get(u.String())
	if err != nil {
		return "", err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode >= 400 {
		return "", fmt.Errorf("%s", string(body))
	}

	return string(body), nil
}
