package factory

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc"
)

func TestCreateElasticProcessor(t *testing.T) {

	args := ArgElasticProcessorFactory{
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		AddressPubkeyConverter:   mock.NewPubkeyConverterMock(32),
		ValidatorPubkeyConverter: &mock.PubkeyConverterMock{},
		DBClient:                 &mock.DatabaseWriterStub{},
		EnabledIndexes:           []string{"blocks"},
		Denomination:             1,
		UseKibana:                false,
		TxHashExtractor:          &mock.TxHashExtractorMock{},
		RewardTxData:             &mock.RewardTxDataMock{},
		IndexTokensHandler:       &elasticproc.IndexTokenHandlerMock{},
	}

	ep, err := CreateElasticProcessor(args)
	require.Nil(t, err)
	require.NotNil(t, ep)
}
