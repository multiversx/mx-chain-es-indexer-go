package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// TODO add more unit tests

const (
	esConflictsPolicy      = "proceed"
	errPolicyAlreadyExists = "document already exists"
)

var log = logger.GetOrCreate("indexer/client")

type (
	responseErrorHandler func(res *esapi.Response) error
	objectsMap           = map[string]interface{}
)

type elasticClient struct {
	elasticBaseUrl string
	client         *elasticsearch.Client

	// countScroll is used to be incremented after each scroll so the scroll duration is different each time,
	// bypassing any possible caching based on the same request
	countScroll int
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
		client:         es,
		elasticBaseUrl: cfg.Addresses[0],
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

	if res.IsError() {
		return errors.New(res.String())
	}

	return nil
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
	reader := bytes.NewReader(buff.Bytes())

	options := make([]func(*esapi.BulkRequest), 0)
	if index != "" {
		options = append(options, ec.client.Bulk.WithIndex(index))
	}

	options = append(options, ec.client.Bulk.WithContext(ctx))

	res, err := ec.client.Bulk(
		reader,
		options...,
	)
	if err != nil {
		log.Warn("elasticClient.DoBulkRequest",
			"indexer do bulk request no response", err.Error())
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
		log.Warn("elasticClient.DoMultiGet",
			"cannot do multi get no response", err.Error())
		return err
	}

	err = parseResponse(res, &resBody, elasticDefaultErrorResponseHandler)
	if err != nil {
		log.Warn("elasticClient.DoMultiGet",
			"error parsing response", err.Error())
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

	writeIndex, err := ec.getWriteIndex(index)
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

// PolicyExists checks if a policy was already created
func (ec *elasticClient) PolicyExists(policy string) bool {
	policyRoute := fmt.Sprintf(
		"%s/%s/ism/policies/%s",
		ec.elasticBaseUrl,
		kibanaPluginPath,
		policy,
	)

	req := newRequest(http.MethodGet, policyRoute, nil)
	res, err := ec.client.Transport.Perform(req)
	if err != nil {
		log.Warn("elasticClient.PolicyExists",
			"error performing request", err.Error())
		return false
	}

	response := &esapi.Response{
		StatusCode: res.StatusCode,
		Body:       res.Body,
		Header:     res.Header,
	}

	existsRes := &data.Response{}
	err = parseResponse(response, existsRes, kibanaResponseErrorHandler)
	if err != nil {
		log.Warn("elasticClient.PolicyExists",
			"error returned by kibana api", err.Error())
		return false
	}

	return existsRes.Status == http.StatusConflict
}

// AliasExists checks if an index alias already exists
func (ec *elasticClient) aliasExists(alias string) bool {
	aliasRoute := fmt.Sprintf(
		"/_alias/%s",
		alias,
	)

	req := newRequest(http.MethodHead, aliasRoute, nil)
	res, err := ec.client.Transport.Perform(req)
	if err != nil {
		log.Warn("elasticClient.AliasExists",
			"error performing request", err.Error())
		return false
	}

	response := &esapi.Response{
		StatusCode: res.StatusCode,
		Body:       res.Body,
		Header:     res.Header,
	}

	return exists(response, nil)
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
	policyRoute := fmt.Sprintf(
		"%s/_opendistro/_ism/policies/%s",
		ec.elasticBaseUrl,
		policyName,
	)

	req := newRequest(http.MethodPut, policyRoute, policy)
	req.Header[headerContentType] = headerContentTypeJSON
	req.Header[headerXSRF] = []string{"false"}
	res, err := ec.client.Transport.Perform(req)
	if err != nil {
		return err
	}

	response := &esapi.Response{
		StatusCode: res.StatusCode,
		Body:       res.Body,
		Header:     res.Header,
	}

	existsRes := &data.Response{}
	err = parseResponse(response, existsRes, kibanaResponseErrorHandler)
	if err != nil {
		return err
	}

	errStr := fmt.Sprintf("%v", existsRes.Error)
	if existsRes.Status == http.StatusConflict && !strings.Contains(errStr, errPolicyAlreadyExists) {
		return dataindexer.ErrCouldNotCreatePolicy
	}

	return nil
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

func (ec *elasticClient) getWriteIndex(alias string) (string, error) {
	res, err := ec.client.Indices.GetAlias(
		ec.client.Indices.GetAlias.WithIndex(alias),
	)
	if err != nil {
		return "", err
	}

	var indexData map[string]struct {
		Aliases map[string]struct {
			IsWriteIndex bool `json:"is_write_index"`
		} `json:"aliases"`
	}
	err = parseResponse(res, &indexData, elasticDefaultErrorResponseHandler)
	if err != nil {
		return "", err
	}

	for index, details := range indexData {
		if len(indexData) == 1 {
			return index, nil
		}

		for _, indexAlias := range details.Aliases {
			if indexAlias.IsWriteIndex {
				return index, nil
			}
		}
	}

	return alias, nil
}

// UpdateByQuery will update all the documents that match the provided query from the provided index
func (ec *elasticClient) UpdateByQuery(ctx context.Context, index string, buff *bytes.Buffer) error {
	reader := bytes.NewReader(buff.Bytes())
	res, err := ec.client.UpdateByQuery(
		[]string{index},
		ec.client.UpdateByQuery.WithBody(reader),
		ec.client.UpdateByQuery.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("%s", res.String())
	}

	return parseResponse(res, nil, elasticDefaultErrorResponseHandler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ec *elasticClient) IsInterfaceNil() bool {
	return ec == nil
}
