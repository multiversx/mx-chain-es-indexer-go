package templatesAndPolicies

import (
	"bytes"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/templates/noKibana"
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
	indexTemplates[data.TransactionsIndex] = noKibana.Transactions.ToBuffer()
	indexTemplates[data.BlockIndex] = noKibana.Blocks.ToBuffer()
	indexTemplates[data.MiniblocksIndex] = noKibana.Miniblocks.ToBuffer()
	indexTemplates[data.RatingIndex] = noKibana.Rating.ToBuffer()
	indexTemplates[data.RoundsIndex] = noKibana.Rounds.ToBuffer()
	indexTemplates[data.ValidatorsIndex] = noKibana.Validators.ToBuffer()
	indexTemplates[data.AccountsIndex] = noKibana.Accounts.ToBuffer()
	indexTemplates[data.AccountsHistoryIndex] = noKibana.AccountsHistory.ToBuffer()
	indexTemplates[data.AccountsESDTIndex] = noKibana.AccountsESDT.ToBuffer()
	indexTemplates[data.AccountsESDTHistoryIndex] = noKibana.AccountsESDTHistory.ToBuffer()
	indexTemplates[data.EpochInfoIndex] = noKibana.EpochInfo.ToBuffer()
	indexTemplates[data.ReceiptsIndex] = noKibana.Receipts.ToBuffer()
	indexTemplates[data.ScResultsIndex] = noKibana.SCResults.ToBuffer()
	indexTemplates[data.SCDeploysIndex] = noKibana.SCDeploys.ToBuffer()
	indexTemplates[data.TokensIndex] = noKibana.Tokens.ToBuffer()
	indexTemplates[data.TagsIndex] = noKibana.Tags.ToBuffer()
	indexTemplates[data.LogsIndex] = noKibana.Logs.ToBuffer()
	indexTemplates[data.DelegatorsIndex] = noKibana.Delegators.ToBuffer()
	indexTemplates[data.OperationsIndex] = noKibana.Operations.ToBuffer()

	return indexTemplates, indexPolicies, nil
}
