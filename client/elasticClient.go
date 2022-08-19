package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

const (
	errPolicyAlreadyExists = "document already exists"
)

var log = logger.GetOrCreate("indexer/client")

type (
	responseErrorHandler func(res *esapi.Response) error
	objectsMap           = map[string]interface{}
)

type elasticClient struct {
	elasticBaseUrl string
	es             *elasticsearch.Client

	// countScroll is used to be incremented after each scroll so the scroll duration is different each time,
	// bypassing any possible caching based on the same request
	countScroll int
}

// NewElasticClient will create a new instance of elasticClient
func NewElasticClient(cfg elasticsearch.Config) (*elasticClient, error) {
	if len(cfg.Addresses) == 0 {
		return nil, indexer.ErrNoElasticUrlProvided
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	ec := &elasticClient{
		es:             es,
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

// CheckAndCreateAlias creates a new alias if it does not already exist
func (ec *elasticClient) CheckAndCreateAlias(alias string, indexName string) error {
	if ec.aliasExists(alias) {
		return nil
	}

	return ec.createAlias(alias, indexName)
}

// DoRequest will do a request to elastic server
func (ec *elasticClient) DoRequest(req *esapi.IndexRequest) error {
	res, err := req.Do(context.Background(), ec.es)
	if err != nil {
		return err
	}

	return parseResponse(res, nil, elasticDefaultErrorResponseHandler)
}

// DoBulkRequest will do a bulk of request to elastic server
func (ec *elasticClient) DoBulkRequest(buff *bytes.Buffer, index string) error {
	reader := bytes.NewReader(buff.Bytes())

	options := make([]func(*esapi.BulkRequest), 0)
	if index != "" {
		options = append(options, ec.es.Bulk.WithIndex(index))
	}

	res, err := ec.es.Bulk(
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

// DoMultiGet wil do a multi get request to elaticsearch server
func (ec *elasticClient) DoMultiGet(ids []string, index string, withSource bool, resBody interface{}) error {
	obj := getDocumentsByIDsQuery(ids, withSource)
	body, err := encode(obj)
	if err != nil {
		return err
	}

	res, err := ec.es.Mget(
		&body,
		ec.es.Mget.WithIndex(index),
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
func (ec *elasticClient) DoQueryRemove(index string, body *bytes.Buffer) error {
	res, err := ec.es.DeleteByQuery(
		[]string{index},
		body,
		ec.es.DeleteByQuery.WithIgnoreUnavailable(true),
	)

	if err != nil {
		log.Warn("elasticClient.DoQueryRemove", "cannot do query remove", err.Error())
		return err
	}

	var decodedBody objectsMap
	err = parseResponse(res, &decodedBody, elasticDefaultErrorResponseHandler)
	if err != nil {
		log.Warn("elasticClient.DoQueryRemove", "error parsing response", err.Error())
		return err
	}

	return nil
}

// TemplateExists checks weather a template is already created
func (ec *elasticClient) templateExists(index string) bool {
	res, err := ec.es.Indices.ExistsTemplate([]string{index})
	return exists(res, err)
}

// IndexExists checks if a given index already exists
func (ec *elasticClient) indexExists(index string) bool {
	res, err := ec.es.Indices.Exists([]string{index})
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
	res, err := ec.es.Transport.Perform(req)
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
	res, err := ec.es.Transport.Perform(req)
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
	res, err := ec.es.Indices.Create(index)
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
	res, err := ec.es.Transport.Perform(req)
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
		return indexer.ErrCouldNotCreatePolicy
	}

	return nil
}

// CreateIndexTemplate creates an elasticsearch index template
func (ec *elasticClient) createIndexTemplate(templateName string, template io.Reader) error {
	res, err := ec.es.Indices.PutTemplate(templateName, template)
	if err != nil {
		return err
	}

	return parseResponse(res, nil, elasticDefaultErrorResponseHandler)
}

// CreateAlias creates an index alias
func (ec *elasticClient) createAlias(alias string, index string) error {
	res, err := ec.es.Indices.PutAlias([]string{index}, alias)
	if err != nil {
		return err
	}

	return parseResponse(res, nil, elasticDefaultErrorResponseHandler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ec *elasticClient) IsInterfaceNil() bool {
	return ec == nil
}
