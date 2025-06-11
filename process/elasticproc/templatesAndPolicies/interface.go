package templatesAndPolicies

import (
	"bytes"

	"github.com/multiversx/mx-chain-es-indexer-go/templates"
)

// TemplatesAndPoliciesHandler  defines the actions that a templates and policies handler should do
type TemplatesAndPoliciesHandler interface {
	GetElasticTemplatesAndPolicies() (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error)
	GetExtraMappings() ([]templates.ExtraMapping, error)
	GetTimestampMsMappings() ([]templates.ExtraMapping, error)
}
