package templatesAndPolicies

import (
	"bytes"

	indexer "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/templates"
	"github.com/multiversx/mx-chain-es-indexer-go/templates/noKibana"
)

type templatesAndPolicyReaderNoKibana struct{}

// NewTemplatesAndPolicyReaderNoKibana will create a new instance of templatesAndPolicyReaderNoKibana
func NewTemplatesAndPolicyReaderNoKibana() *templatesAndPolicyReaderNoKibana {
	return new(templatesAndPolicyReaderNoKibana)
}

// GetElasticTemplatesAndPolicies will return templates and policies
func (tr *templatesAndPolicyReaderNoKibana) GetElasticTemplatesAndPolicies() (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error) {
	indexPolicies := make(map[string]*bytes.Buffer)
	indexTemplates := make(map[string]*bytes.Buffer)

	indexTemplates["opendistro"] = noKibana.OpenDistro.ToBuffer()
	indexTemplates[indexer.TransactionsIndex] = noKibana.Transactions.ToBuffer()
	indexTemplates[indexer.BlockIndex] = noKibana.Blocks.ToBuffer()
	indexTemplates[indexer.MiniblocksIndex] = noKibana.Miniblocks.ToBuffer()
	indexTemplates[indexer.RatingIndex] = noKibana.Rating.ToBuffer()
	indexTemplates[indexer.RoundsIndex] = noKibana.Rounds.ToBuffer()
	indexTemplates[indexer.ValidatorsIndex] = noKibana.Validators.ToBuffer()
	indexTemplates[indexer.AccountsIndex] = noKibana.Accounts.ToBuffer()
	indexTemplates[indexer.AccountsHistoryIndex] = noKibana.AccountsHistory.ToBuffer()
	indexTemplates[indexer.AccountsESDTIndex] = noKibana.AccountsESDT.ToBuffer()
	indexTemplates[indexer.AccountsESDTHistoryIndex] = noKibana.AccountsESDTHistory.ToBuffer()
	indexTemplates[indexer.EpochInfoIndex] = noKibana.EpochInfo.ToBuffer()
	indexTemplates[indexer.ReceiptsIndex] = noKibana.Receipts.ToBuffer()
	indexTemplates[indexer.ScResultsIndex] = noKibana.SCResults.ToBuffer()
	indexTemplates[indexer.SCDeploysIndex] = noKibana.SCDeploys.ToBuffer()
	indexTemplates[indexer.TokensIndex] = noKibana.Tokens.ToBuffer()
	indexTemplates[indexer.TagsIndex] = noKibana.Tags.ToBuffer()
	indexTemplates[indexer.LogsIndex] = noKibana.Logs.ToBuffer()
	indexTemplates[indexer.DelegatorsIndex] = noKibana.Delegators.ToBuffer()
	indexTemplates[indexer.OperationsIndex] = noKibana.Operations.ToBuffer()
	indexTemplates[indexer.ESDTsIndex] = noKibana.ESDTs.ToBuffer()
	indexTemplates[indexer.ValuesIndex] = noKibana.Values.ToBuffer()
	indexTemplates[indexer.EventsIndex] = noKibana.Events.ToBuffer()

	return indexTemplates, indexPolicies, nil
}

// GetExtraMappings will return an array of indices extra mappings
func (tr *templatesAndPolicyReaderNoKibana) GetExtraMappings() ([]templates.ExtraMapping, error) {
	transactionsExtraMappings := templates.ExtraMapping{
		Index:    indexer.TransactionsIndex,
		Mappings: noKibana.InnerTxs.ToBuffer(),
	}
	operationsExtraMappings := templates.ExtraMapping{
		Index:    indexer.OperationsIndex,
		Mappings: noKibana.InnerTxs.ToBuffer(),
	}

	return []templates.ExtraMapping{
		transactionsExtraMappings,
		operationsExtraMappings,
	}, nil
}
