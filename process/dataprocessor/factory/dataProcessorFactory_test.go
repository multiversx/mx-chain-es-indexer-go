package factory

import (
	"bytes"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/outputs"
	"github.com/stretchr/testify/require"
)

func createMockArgDataProcessorFactory() ArgDataProcessorFactory {
	return ArgDataProcessorFactory{
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		AddressPubkeyConverter:   mock.NewPubkeyConverterMock(32),
		ValidatorPubkeyConverter: &mock.PubkeyConverterMock{},
		DBClient:                 &mock.DatabaseWriterStub{},
		EnabledIndexes:           []string{"blocks"},
		Denomination:             1,
	}
}

func TestCreateDataProcessor(t *testing.T) {
	t.Parallel()

	dp, err := CreateDataProcessor(createMockArgDataProcessorFactory())
	require.Nil(t, err)
	require.NotNil(t, dp)
}

func TestCreateDataProcessor_NoEnabledIndexes(t *testing.T) {
	t.Parallel()

	args := createMockArgDataProcessorFactory()
	args.EnabledIndexes = nil

	dp, err := CreateDataProcessor(args)
	require.Nil(t, dp)
	require.Equal(t, dataindexer.ErrEmptyEnabledIndexes, err)
}

func TestCreateDataProcessor_WithExtraOutputs(t *testing.T) {
	t.Parallel()

	closeCalled := false
	args := createMockArgDataProcessorFactory()
	args.ExtraOutputs = []outputs.OutputWithPolicy{
		{
			Name: "extra",
			Output: &mock.OutputStub{
				CloseCalled: func() error {
					closeCalled = true
					return nil
				},
			},
		},
	}

	dp, err := CreateDataProcessor(args)
	require.Nil(t, err)

	err = dp.Close()
	require.Nil(t, err)
	require.True(t, closeCalled)
}

func TestCreateDataProcessor_ExtraOutputsAreCalledBeforeElasticsearch(t *testing.T) {
	t.Parallel()

	extraOutputCalled := false
	elasticWrittenAfterExtraOutput := false
	args := createMockArgDataProcessorFactory()
	// the rounds info is written by one bulk request, the Elasticsearch output has no other dependencies for it
	args.EnabledIndexes = []string{dataindexer.RoundsIndex}
	args.DBClient = &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(_ *bytes.Buffer, _ string) error {
			elasticWrittenAfterExtraOutput = extraOutputCalled
			return nil
		},
	}
	args.ExtraOutputs = []outputs.OutputWithPolicy{
		{
			Name: "extra",
			Output: &mock.OutputStub{
				SaveRoundsInfoCalled: func(_ *outport.RoundsInfo) error {
					extraOutputCalled = true
					return nil
				},
			},
		},
	}

	dp, err := CreateDataProcessor(args)
	require.Nil(t, err)

	err = dp.SaveRoundsInfo(&outport.RoundsInfo{})
	require.Nil(t, err)
	require.True(t, extraOutputCalled)
	require.True(t, elasticWrittenAfterExtraOutput)
}

func TestCreateElasticOutput(t *testing.T) {
	t.Parallel()

	output, err := CreateElasticOutput(createMockArgDataProcessorFactory())
	require.Nil(t, err)
	require.NotNil(t, output)
}
