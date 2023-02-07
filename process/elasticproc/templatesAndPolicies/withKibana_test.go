package templatesAndPolicies

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplatesAndPolicyReaderWithKibana_GetElasticTemplatesAndPolicies(t *testing.T) {
	t.Parallel()

	reader := NewTemplatesAndPolicyReaderWithKibana()

	templates, policies, err := reader.GetElasticTemplatesAndPolicies()
	require.Nil(t, err)
	require.Len(t, policies, 12)
	require.Len(t, templates, 21)
}
