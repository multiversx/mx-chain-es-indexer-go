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

	allTemplates := map[string]*bytes.Buffer{
		indexer.TransactionsIndex:        indices.Transactions.ToBuffer(),
		indexer.BlockIndex:               indices.Blocks.ToBuffer(),
		indexer.MiniblocksIndex:          indices.Miniblocks.ToBuffer(),
		indexer.RatingIndex:              indices.Rating.ToBuffer(),
		indexer.RoundsIndex:              indices.Rounds.ToBuffer(),
		indexer.ValidatorsIndex:          indices.Validators.ToBuffer(),
		indexer.AccountsIndex:            indices.Accounts.ToBuffer(),
		indexer.AccountsHistoryIndex:     indices.AccountsHistory.ToBuffer(),
		indexer.AccountsESDTIndex:        indices.AccountsESDT.ToBuffer(),
		indexer.AccountsESDTHistoryIndex: indices.AccountsESDTHistory.ToBuffer(),
		indexer.EpochInfoIndex:           indices.EpochInfo.ToBuffer(),
		indexer.ReceiptsIndex:            indices.Receipts.ToBuffer(),
		indexer.ScResultsIndex:           indices.SCResults.ToBuffer(),
		indexer.SCDeploysIndex:           indices.SCDeploys.ToBuffer(),
		indexer.TokensIndex:              indices.Tokens.ToBuffer(),
		indexer.TagsIndex:                indices.Tags.ToBuffer(),
		indexer.LogsIndex:                indices.Logs.ToBuffer(),
		indexer.DelegatorsIndex:          indices.Delegators.ToBuffer(),
		indexer.OperationsIndex:          indices.Operations.ToBuffer(),
		indexer.ESDTsIndex:               indices.ESDTs.ToBuffer(),
		indexer.ValuesIndex:              indices.Values.ToBuffer(),
		indexer.EventsIndex:              indices.Events.ToBuffer(),
		indexer.ExecutionResultsIndex:    indices.ExecutionResults.ToBuffer(),
	}

	// Empty availableIndices preserves the legacy behaviour (return everything).
	// In production availableIndices is the already-filtered EnabledIndexes list.
	if len(tr.availableIndices) == 0 {
		return allTemplates, indexPolicies, nil
	}

	for _, index := range tr.availableIndices {
		if template, ok := allTemplates[index]; ok {
			indexTemplates[index] = template
		}
	}

	return indexTemplates, indexPolicies, nil
}

// GetTimestampMsMappings will return the timestampMs field mappings for the enabled indices
func (tr *templatesAndPolicyReader) GetTimestampMsMappings() ([]templates.ExtraMapping, error) {
	allMappings := []templates.ExtraMapping{
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
	}

	if len(tr.availableIndices) == 0 {
		return allMappings, nil
	}

	enabled := make(map[string]struct{}, len(tr.availableIndices))
	for _, index := range tr.availableIndices {
		enabled[index] = struct{}{}
	}

	filtered := make([]templates.ExtraMapping, 0, len(allMappings))
	for _, mapping := range allMappings {
		if _, ok := enabled[mapping.Index]; ok {
			filtered = append(filtered, mapping)
		}
	}

	return filtered, nil
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
