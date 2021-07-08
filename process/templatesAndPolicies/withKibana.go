package templatesAndPolicies

import (
	"bytes"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
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
	indexTemplates[indexer.TransactionsIndex] = withKibana.Transactions.ToBuffer()
	indexTemplates[indexer.BlockIndex] = withKibana.Blocks.ToBuffer()
	indexTemplates[indexer.MiniblocksIndex] = withKibana.Miniblocks.ToBuffer()
	indexTemplates[indexer.TpsIndex] = withKibana.TPS.ToBuffer()
	indexTemplates[indexer.RatingIndex] = withKibana.Rating.ToBuffer()
	indexTemplates[indexer.RoundsIndex] = withKibana.Rounds.ToBuffer()
	indexTemplates[indexer.ValidatorsIndex] = withKibana.Validators.ToBuffer()
	indexTemplates[indexer.AccountsIndex] = withKibana.Accounts.ToBuffer()
	indexTemplates[indexer.AccountsHistoryIndex] = withKibana.AccountsHistory.ToBuffer()
	indexTemplates[indexer.AccountsESDTIndex] = withKibana.AccountsESDT.ToBuffer()
	indexTemplates[indexer.AccountsESDTHistoryIndex] = withKibana.AccountsESDTHistory.ToBuffer()
	indexTemplates[indexer.EpochInfoIndex] = withKibana.EpochInfo.ToBuffer()
	indexTemplates[indexer.ReceiptsIndex] = withKibana.Receipts.ToBuffer()
	indexTemplates[indexer.ScResultsIndex] = withKibana.SCResults.ToBuffer()
	indexTemplates[indexer.SCDeploysIndex] = withKibana.SCDeploys.ToBuffer()
	indexTemplates[indexer.TokensIndex] = withKibana.Tokens.ToBuffer()
	indexTemplates[indexer.TagsIndex] = withKibana.Tags.ToBuffer()
	indexTemplates[indexer.LogsIndex] = withKibana.Logs.ToBuffer()

	return indexTemplates
}

func getPolicies() map[string]*bytes.Buffer {
	indexesPolicies := make(map[string]*bytes.Buffer)

	indexesPolicies[indexer.TransactionsPolicy] = withKibana.TransactionsPolicy.ToBuffer()
	indexesPolicies[indexer.BlockPolicy] = withKibana.BlocksPolicy.ToBuffer()
	indexesPolicies[indexer.MiniblocksPolicy] = withKibana.MiniblocksPolicy.ToBuffer()
	indexesPolicies[indexer.RatingPolicy] = withKibana.RatingPolicy.ToBuffer()
	indexesPolicies[indexer.RoundsPolicy] = withKibana.RoundsPolicy.ToBuffer()
	indexesPolicies[indexer.ValidatorsPolicy] = withKibana.ValidatorsPolicy.ToBuffer()
	indexesPolicies[indexer.AccountsHistoryPolicy] = withKibana.AccountsHistoryPolicy.ToBuffer()
	indexesPolicies[indexer.AccountsPolicy] = withKibana.AccountsPolicy.ToBuffer()
	indexesPolicies[indexer.AccountsESDTPolicy] = withKibana.AccountsESDTPolicy.ToBuffer()
	indexesPolicies[indexer.AccountsESDTHistoryPolicy] = withKibana.AccountsESDTHistoryPolicy.ToBuffer()
	indexesPolicies[indexer.AccountsHistoryPolicy] = withKibana.AccountsHistoryPolicy.ToBuffer()
	indexesPolicies[indexer.ReceiptsPolicy] = withKibana.ReceiptsPolicy.ToBuffer()
	indexesPolicies[indexer.ScResultsPolicy] = withKibana.ScResultsPolicy.ToBuffer()

	return indexesPolicies
}
