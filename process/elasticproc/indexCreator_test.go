package elasticproc

import (
	"bytes"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/templatesAndPolicies"
	"github.com/multiversx/mx-chain-es-indexer-go/templates"
	"github.com/stretchr/testify/require"
)

func createMockIndexCreatorArgs() (*mock.DatabaseWriterStub, TemplatesAndPoliciesHandler) {
	return &mock.DatabaseWriterStub{}, templatesAndPolicies.NewTemplatesAndPolicyReader(false, "", nil, nil)
}

func TestNewIndexCreator_NilDatabaseClient(t *testing.T) {
	t.Parallel()

	_, mappingsHandler := createMockIndexCreatorArgs()
	ic, err := NewIndexCreator(nil, mappingsHandler)
	require.Nil(t, ic)
	require.Equal(t, dataindexer.ErrNilDatabaseClient, err)
}

func TestNewIndexCreator_NilMappingsHandler(t *testing.T) {
	t.Parallel()

	dbClient, _ := createMockIndexCreatorArgs()
	ic, err := NewIndexCreator(dbClient, nil)
	require.Nil(t, ic)
	require.Equal(t, dataindexer.ErrNilMappingsHandler, err)
}

func TestNewIndexCreator_ShouldWork(t *testing.T) {
	t.Parallel()

	dbClient, mappingsHandler := createMockIndexCreatorArgs()
	ic, err := NewIndexCreator(dbClient, mappingsHandler)
	require.NotNil(t, ic)
	require.Nil(t, err)
}

func TestIndexCreator_CreateIndexes(t *testing.T) {
	t.Parallel()

	dbClient, mappingsHandler := createMockIndexCreatorArgs()
	ic, err := NewIndexCreator(dbClient, mappingsHandler)
	require.NotNil(t, ic)
	require.Nil(t, err)

	err = ic.CreateIndexes()
	require.Nil(t, err)
}

func TestIndexCreator_CreateIndexes_GetElasticTemplatesAndPoliciesFails(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	dbClient, _ := createMockIndexCreatorArgs()
	mappingsHandler := &mock.TemplatesAndPoliciesHandlerStub{
		GetElasticTemplatesAndPoliciesCalled: func() (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error) {
			return nil, nil, expectedErr
		},
	}

	ic, err := NewIndexCreator(dbClient, mappingsHandler)
	require.NotNil(t, ic)
	require.Nil(t, err)

	err = ic.CreateIndexes()
	require.Equal(t, expectedErr, err)
}

func TestIndexCreator_CreateIndexes_CheckAndCreateTemplateFails(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("template error")
	dbClient := &mock.DatabaseWriterStub{
		CheckAndCreateTemplateCalled: func(_ string, _ *bytes.Buffer) error {
			return expectedErr
		},
	}

	mappingsHandler := &mock.TemplatesAndPoliciesHandlerStub{
		GetElasticTemplatesAndPoliciesCalled: func() (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error) {
			return map[string]*bytes.Buffer{
				"test-index": bytes.NewBufferString("template"),
			}, nil, nil
		},
	}

	ic, err := NewIndexCreator(dbClient, mappingsHandler)
	require.NotNil(t, ic)
	require.Nil(t, err)

	err = ic.CreateIndexes()
	require.True(t, errors.Is(err, expectedErr))
}

func TestIndexCreator_CreateIndexes_CheckAndCreateIndexFails(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("index error")
	dbClient := &mock.DatabaseWriterStub{
		CheckAndCreateIndexCalled: func(_ string) error {
			return expectedErr
		},
	}

	mappingsHandler := &mock.TemplatesAndPoliciesHandlerStub{
		GetElasticTemplatesAndPoliciesCalled: func() (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error) {
			return map[string]*bytes.Buffer{
				"test-index": bytes.NewBufferString("template"),
			}, nil, nil
		},
	}

	ic, err := NewIndexCreator(dbClient, mappingsHandler)
	require.NotNil(t, ic)
	require.Nil(t, err)

	err = ic.CreateIndexes()
	require.True(t, errors.Is(err, expectedErr))
}

func TestIndexCreator_CreateIndexes_CheckAndCreateAliasFails(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("alias error")
	dbClient := &mock.DatabaseWriterStub{
		CheckAndCreateAliasCalled: func(_ string, _ string) error {
			return expectedErr
		},
	}

	mappingsHandler := &mock.TemplatesAndPoliciesHandlerStub{
		GetElasticTemplatesAndPoliciesCalled: func() (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error) {
			return map[string]*bytes.Buffer{
				"test-index": bytes.NewBufferString("template"),
			}, nil, nil
		},
	}

	ic, err := NewIndexCreator(dbClient, mappingsHandler)
	require.NotNil(t, ic)
	require.Nil(t, err)

	err = ic.CreateIndexes()
	require.True(t, errors.Is(err, expectedErr))
}

func TestIndexCreator_CreateIndexes_SetWriteIndexTrueFails(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("write index error")
	dbClient := &mock.DatabaseWriterStub{
		SetWriteIndexTrueCalled: func(_ string, _ string) error {
			return expectedErr
		},
	}

	mappingsHandler := &mock.TemplatesAndPoliciesHandlerStub{
		GetElasticTemplatesAndPoliciesCalled: func() (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error) {
			return nil, map[string]*bytes.Buffer{
				"test-index": bytes.NewBufferString("policy"),
			}, nil
		},
	}

	ic, err := NewIndexCreator(dbClient, mappingsHandler)
	require.NotNil(t, ic)
	require.Nil(t, err)

	err = ic.CreateIndexes()
	require.True(t, errors.Is(err, expectedErr))
}

func TestIndexCreator_CreateIndexes_CheckAndCreatePolicyFails(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("policy error")
	dbClient := &mock.DatabaseWriterStub{
		CheckAndCreatePolicyCalled: func(_ string, _ *bytes.Buffer) error {
			return expectedErr
		},
	}

	mappingsHandler := &mock.TemplatesAndPoliciesHandlerStub{
		GetElasticTemplatesAndPoliciesCalled: func() (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error) {
			return nil, map[string]*bytes.Buffer{
				"test-index": bytes.NewBufferString("policy"),
			}, nil
		},
	}

	ic, err := NewIndexCreator(dbClient, mappingsHandler)
	require.NotNil(t, ic)
	require.Nil(t, err)

	err = ic.CreateIndexes()
	require.True(t, errors.Is(err, expectedErr))
}

func TestIndexCreator_CreateIndexes_GetTimestampMsMappingsFails(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("mappings error")
	dbClient, _ := createMockIndexCreatorArgs()
	mappingsHandler := &mock.TemplatesAndPoliciesHandlerStub{
		GetElasticTemplatesAndPoliciesCalled: func() (map[string]*bytes.Buffer, map[string]*bytes.Buffer, error) {
			return nil, nil, nil
		},
		GetTimestampMsMappingsCalled: func() ([]templates.ExtraMapping, error) {
			return nil, expectedErr
		},
	}

	ic, err := NewIndexCreator(dbClient, mappingsHandler)
	require.NotNil(t, ic)
	require.Nil(t, err)

	err = ic.CreateIndexes()
	require.Equal(t, expectedErr, err)
}

func TestIndexCreator_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var ic *indexCreator
	require.True(t, ic.IsInterfaceNil())

	ic = &indexCreator{}
	require.False(t, ic.IsInterfaceNil())
}
