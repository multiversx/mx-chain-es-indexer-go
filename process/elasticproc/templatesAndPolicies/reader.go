package templatesAndPolicies

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"

	indexer "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/templates"
	"github.com/multiversx/mx-chain-es-indexer-go/templates/indices"
)

const (
	indicesFolder  = "indices"
	policiesFolder = "policies"
)

type templatesAndPolicyReader struct {
	useTemplatesFromFiles bool
	configPath            string
	availableIndices      []string
	indicesWithPolicies   []string
}

// NewTemplatesAndPolicyReader will create a new instance of templatesAndPolicyReader
func NewTemplatesAndPolicyReader(
	useTemplatesFromFiles bool,
	configPath string,
	availableIndices []string,
	indicesWithPolicies []string,
) *templatesAndPolicyReader {
	return &templatesAndPolicyReader{
		useTemplatesFromFiles: useTemplatesFromFiles,
		configPath:            configPath,
		availableIndices:      availableIndices,
		indicesWithPolicies:   indicesWithPolicies,
	}
}

// GetElasticTemplatesAndPolicies will return templates and policies
func (tr *templatesAndPolicyReader) GetElasticTemplatesAndPolicies() (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error) {
	if tr.useTemplatesFromFiles {
		return tr.getElasticTemplatesAndPoliciesFromJsonFiles()
	}

	indexPolicies := make(map[string]*bytes.Buffer)
	indexTemplates := make(map[string]*bytes.Buffer)

	indexTemplates[indexer.TransactionsIndex] = indices.Transactions.ToBuffer()
	indexTemplates[indexer.BlockIndex] = indices.Blocks.ToBuffer()
	indexTemplates[indexer.MiniblocksIndex] = indices.Miniblocks.ToBuffer()
	indexTemplates[indexer.RatingIndex] = indices.Rating.ToBuffer()
	indexTemplates[indexer.RoundsIndex] = indices.Rounds.ToBuffer()
	indexTemplates[indexer.ValidatorsIndex] = indices.Validators.ToBuffer()
	indexTemplates[indexer.AccountsIndex] = indices.Accounts.ToBuffer()
	indexTemplates[indexer.AccountsHistoryIndex] = indices.AccountsHistory.ToBuffer()
	indexTemplates[indexer.AccountsESDTIndex] = indices.AccountsESDT.ToBuffer()
	indexTemplates[indexer.AccountsESDTHistoryIndex] = indices.AccountsESDTHistory.ToBuffer()
	indexTemplates[indexer.EpochInfoIndex] = indices.EpochInfo.ToBuffer()
	indexTemplates[indexer.ReceiptsIndex] = indices.Receipts.ToBuffer()
	indexTemplates[indexer.ScResultsIndex] = indices.SCResults.ToBuffer()
	indexTemplates[indexer.SCDeploysIndex] = indices.SCDeploys.ToBuffer()
	indexTemplates[indexer.TokensIndex] = indices.Tokens.ToBuffer()
	indexTemplates[indexer.TagsIndex] = indices.Tags.ToBuffer()
	indexTemplates[indexer.LogsIndex] = indices.Logs.ToBuffer()
	indexTemplates[indexer.DelegatorsIndex] = indices.Delegators.ToBuffer()
	indexTemplates[indexer.OperationsIndex] = indices.Operations.ToBuffer()
	indexTemplates[indexer.ESDTsIndex] = indices.ESDTs.ToBuffer()
	indexTemplates[indexer.ValuesIndex] = indices.Values.ToBuffer()
	indexTemplates[indexer.EventsIndex] = indices.Events.ToBuffer()
	indexTemplates[indexer.ExecutionResultsIndex] = indices.ExecutionResults.ToBuffer()

	return indexTemplates, indexPolicies, nil
}

// GetTimestampMsMappings will return the timestampMs field mappings for all indices
func (tr *templatesAndPolicyReader) GetTimestampMsMappings() ([]templates.ExtraMapping, error) {
	return []templates.ExtraMapping{
		{
			Index:    indexer.TransactionsIndex,
			Mappings: indices.TimestampMs.ToBuffer(),
		},
		{
			Index:    indexer.BlockIndex,
			Mappings: indices.TimestampMs.ToBuffer(),
		},
		{
			Index:    indexer.MiniblocksIndex,
			Mappings: indices.TimestampMs.ToBuffer(),
		},
		{
			Index:    indexer.RoundsIndex,
			Mappings: indices.TimestampMs.ToBuffer(),
		},
		{
			Index:    indexer.AccountsIndex,
			Mappings: indices.TimestampMs.ToBuffer(),
		},
		{
			Index:    indexer.AccountsESDTIndex,
			Mappings: indices.TimestampMs.ToBuffer(),
		},
		{
			Index:    indexer.AccountsHistoryIndex,
			Mappings: indices.TimestampMs.ToBuffer(),
		},
		{
			Index:    indexer.AccountsESDTHistoryIndex,
			Mappings: indices.TimestampMs.ToBuffer(),
		},
		{
			Index:    indexer.ReceiptsIndex,
			Mappings: indices.TimestampMs.ToBuffer(),
		},

		{
			Index:    indexer.ScResultsIndex,
			Mappings: indices.TimestampMs.ToBuffer(),
		},
		{
			Index:    indexer.LogsIndex,
			Mappings: indices.TimestampMs.ToBuffer(),
		},
		{
			Index:    indexer.OperationsIndex,
			Mappings: indices.TimestampMs.ToBuffer(),
		},
		{
			Index:    indexer.EventsIndex,
			Mappings: indices.TimestampMs.ToBuffer(),
		},
		{
			Index:    indexer.TokensIndex,
			Mappings: indices.TokensTimestampMs.ToBuffer(),
		},
		{
			Index:    indexer.ESDTsIndex,
			Mappings: indices.TokensTimestampMs.ToBuffer(),
		},
		{
			Index:    indexer.DelegatorsIndex,
			Mappings: indices.DelegatorsTimestampMs.ToBuffer(),
		},
		{
			Index:    indexer.SCDeploysIndex,
			Mappings: indices.DeploysTimestampMs.ToBuffer(),
		},
	}, nil
}

// GetExtraMappings will return an array of indices extra mappings
func (tr *templatesAndPolicyReader) GetExtraMappings() ([]templates.ExtraMapping, error) {
	return []templates.ExtraMapping{}, nil
}

func (tr *templatesAndPolicyReader) getElasticTemplatesAndPoliciesFromJsonFiles() (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error) {
	pathToMappings := path.Join(tr.configPath, indicesFolder)
	indicesTemplateMap, err := tr.getElasticTemplatesFromJson(pathToMappings, tr.availableIndices)
	if err != nil {
		return nil, nil, fmt.Errorf("%w, cannot load templates", err)
	}

	pathToPolicies := path.Join(tr.configPath, policiesFolder)
	indicesPolicyMap, err := tr.getElasticTemplatesFromJson(pathToPolicies, tr.indicesWithPolicies)
	if err != nil {
		return nil, nil, fmt.Errorf("%w, cannot load templates", err)
	}

	return indicesTemplateMap, indicesPolicyMap, nil
}

func (tr *templatesAndPolicyReader) getElasticTemplatesFromJson(filePath string, indices []string) (map[string]*bytes.Buffer, error) {
	indexTemplates := make(map[string]*bytes.Buffer)
	var err error

	for _, index := range indices {
		indexTemplates[index], err = getDataFromByIndex(filePath, index)
		if err != nil {
			return nil, err
		}
	}

	return indexTemplates, nil
}

func getDataFromByIndex(path string, index string) (*bytes.Buffer, error) {
	indexTemplate := &bytes.Buffer{}

	fileName := fmt.Sprintf("%s.json", index)
	filePath := filepath.Join(path, fileName)
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("getDataFromByIndex: %w, path %s, error %s", err, filePath, err.Error())
	}

	indexTemplate.Grow(len(fileBytes))
	_, err = indexTemplate.Write(fileBytes)
	if err != nil {
		return nil, fmt.Errorf("getDataFromByIndex: %w, path %s, error %s", err, filePath, err.Error())
	}

	return indexTemplate, nil
}
