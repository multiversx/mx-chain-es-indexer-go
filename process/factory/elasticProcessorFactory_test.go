package factory

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateElasticProcessor(t *testing.T) {
	esClient := &mock.DatabaseWriterStub{}
	args := ArgElasticProcessorFactory{
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		AddressPubkeyConverter:   &mock.PubkeyConverterMock{},
		ValidatorPubkeyConverter: &mock.PubkeyConverterMock{},
		DBClient:                 esClient,
		AccountsDB:               &mock.AccountsStub{},
		ShardCoordinator:         &mock.ShardCoordinatorMock{},
		TransactionFeeCalculator: &mock.EconomicsHandlerStub{},
		EnabledIndexes:           []string{"blocks"},
		Denomination:             1,
		IsInImportDBMode:         false,
		UseKibana:                false,
	}

	ep, err := CreateElasticProcessor(args)
	require.Nil(t, err)
	require.NotNil(t, ep)
}
