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
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-es-indexer-go/client"
	"github.com/multiversx/mx-chain-es-indexer-go/client/logging"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/factory"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/tokens"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/transactions"
)

var (
	// nolint
	log                = logger.GetOrCreate("integration-tests")
	pubKeyConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, addressPrefix)
	sovEsdtPrefix      = "sov"
)

// nolint
func setLogLevelDebug() {
	_ = logger.SetLogLevel("process:DEBUG")
}

// nolint
func createESClient(url string) (elasticproc.DatabaseClientHandler, error) {
	return client.NewElasticClient(elasticsearch.Config{
		Addresses: []string{url},
		Logger:    &logging.CustomLogger{},
	})
}

// nolint
func createMainChainESClient(url string, enabled bool) (elasticproc.MainChainDatabaseClientHandler, error) {
	esClient, _ := createESClient(url)
	return client.NewMainChainElasticClient(esClient, enabled)
}

// nolint
func decodeAddress(address string) []byte {
	decoded, err := pubKeyConverter.Decode(address)
	log.LogIfError(err, "address", address)

	return decoded
}

// CreateElasticProcessor -
func CreateElasticProcessor(
	esClient elasticproc.DatabaseClientHandler,
) (dataindexer.ElasticProcessor, error) {
	args := factory.ArgElasticProcessorFactory{
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		AddressPubkeyConverter:   pubKeyConverter,
		ValidatorPubkeyConverter: mock.NewPubkeyConverterMock(32),
		DBClient:                 esClient,
		EnabledIndexes: []string{dataindexer.TransactionsIndex, dataindexer.LogsIndex, dataindexer.AccountsESDTIndex, dataindexer.ScResultsIndex,
			dataindexer.ReceiptsIndex, dataindexer.BlockIndex, dataindexer.AccountsIndex, dataindexer.TokensIndex, dataindexer.TagsIndex, dataindexer.EventsIndex,
			dataindexer.OperationsIndex, dataindexer.DelegatorsIndex, dataindexer.ESDTsIndex, dataindexer.SCDeploysIndex, dataindexer.MiniblocksIndex, dataindexer.ValuesIndex},
		Denomination:       18,
		TxHashExtractor:    transactions.NewTxHashExtractor(),
		RewardTxData:       transactions.NewRewardTxData(),
		IndexTokensHandler: tokens.NewDisabledIndexTokensHandler(),
	}

	return factory.CreateElasticProcessor(args)
}

// CreateSovereignElasticProcessor -
func CreateSovereignElasticProcessor(
	esClient elasticproc.DatabaseClientHandler,
	mainEsClient elasticproc.MainChainDatabaseClientHandler,
) (dataindexer.ElasticProcessor, error) {
	sovIndexTokens, _ := tokens.NewSovereignIndexTokensHandler(mainEsClient, sovEsdtPrefix)

	args := factory.ArgElasticProcessorFactory{
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		AddressPubkeyConverter:   pubKeyConverter,
		ValidatorPubkeyConverter: mock.NewPubkeyConverterMock(32),
		DBClient:                 esClient,
		EnabledIndexes: []string{dataindexer.TransactionsIndex, dataindexer.LogsIndex, dataindexer.AccountsESDTIndex, dataindexer.ScResultsIndex,
			dataindexer.ReceiptsIndex, dataindexer.BlockIndex, dataindexer.AccountsIndex, dataindexer.TokensIndex, dataindexer.TagsIndex, dataindexer.EventsIndex,
			dataindexer.OperationsIndex, dataindexer.DelegatorsIndex, dataindexer.ESDTsIndex, dataindexer.SCDeploysIndex, dataindexer.MiniblocksIndex, dataindexer.ValuesIndex},
		Denomination:       18,
		TxHashExtractor:    transactions.NewSovereignTxHashExtractor(),
		RewardTxData:       transactions.NewSovereignRewardTxData(),
		IndexTokensHandler: sovIndexTokens,
	}

	return factory.CreateElasticProcessor(args)
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

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode >= 400 {
		return "", fmt.Errorf("%s", string(body))
	}

	return string(body), nil
}
