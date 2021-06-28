package factory

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateElasticProcessor(t *testing.T) {

	args := ArgElasticProcessorFactory{
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		AddressPubkeyConverter:   &mock.PubkeyConverterMock{},
		ValidatorPubkeyConverter: &mock.PubkeyConverterMock{},
		DBClient:                 &mock.DatabaseWriterStub{},
		AccountsDB:               &mock.AccountsStub{},
		ShardCoordinator:         &mock.ShardCoordinatorMock{},
		TransactionFeeCalculator: &mock.EconomicsHandlerStub{},
		EnabledIndexes:           []string{"tps"},
		TemplatesPath:            "",
		Denomination:             1,
		SaveTxsLogsEnabled:       false,
		IsInImportDBMode:         false,
		UseKibana:                false,
	}

	ep, err := CreateElasticProcessor(args)
	require.Nil(t, err)
	require.NotNil(t, ep)
}