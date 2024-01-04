package factory

import (
	errorsGo "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func createMockIndexerFactoryArgs() ArgsIndexerFactory {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	return ArgsIndexerFactory{
		Enabled:                  true,
		Url:                      ts.URL,
		UserName:                 "",
		Password:                 "",
		Marshalizer:              &mock.MarshalizerMock{},
		HeaderMarshaller:         &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		AddressPubkeyConverter:   mock.NewPubkeyConverterMock(32),
		ValidatorPubkeyConverter: &mock.PubkeyConverterMock{},
		TemplatesPath:            "../testdata",
		EnabledIndexes:           []string{"blocks", "transactions", "miniblocks", "validators", "round", "accounts", "rating"},
	}
}

func TestNewIndexerFactory(t *testing.T) {
	tests := []struct {
		name     string
		argsFunc func() ArgsIndexerFactory
		exError  error
	}{
		{
			name: "NilAddressPubkeyConverter",
			argsFunc: func() ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.AddressPubkeyConverter = nil
				return args
			},
			exError: dataindexer.ErrNilPubkeyConverter,
		},
		{
			name: "NilValidatorPubkeyConverter",
			argsFunc: func() ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.ValidatorPubkeyConverter = nil
				return args
			},
			exError: dataindexer.ErrNilPubkeyConverter,
		},
		{
			name: "NilMarshalizer",
			argsFunc: func() ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.Marshalizer = nil
				return args
			},
			exError: dataindexer.ErrNilMarshalizer,
		},
		{
			name: "NilHasher",
			argsFunc: func() ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.Hasher = nil
				return args
			},
			exError: dataindexer.ErrNilHasher,
		},
		{
			name: "NilHeaderMarshaller",
			argsFunc: func() ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.HeaderMarshaller = nil
				return args
			},
			exError: dataindexer.ErrNilMarshalizer,
		},
		{
			name: "EmptyUrl",
			argsFunc: func() ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.Url = ""
				return args
			},
			exError: dataindexer.ErrNilUrl,
		},
		{
			name: "All arguments ok",
			argsFunc: func() ArgsIndexerFactory {
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

func TestIndexerFactoryCreate_ElasticIndexer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	args := createMockIndexerFactoryArgs()
	args.Url = ts.URL

	elasticIndexer, err := NewIndexer(args)
	require.NoError(t, err)

	err = elasticIndexer.Close()
	require.NoError(t, err)

	err = elasticIndexer.Close()
	require.NoError(t, err)
}
