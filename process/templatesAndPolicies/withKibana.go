package templatesAndPolicies

import (
	"bytes"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/templates/withKibana"
)

type templatesAndPolicyReaderWithKibana struct{}

// NewTemplatesAndPolicyReaderWithKibana will create a new instance of templatesAndPolicyReaderWithKibana
func NewTemplatesAndPolicyReaderWithKibana() *templatesAndPolicyReaderWithKibana {
	return new(templatesAndPolicyReaderWithKibana)
}

// GetElasticTemplatesAndPolicies will return templates and policies
func (tr *templatesAndPolicyReaderWithKibana) GetElasticTemplatesAndPolicies() (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error) {
	indexTemplates := getTemplatesKibana()
	indexPolicies := getPolicies()

	return indexTemplates, indexPolicies, nil
}

func getTemplatesKibana() map[string]*bytes.Buffer {
	indexTemplates := make(map[string]*bytes.Buffer)

	indexTemplates["opendistro"] = withKibana.OpenDistro.ToBuffer()
	indexTemplates[data.TransactionsIndex] = withKibana.Transactions.ToBuffer()
	indexTemplates[data.BlockIndex] = withKibana.Blocks.ToBuffer()
	indexTemplates[data.MiniblocksIndex] = withKibana.Miniblocks.ToBuffer()
	indexTemplates[data.RatingIndex] = withKibana.Rating.ToBuffer()
	indexTemplates[data.RoundsIndex] = withKibana.Rounds.ToBuffer()
	indexTemplates[data.ValidatorsIndex] = withKibana.Validators.ToBuffer()
	indexTemplates[data.AccountsIndex] = withKibana.Accounts.ToBuffer()
	indexTemplates[data.AccountsHistoryIndex] = withKibana.AccountsHistory.ToBuffer()
	indexTemplates[data.AccountsESDTIndex] = withKibana.AccountsESDT.ToBuffer()
	indexTemplates[data.AccountsESDTHistoryIndex] = withKibana.AccountsESDTHistory.ToBuffer()
	indexTemplates[data.EpochInfoIndex] = withKibana.EpochInfo.ToBuffer()
	indexTemplates[data.ReceiptsIndex] = withKibana.Receipts.ToBuffer()
	indexTemplates[data.ScResultsIndex] = withKibana.SCResults.ToBuffer()
	indexTemplates[data.SCDeploysIndex] = withKibana.SCDeploys.ToBuffer()
	indexTemplates[data.TokensIndex] = withKibana.Tokens.ToBuffer()
	indexTemplates[data.TagsIndex] = withKibana.Tags.ToBuffer()
	indexTemplates[data.LogsIndex] = withKibana.Logs.ToBuffer()
	indexTemplates[data.DelegatorsIndex] = withKibana.Delegators.ToBuffer()
	indexTemplates[data.OperationsIndex] = withKibana.Operations.ToBuffer()

	return indexTemplates
}

func getPolicies() map[string]*bytes.Buffer {
	indexesPolicies := make(map[string]*bytes.Buffer)

	indexesPolicies[data.TransactionsPolicy] = withKibana.TransactionsPolicy.ToBuffer()
	indexesPolicies[data.BlockPolicy] = withKibana.BlocksPolicy.ToBuffer()
	indexesPolicies[data.MiniblocksPolicy] = withKibana.MiniblocksPolicy.ToBuffer()
	indexesPolicies[data.RatingPolicy] = withKibana.RatingPolicy.ToBuffer()
	indexesPolicies[data.RoundsPolicy] = withKibana.RoundsPolicy.ToBuffer()
	indexesPolicies[data.ValidatorsPolicy] = withKibana.ValidatorsPolicy.ToBuffer()
	indexesPolicies[data.AccountsHistoryPolicy] = withKibana.AccountsHistoryPolicy.ToBuffer()
	indexesPolicies[data.AccountsPolicy] = withKibana.AccountsPolicy.ToBuffer()
	indexesPolicies[data.AccountsESDTPolicy] = withKibana.AccountsESDTPolicy.ToBuffer()
	indexesPolicies[data.AccountsESDTHistoryPolicy] = withKibana.AccountsESDTHistoryPolicy.ToBuffer()
	indexesPolicies[data.AccountsHistoryPolicy] = withKibana.AccountsHistoryPolicy.ToBuffer()
	indexesPolicies[data.ReceiptsPolicy] = withKibana.ReceiptsPolicy.ToBuffer()
	indexesPolicies[data.ScResultsPolicy] = withKibana.ScResultsPolicy.ToBuffer()

	return indexesPolicies
}
