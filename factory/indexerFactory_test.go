package factory

import (
	errorsGo "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/stretchr/testify/require"
)

func createMockIndexerFactoryArgs() *ArgsIndexerFactory {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	return &ArgsIndexerFactory{
		Enabled:                  true,
		IndexerCacheSize:         100,
		Url:                      ts.URL,
		UserName:                 "",
		Password:                 "",
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		AddressPubkeyConverter:   mock.NewPubkeyConverterMock(32),
		ValidatorPubkeyConverter: &mock.PubkeyConverterMock{},
		TemplatesPath:            "../testdata",
		EnabledIndexes:           []string{"blocks", "transactions", "miniblocks", "validators", "round", "accounts", "rating"},
		AccountsDB:               &mock.AccountsStub{},
		TransactionFeeCalculator: &mock.EconomicsHandlerStub{},
		ShardCoordinator:         &mock.ShardCoordinatorMock{},
		IsInImportDBMode:         false,
	}
}

func TestNewIndexerFactory(t *testing.T) {
	tests := []struct {
		name     string
		argsFunc func() *ArgsIndexerFactory
		exError  error
	}{
		{
			name: "InvalidCacheSize",
			argsFunc: func() *ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.IndexerCacheSize = -1
				return args
			},
			exError: indexer.ErrNegativeCacheSize,
		},
		{
			name: "NilAddressPubkeyConverter",
			argsFunc: func() *ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.AddressPubkeyConverter = nil
				return args
			},
			exError: indexer.ErrNilPubkeyConverter,
		},
		{
			name: "NilValidatorPubkeyConverter",
			argsFunc: func() *ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.ValidatorPubkeyConverter = nil
				return args
			},
			exError: indexer.ErrNilPubkeyConverter,
		},
		{
			name: "NilMarshalizer",
			argsFunc: func() *ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.Marshalizer = nil
				return args
			},
			exError: indexer.ErrNilMarshalizer,
		},
		{
			name: "NilHasher",
			argsFunc: func() *ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.Hasher = nil
				return args
			},
			exError: indexer.ErrNilHasher,
		},
		{
			name: "NilAccountsDB",
			argsFunc: func() *ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.AccountsDB = nil
				return args
			},
			exError: indexer.ErrNilAccountsDB,
		},
		{
			name: "EmptyUrl",
			argsFunc: func() *ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.Url = ""
				return args
			},
			exError: indexer.ErrNilUrl,
		},
		{
			name: "NilTransactionFeeCalculator",
			argsFunc: func() *ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.TransactionFeeCalculator = nil
				return args
			},
			exError: indexer.ErrNilTransactionFeeCalculator,
		},
		{
			name: "NilShardCoordinator",
			argsFunc: func() *ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.ShardCoordinator = nil
				return args
			},
			exError: indexer.ErrNilShardCoordinator,
		},
		{
			name: "All arguments ok",
			argsFunc: func() *ArgsIndexerFactory {
				return createMockIndexerFactoryArgs()
			},
			exError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewIndexer(tt.argsFunc())
			require.True(t, errorsGo.Is(err, tt.exError))
		})
	}
}

func TestIndexerFactoryCreate_NilIndexer(t *testing.T) {
	t.Parallel()

	args := createMockIndexerFactoryArgs()
	args.Enabled = false
	nilIndexer, err := NewIndexer(args)
	require.NoError(t, err)

	_, ok := nilIndexer.(*indexer.NilIndexer)
	require.True(t, ok)
}

func TestIndexerFactoryCreate_ElasticIndexer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	args := createMockIndexerFactoryArgs()
	args.Url = ts.URL

	elasticIndexer, err := NewIndexer(args)
	require.NoError(t, err)

	err = elasticIndexer.Close()
	require.NoError(t, err)
	require.False(t, elasticIndexer.IsNilIndexer())

	err = elasticIndexer.Close()
	require.NoError(t, err)
}
