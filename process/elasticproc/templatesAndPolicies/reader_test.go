package templatesAndPolicies

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplatesAndPolicyReaderNoKibana_GetElasticTemplatesAndPolicies(t *testing.T) {
	t.Parallel()

	reader := NewTemplatesAndPolicyReader()

	templates, policies, err := reader.GetElasticTemplatesAndPolicies()
	require.Nil(t, err)
	require.Len(t, policies, 0)
	require.Len(t, templates, 23)
}
