package templatesAndPolicies

import "bytes"

// TemplatesAndPoliciesHandler  defines the actions that a templates and policies handler should do
type TemplatesAndPoliciesHandler interface {
	GetElasticTemplatesAndPolicies() (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error)
}
