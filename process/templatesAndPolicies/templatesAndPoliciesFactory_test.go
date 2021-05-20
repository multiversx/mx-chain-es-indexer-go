package templatesAndPolicies

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateTemplatesAndPoliciesReader_NoKibana(t *testing.T) {
	t.Parallel()

	reader := CreateTemplatesAndPoliciesReader(false)

	_, ok := reader.(*templatesAndPolicyReaderNoKibana)
	require.True(t, ok)
}

func TestCreateTemplatesAndPoliciesReader_WithKibana(t *testing.T) {
	t.Parallel()

	reader := CreateTemplatesAndPoliciesReader(true)

	_, ok := reader.(*templatesAndPolicyReaderWithKibana)
	require.True(t, ok)
}
