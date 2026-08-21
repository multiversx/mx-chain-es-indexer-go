package elasticproc

import (
	"bytes"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	elasticIndexer "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/templates"
)

type indexCreator struct {
	elasticClient   DatabaseClientHandler
	mappingsHandler TemplatesAndPoliciesHandler
}

// NewIndexCreator will create a new instance of indexCreator
func NewIndexCreator(elasticClient DatabaseClientHandler, mappingsHandler TemplatesAndPoliciesHandler) (*indexCreator, error) {
	if check.IfNilReflect(elasticClient) {
		return nil, elasticIndexer.ErrNilDatabaseClient
	}
	if check.IfNilReflect(mappingsHandler) {
		return nil, elasticIndexer.ErrNilMappingsHandler
	}

	return &indexCreator{
		elasticClient:   elasticClient,
		mappingsHandler: mappingsHandler,
	}, nil
}

func (ic *indexCreator) CreateIndexes() error {
	indexTemplates, indexPolices, err := ic.mappingsHandler.GetElasticTemplatesAndPolicies()
	if err != nil {
		return err
	}

	err = ic.createIndices(indexTemplates)
	if err != nil {
		return err
	}

	err = ic.createPolicies(indexPolices)
	if err != nil {
		return err
	}

	extraMappings, err := ic.mappingsHandler.GetTimestampMsMappings()
	if err != nil {
		return err
	}

	return ic.addExtraMappings(extraMappings)
}

func (ic *indexCreator) createIndices(indexTemplateMap map[string]*bytes.Buffer) error {
	for index, indexData := range indexTemplateMap {
		err := ic.elasticClient.CheckAndCreateTemplate(index, indexData)
		if err != nil {
			return fmt.Errorf("elasticClient.CreateIndexWithMapping index: %s, error: %w", index, err)
		}

		indexWithSuffix := fmt.Sprintf("%s-%s", index, elasticIndexer.IndexSuffix)
		err = ic.elasticClient.CheckAndCreateIndex(indexWithSuffix)
		if err != nil {
			return fmt.Errorf("elasticClient.CheckAndCreateIndex index: %s, error: %w", index, err)
		}

		err = ic.elasticClient.CheckAndCreateAlias(index, indexWithSuffix)
		if err != nil {
			return fmt.Errorf("elasticClient.CheckAndCreateAlias index: %s, error: %w", index, err)
		}
	}

	return nil
}

func (ic *indexCreator) createPolicies(indexPolicyMap map[string]*bytes.Buffer) error {
	for index, policy := range indexPolicyMap {
		policyName := fmt.Sprintf("%s-%s", index, "policy")
		if ic.elasticClient.PolicyExists(policyName) {
			continue
		}

		indexWithSuffix := fmt.Sprintf("%s-%s", index, elasticIndexer.IndexSuffix)
		err := ic.elasticClient.SetWriteIndexTrue(index, indexWithSuffix)
		if err != nil {
			return fmt.Errorf("elasticClient.SetWriteIndexTrue index: %s, error: %w", index, err)
		}
		log.Info("elasticClient.SetWriteIndexTrue", "index", index)

		err = ic.elasticClient.CheckAndCreatePolicy(policyName, policy)
		if err != nil {
			return fmt.Errorf("databaseClient.PutPolicy index: %s, error: %w", index, err)
		}

		log.Info("databaseClient.PutPolicy", "index", index)
	}

	return nil
}

func (ic *indexCreator) addExtraMappings(extraMappings []templates.ExtraMapping) error {
	for _, mappingsTuple := range extraMappings {
		err := ic.elasticClient.PutMappings(mappingsTuple.Index, mappingsTuple.Mappings)
		if err != nil {
			log.Warn("cannot add extra mappings", "index", mappingsTuple.Index, "error", err)
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ic *indexCreator) IsInterfaceNil() bool {
	return ic == nil
}
