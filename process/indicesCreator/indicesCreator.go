package indicesCreator

import (
	"bytes"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var (
	log = logger.GetOrCreate("indexer/process/indicescreator")

	indexes = []string{
		data.TransactionsIndex, data.BlockIndex, data.MiniblocksIndex, data.RatingIndex, data.RoundsIndex, data.ValidatorsIndex,
		data.AccountsIndex, data.AccountsHistoryIndex, data.ReceiptsIndex, data.ScResultsIndex, data.AccountsESDTHistoryIndex, data.AccountsESDTIndex,
		data.EpochInfoIndex, data.SCDeploysIndex, data.TokensIndex, data.TagsIndex, data.LogsIndex, data.DelegatorsIndex, data.OperationsIndex,
	}

	indexesPolicies = []string{data.TransactionsPolicy, data.BlockPolicy, data.MiniblocksPolicy, data.RatingPolicy, data.RoundsPolicy, data.ValidatorsPolicy,
		data.AccountsPolicy, data.AccountsESDTPolicy, data.AccountsHistoryPolicy, data.AccountsESDTHistoryPolicy, data.AccountsESDTIndex, data.ReceiptsPolicy, data.ScResultsPolicy}
)

type creator struct {
	elasticClient ElasticClientIndices
}

func NewIndicesCreator(elasticClient ElasticClientIndices) (*creator, error) {
	return &creator{
		elasticClient: elasticClient,
	}, nil
}

func (c *creator) CreateIndicesIfNeeded(indexTemplates, _ map[string]*bytes.Buffer, useKibana bool) error {
	err := c.createOpenDistroTemplates(indexTemplates)
	if err != nil {
		return err
	}

	if useKibana {
		// TODO: Re-activate after we think of a solid way to handle forks+rotating indexes
		// err = ei.createIndexPolicies(indexPolicies)
		// if err != nil {
		//	return err
		// }
	}

	err = c.createIndexTemplates(indexTemplates)
	if err != nil {
		return err
	}

	err = c.createIndexes()
	if err != nil {
		return err
	}

	err = c.createAliases()
	if err != nil {
		return err
	}

	return nil
}

func (c *creator) createIndexPolicies(indexPolicies map[string]*bytes.Buffer) error {
	for _, indexPolicyName := range indexesPolicies {
		indexPolicy := getTemplateByName(indexPolicyName, indexPolicies)
		if indexPolicy != nil {
			err := c.elasticClient.CheckAndCreatePolicy(indexPolicyName, indexPolicy)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *creator) createOpenDistroTemplates(indexTemplates map[string]*bytes.Buffer) error {
	opendistroTemplate := getTemplateByName(data.OpenDistroIndex, indexTemplates)
	if opendistroTemplate != nil {
		err := c.elasticClient.CheckAndCreateTemplate(data.OpenDistroIndex, opendistroTemplate)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *creator) createIndexTemplates(indexTemplates map[string]*bytes.Buffer) error {
	for _, index := range indexes {
		indexTemplate := getTemplateByName(index, indexTemplates)
		if indexTemplate != nil {
			err := c.elasticClient.CheckAndCreateTemplate(index, indexTemplate)
			if err != nil {
				return fmt.Errorf("index: %s, error: %w", index, err)
			}
		}
	}
	return nil
}

func (c *creator) createIndexes() error {

	for _, index := range indexes {
		indexName := fmt.Sprintf("%s-%s", index, data.IndexSuffix)
		err := c.elasticClient.CheckAndCreateIndex(indexName)
		if err != nil {
			return fmt.Errorf("index: %s, error: %w", index, err)
		}
	}
	return nil
}

func (c *creator) createAliases() error {
	for _, index := range indexes {
		indexName := fmt.Sprintf("%s-%s", index, data.IndexSuffix)
		err := c.elasticClient.CheckAndCreateAlias(index, indexName)
		if err != nil {
			return err
		}
	}

	return nil
}

func getTemplateByName(templateName string, templateList map[string]*bytes.Buffer) *bytes.Buffer {
	if template, ok := templateList[templateName]; ok {
		return template
	}

	log.Debug("creator.getTemplateByName", "could not find template", templateName)
	return nil
}
