package process

import (
	"bytes"

	"github.com/ElrondNetwork/elastic-indexer-go/templates/noKibana"
	"github.com/ElrondNetwork/elastic-indexer-go/templates/withKibana"
)

type templatesAndPoliciesReader struct {
	useKibana bool
}

// NewTemplatesAndPoliciesReader will create a new instance of templatesAndPoliciesReader
func NewTemplatesAndPoliciesReader(useKibana bool) *templatesAndPoliciesReader {
	return &templatesAndPoliciesReader{
		useKibana: useKibana,
	}
}

// GetElasticTemplatesAndPolicies will return elastic templates and policies
func (tpr *templatesAndPoliciesReader) GetElasticTemplatesAndPolicies() (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error) {
	indexTemplates := make(map[string]*bytes.Buffer)
	indexPolicies := make(map[string]*bytes.Buffer)

	if tpr.useKibana {
		indexTemplates = getTemplatesKibana()
		indexPolicies = getPolicies()

		return indexTemplates, indexPolicies, nil
	}

	indexTemplates = getTemplatesNoKibana()

	return indexTemplates, indexPolicies, nil
}

func getTemplatesKibana() map[string]*bytes.Buffer {
	indexTemplates := make(map[string]*bytes.Buffer)

	indexTemplates["opendistro"] = withKibana.OpenDistro.ToBuffer()
	indexTemplates[txIndex] = withKibana.Transactions.ToBuffer()
	indexTemplates[blockIndex] = withKibana.Blocks.ToBuffer()
	indexTemplates[miniblocksIndex] = withKibana.Miniblocks.ToBuffer()
	indexTemplates[tpsIndex] = withKibana.TPS.ToBuffer()
	indexTemplates[ratingIndex] = withKibana.Rating.ToBuffer()
	indexTemplates[roundIndex] = withKibana.Rounds.ToBuffer()
	indexTemplates[validatorsIndex] = withKibana.Validators.ToBuffer()
	indexTemplates[accountsIndex] = withKibana.Accounts.ToBuffer()
	indexTemplates[accountsHistoryIndex] = withKibana.AccountsHistory.ToBuffer()
	indexTemplates[accountsESDTIndex] = withKibana.AccountsESDT.ToBuffer()
	indexTemplates[accountsESDTHistoryIndex] = withKibana.AccountsESDTHistory.ToBuffer()
	indexTemplates[epochInfoIndex] = withKibana.EpochInfo.ToBuffer()
	indexTemplates[receiptsIndex] = withKibana.Receipts.ToBuffer()
	indexTemplates[scResultsIndex] = withKibana.SCResults.ToBuffer()

	return indexTemplates
}

func getTemplatesNoKibana() map[string]*bytes.Buffer {
	indexTemplates := make(map[string]*bytes.Buffer)

	indexTemplates["opendistro"] = noKibana.OpenDistro.ToBuffer()
	indexTemplates[txIndex] = noKibana.Transactions.ToBuffer()
	indexTemplates[blockIndex] = noKibana.Blocks.ToBuffer()
	indexTemplates[miniblocksIndex] = noKibana.Miniblocks.ToBuffer()
	indexTemplates[tpsIndex] = noKibana.TPS.ToBuffer()
	indexTemplates[ratingIndex] = noKibana.Rating.ToBuffer()
	indexTemplates[roundIndex] = noKibana.Rounds.ToBuffer()
	indexTemplates[validatorsIndex] = noKibana.Validators.ToBuffer()
	indexTemplates[accountsIndex] = noKibana.Accounts.ToBuffer()
	indexTemplates[accountsHistoryIndex] = noKibana.AccountsHistory.ToBuffer()
	indexTemplates[accountsESDTIndex] = noKibana.AccountsESDT.ToBuffer()
	indexTemplates[accountsESDTHistoryIndex] = noKibana.AccountsESDTHistory.ToBuffer()
	indexTemplates[epochInfoIndex] = noKibana.EpochInfo.ToBuffer()
	indexTemplates[receiptsIndex] = noKibana.Receipts.ToBuffer()
	indexTemplates[scResultsIndex] = noKibana.SCResults.ToBuffer()

	return indexTemplates
}

func getPolicies() map[string]*bytes.Buffer {
	indexesPolicies := make(map[string]*bytes.Buffer)

	indexesPolicies[txPolicy] = withKibana.TransactionsPolicy.ToBuffer()
	indexesPolicies[blockPolicy] = withKibana.BlocksPolicy.ToBuffer()
	indexesPolicies[miniblocksPolicy] = withKibana.MiniblocksPolicy.ToBuffer()
	indexesPolicies[ratingPolicy] = withKibana.RatingPolicy.ToBuffer()
	indexesPolicies[roundPolicy] = withKibana.RoundsPolicy.ToBuffer()
	indexesPolicies[validatorsPolicy] = withKibana.ValidatorsPolicy.ToBuffer()
	indexesPolicies[accountsHistoryPolicy] = withKibana.AccountsHistoryPolicy.ToBuffer()
	indexesPolicies[accountsPolicy] = withKibana.AccountsPolicy.ToBuffer()
	indexesPolicies[accountsESDTPolicy] = withKibana.AccountsESDTPolicy.ToBuffer()
	indexesPolicies[accountsESDTHistoryPolicy] = withKibana.AccountsESDTHistoryPolicy.ToBuffer()
	indexesPolicies[accountsHistoryPolicy] = withKibana.AccountsHistoryPolicy.ToBuffer()
	indexesPolicies[receiptsPolicy] = withKibana.ReceiptsPolicy.ToBuffer()
	indexesPolicies[scResultsPolicy] = withKibana.ScResultsPolicy.ToBuffer()

	return indexesPolicies
}
