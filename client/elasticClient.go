package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const (
	esConflictsPolicy = "proceed"
)

var log = logger.GetOrCreate("indexer/client")

type (
	responseErrorHandler func(res *esapi.Response) error
	objectsMap           = map[string]interface{}
)

type elasticClient struct {
	client *elasticsearch.Client

	// countScroll is used to be incremented after each scroll so the scroll duration is different each time,
	// bypassing any possible caching based on the same request
	countScroll int
	mutex       sync.Mutex
}

// NewElasticClient will create a new instance of elasticClient
func NewElasticClient(cfg elasticsearch.Config) (*elasticClient, error) {
	if len(cfg.Addresses) == 0 {
		return nil, dataindexer.ErrNoElasticUrlProvided
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	ec := &elasticClient{
		client: es,
	}

	return ec, nil
}

// CheckAndCreateTemplate creates an index template if it does not already exist
func (ec *elasticClient) CheckAndCreateTemplate(templateName string, template *bytes.Buffer) error {
	if ec.templateExists(templateName) {
		return nil
	}

	return ec.createIndexTemplate(templateName, template)
}

// CheckAndCreatePolicy creates a new index policy if it does not already exist
func (ec *elasticClient) CheckAndCreatePolicy(policyName string, policy *bytes.Buffer) error {
	if ec.PolicyExists(policyName) {
		return nil
	}

	return ec.createPolicy(policyName, policy)
}

// SetWriteIndexTrue will set the provided index as write index
func (ec *elasticClient) SetWriteIndexTrue(alias string, providedIndex string) error {
	writeIndexFromCluster, isSet, err := ec.getWriteIndex(alias)
	if err != nil {
		return fmt.Errorf("err getting write index from cluster: %v", err)
	}
	if isSet {
		return nil
	}
	if writeIndexFromCluster == alias {
		writeIndexFromCluster = providedIndex
	}

	body := fmt.Sprintf(`{"actions":[{"add":{"index":"%s","alias":"%s","is_write_index":true}}]}`, writeIndexFromCluster, alias)
	res, err := ec.client.Indices.UpdateAliases(
		bytes.NewBufferString(body),
	)
	if err != nil {
		return err
	}

	defer closeBody(res)

	if res.IsError() {
		return fmt.Errorf("failed to set write index for alias %s: %s", alias, res.String())
	}

	return nil
}

// CheckAndCreateIndex creates a new index if it does not already exist
func (ec *elasticClient) CheckAndCreateIndex(indexName string) error {
	if ec.indexExists(indexName) {
		return nil
	}

	return ec.createIndex(indexName)
}

// PutMappings will put the provided mappings to a given index
func (ec *elasticClient) PutMappings(indexName string, mappings *bytes.Buffer) error {
	res, err := ec.client.Indices.PutMapping(
		mappings,
		ec.client.Indices.PutMapping.WithIndex(indexName),
	)
	if err != nil {
		return err
	}

	return parseResponse(res, nil, elasticDefaultErrorResponseHandler)
}

// CheckAndCreateAlias creates a new alias if it does not already exist
func (ec *elasticClient) CheckAndCreateAlias(alias string, indexName string) error {
	if ec.aliasExists(alias) {
		return nil
	}

	return ec.createAlias(alias, indexName)
}

// DoBulkRequest will do a bulk of request to elastic server
func (ec *elasticClient) DoBulkRequest(ctx context.Context, buff *bytes.Buffer, index string) error {
	options := make([]func(*esapi.BulkRequest), 0, 2)
	if index != "" {
		options = append(options, ec.client.Bulk.WithIndex(index))
	}

	options = append(options, ec.client.Bulk.WithContext(ctx))

	res, err := ec.client.Bulk(
		buff,
		options...,
	)
	if err != nil {
		log.Warn("elasticClient.DoBulkRequest", "error", err.Error())
		return err
	}

	return elasticBulkRequestResponseHandler(res)
}

// DoMultiGet wil do a multi get request to Elasticsearch server
func (ec *elasticClient) DoMultiGet(ctx context.Context, ids []string, index string, withSource bool, resBody interface{}) error {
	obj := getDocumentsByIDsQuery(ids, withSource)
	body, err := encode(obj)
	if err != nil {
		return err
	}

	res, err := ec.client.Mget(
		&body,
		ec.client.Mget.WithIndex(index),
		ec.client.Mget.WithContext(ctx),
	)
	if err != nil {
		log.Warn("elasticClient.DoMultiGet", "error", err.Error())
		return err
	}

	err = parseResponse(res, &resBody, elasticDefaultErrorResponseHandler)
	if err != nil {
		log.Warn("elasticClient.DoMultiGet", "error parsing response", err.Error())
		return err
	}

	return nil
}

// DoQueryRemove will do a query remove to elasticsearch server
func (ec *elasticClient) DoQueryRemove(ctx context.Context, index string, body *bytes.Buffer) error {
	err := ec.doRefresh(index)
	if err != nil {
		log.Warn("elasticClient.doRefresh", "cannot do refresh", err)
	}

	writeIndex, _, err := ec.getWriteIndex(index)
	if err != nil {
		log.Warn("elasticClient.getWriteIndex", "cannot do get write index", err)
		return err
	}

	res, err := ec.client.DeleteByQuery(
		[]string{writeIndex},
		body,
		ec.client.DeleteByQuery.WithIgnoreUnavailable(true),
		ec.client.DeleteByQuery.WithConflicts(esConflictsPolicy),
		ec.client.DeleteByQuery.WithContext(ctx),
	)

	if err != nil {
		log.Warn("elasticClient.DoQueryRemove", "cannot do query remove", err)
		return err
	}

	err = parseResponse(res, nil, elasticDefaultErrorResponseHandler)
	if err != nil {
		log.Warn("elasticClient.DoQueryRemove", "error parsing response", err)
		return err
	}

	return nil
}

func (ec *elasticClient) doRefresh(index string) error {
	res, err := ec.client.Indices.Refresh(
		ec.client.Indices.Refresh.WithIndex(index),
		ec.client.Indices.Refresh.WithIgnoreUnavailable(true),
	)
	if err != nil {
		return err
	}

	return parseResponse(res, nil, elasticDefaultErrorResponseHandler)
}

// TemplateExists checks weather a template is already created
func (ec *elasticClient) templateExists(index string) bool {
	res, err := ec.client.Indices.ExistsTemplate([]string{index})
	return exists(res, err)
}

// IndexExists checks if a given index already exists
func (ec *elasticClient) indexExists(index string) bool {
	res, err := ec.client.Indices.Exists([]string{index})
	return exists(res, err)
}

func (ec *elasticClient) aliasExists(alias string) bool {
	res, err := ec.client.Indices.ExistsAlias([]string{alias})
	return exists(res, err)
}

// PolicyExists checks if a policy was already created
func (ec *elasticClient) PolicyExists(policy string) bool {
	res, err := ec.client.ILM.GetLifecycle(
		ec.client.ILM.GetLifecycle.WithPolicy(policy),
	)
	return exists(res, err)
}

// CreateIndex creates an elasticsearch index
func (ec *elasticClient) createIndex(index string) error {
	res, err := ec.client.Indices.Create(index)
	if err != nil {
		return err
	}

	return parseResponse(res, nil, elasticDefaultErrorResponseHandler)
}

// CreatePolicy creates a new policy for elastic indexes. Policies define rollover parameters
func (ec *elasticClient) createPolicy(policyName string, policy *bytes.Buffer) error {
	res, err := ec.client.ILM.PutLifecycle(
		policyName,
		ec.client.ILM.PutLifecycle.WithBody(policy),
	)
	if err != nil {
		return err
	}

	return parseResponse(res, nil, elasticDefaultErrorResponseHandler)
}

// CreateIndexTemplate creates an elasticsearch index template
func (ec *elasticClient) createIndexTemplate(templateName string, template io.Reader) error {
	res, err := ec.client.Indices.PutIndexTemplate(templateName, template)
	if err != nil {
		return err
	}

	return parseResponse(res, nil, elasticDefaultErrorResponseHandler)
}

// CreateAlias creates an index alias
func (ec *elasticClient) createAlias(alias string, index string) error {
	res, err := ec.client.Indices.PutAlias([]string{index}, alias)
	if err != nil {
		return err
	}

	return parseResponse(res, nil, elasticDefaultErrorResponseHandler)
}

func (ec *elasticClient) getWriteIndex(alias string) (string, bool, error) {
	res, err := ec.client.Indices.GetAlias(
		ec.client.Indices.GetAlias.WithIndex(alias),
	)
	if err != nil {
		return "", false, err
	}

	var indexData map[string]struct {
		Aliases map[string]struct {
			IsWriteIndex bool `json:"is_write_index"`
		} `json:"aliases"`
	}
	err = parseResponse(res, &indexData, elasticDefaultErrorResponseHandler)
	if err != nil {
		return "", false, err
	}

	for index, details := range indexData {
		for _, indexAlias := range details.Aliases {
			if indexAlias.IsWriteIndex {
				return index, true, nil
			}
		}
		if len(indexData) == 1 {
			return index, false, nil
		}
	}

	return alias, false, nil
}

// UpdateByQuery will update all the documents that match the provided query from the provided index
func (ec *elasticClient) UpdateByQuery(ctx context.Context, index string, buff *bytes.Buffer) error {
	res, err := ec.client.UpdateByQuery(
		[]string{index},
		ec.client.UpdateByQuery.WithBody(buff),
		ec.client.UpdateByQuery.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	return parseResponse(res, nil, elasticDefaultErrorResponseHandler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ec *elasticClient) IsInterfaceNil() bool {
	return ec == nil
}
