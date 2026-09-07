package templatesAndPolicies

import (
	"testing"

	indexer "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/templates"
	"github.com/stretchr/testify/require"
)

func TestTemplatesAndPolicyReader_GetTimestampMsMappings_EmptyReturnsEmpty(t *testing.T) {
	t.Parallel()

	for _, available := range [][]string{nil, {}} {
		reader := NewTemplatesAndPolicyReader(false, "", available, nil)

		mappings, err := reader.GetTimestampMsMappings()
		require.Nil(t, err)
		require.Empty(t, mappings)
	}
}

func TestTemplatesAndPolicyReader_GetTimestampMsMappings_ReturnsAllWhenAllProvided(t *testing.T) {
	t.Parallel()

	available := []string{
		indexer.TransactionsIndex,
		indexer.BlockIndex,
		indexer.MiniblocksIndex,
		indexer.RoundsIndex,
		indexer.AccountsIndex,
		indexer.AccountsESDTIndex,
		indexer.AccountsHistoryIndex,
		indexer.AccountsESDTHistoryIndex,
		indexer.ReceiptsIndex,
		indexer.ScResultsIndex,
		indexer.LogsIndex,
		indexer.OperationsIndex,
		indexer.EventsIndex,
		indexer.TokensIndex,
		indexer.ESDTsIndex,
		indexer.DelegatorsIndex,
		indexer.SCDeploysIndex,
	}
	reader := NewTemplatesAndPolicyReader(false, "", available, nil)

	mappings, err := reader.GetTimestampMsMappings()
	require.Nil(t, err)
	require.Len(t, mappings, 17)

	indexes := collectMappingIndexes(mappings)
	require.Contains(t, indexes, indexer.TransactionsIndex)
	require.Contains(t, indexes, indexer.ScResultsIndex)
	require.Contains(t, indexes, indexer.LogsIndex)
	require.Contains(t, indexes, indexer.OperationsIndex)
	require.Contains(t, indexes, indexer.EventsIndex)
	for _, m := range mappings {
		require.NotNil(t, m.Mappings)
		require.NotEmpty(t, m.Index)
	}
}

func TestTemplatesAndPolicyReader_GetTimestampMsMappings_FiltersDisabledIndices(t *testing.T) {
	t.Parallel()

	// Simulates EnabledIndexes = AvailableIndices - ["transactions", "scresults", "logs"]
	available := []string{
		indexer.BlockIndex,
		indexer.MiniblocksIndex,
		indexer.OperationsIndex,
		indexer.EventsIndex,
		indexer.TokensIndex,
	}
	reader := NewTemplatesAndPolicyReader(false, "", available, nil)

	mappings, err := reader.GetTimestampMsMappings()
	require.Nil(t, err)
	require.Len(t, mappings, len(available))
	require.Equal(t,
		[]string{
			indexer.BlockIndex,
			indexer.MiniblocksIndex,
			indexer.OperationsIndex,
			indexer.EventsIndex,
			indexer.TokensIndex,
		},
		collectMappingIndexes(mappings),
	)
}

func TestTemplatesAndPolicyReader_GetTimestampMsMappings_NoTimestampIndicesReturnsEmpty(t *testing.T) {
	t.Parallel()

	// rating, validators, epochinfo, tags, values and executionresults have no timestampMs mapping
	reader := NewTemplatesAndPolicyReader(false, "", []string{
		indexer.RatingIndex,
		indexer.ValidatorsIndex,
		indexer.EpochInfoIndex,
		indexer.TagsIndex,
		indexer.ValuesIndex,
		indexer.ExecutionResultsIndex,
	}, nil)

	mappings, err := reader.GetTimestampMsMappings()
	require.Nil(t, err)
	require.Empty(t, mappings)
}

func TestTemplatesAndPolicyReader_GetTimestampMsMappings_UnknownIndicesAreIgnored(t *testing.T) {
	t.Parallel()

	reader := NewTemplatesAndPolicyReader(false, "", []string{
		indexer.BlockIndex,
		"unknown-index",
	}, nil)

	mappings, err := reader.GetTimestampMsMappings()
	require.Nil(t, err)
	require.Equal(t, []string{indexer.BlockIndex}, collectMappingIndexes(mappings))
}

func collectMappingIndexes(mappings []templates.ExtraMapping) []string {
	indexes := make([]string, 0, len(mappings))
	for _, m := range mappings {
		indexes = append(indexes, m.Index)
	}
	return indexes
}
