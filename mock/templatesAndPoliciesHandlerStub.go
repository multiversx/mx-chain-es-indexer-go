package mock

import (
	"bytes"

	"github.com/multiversx/mx-chain-es-indexer-go/templates"
)

// TemplatesAndPoliciesHandlerStub -
type TemplatesAndPoliciesHandlerStub struct {
	GetElasticTemplatesAndPoliciesCalled func() (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error)
	GetExtraMappingsCalled               func() ([]templates.ExtraMapping, error)
	GetTimestampMsMappingsCalled         func() ([]templates.ExtraMapping, error)
}

// GetElasticTemplatesAndPolicies -
func (stub *TemplatesAndPoliciesHandlerStub) GetElasticTemplatesAndPolicies() (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error) {
	if stub.GetElasticTemplatesAndPoliciesCalled != nil {
		return stub.GetElasticTemplatesAndPoliciesCalled()
	}
	return nil, nil, nil
}

// GetExtraMappings -
func (stub *TemplatesAndPoliciesHandlerStub) GetExtraMappings() ([]templates.ExtraMapping, error) {
	if stub.GetExtraMappingsCalled != nil {
		return stub.GetExtraMappingsCalled()
	}
	return nil, nil
}

// GetTimestampMsMappings -
func (stub *TemplatesAndPoliciesHandlerStub) GetTimestampMsMappings() ([]templates.ExtraMapping, error) {
	if stub.GetTimestampMsMappingsCalled != nil {
		return stub.GetTimestampMsMappingsCalled()
	}
	return nil, nil
}
