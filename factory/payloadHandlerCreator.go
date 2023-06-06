package factory

import (
	"github.com/multiversx/mx-chain-communication-go/websocket"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	factoryHasher "github.com/multiversx/mx-chain-core-go/hashing/factory"
	"github.com/multiversx/mx-chain-core-go/marshal"
	factoryMarshaller "github.com/multiversx/mx-chain-core-go/marshal/factory"
	"github.com/multiversx/mx-chain-es-indexer-go/config"
	"github.com/multiversx/mx-chain-es-indexer-go/process/factory"
	"github.com/multiversx/mx-chain-es-indexer-go/process/wsindexer"
)

type payloadHandlerCreator struct {
	cfg        config.Config
	clusterCfg config.ClusterConfig
	importDB   bool
	marshaller marshal.Marshalizer
}

func newPayloadHandlerCreator(cfg config.Config, clusterCfg config.ClusterConfig, importDB bool, marshaller marshal.Marshalizer) (*payloadHandlerCreator, error) {
	return &payloadHandlerCreator{
		cfg:        cfg,
		clusterCfg: clusterCfg,
		importDB:   importDB,
		marshaller: marshaller,
	}, nil
}

// Create will create a new instance of PayloadHandler
func (phc *payloadHandlerCreator) Create() (websocket.PayloadHandler, error) {
	dataIndexer, err := phc.createDataIndexer()
	if err != nil {
		return nil, err
	}

	return wsindexer.NewIndexer(phc.marshaller, dataIndexer)
}

func (phc *payloadHandlerCreator) createDataIndexer() (wsindexer.DataIndexer, error) {
	marshaller, err := factoryMarshaller.NewMarshalizer(phc.cfg.Config.Marshaller.Type)
	if err != nil {
		return nil, err
	}
	hasher, err := factoryHasher.NewHasher(phc.cfg.Config.Hasher.Type)
	if err != nil {
		return nil, err
	}
	addressPubkeyConverter, err := pubkeyConverter.NewBech32PubkeyConverter(phc.cfg.Config.AddressConverter.Length, phc.cfg.Config.AddressConverter.Prefix)
	if err != nil {
		return nil, err
	}
	validatorPubkeyConverter, err := pubkeyConverter.NewHexPubkeyConverter(phc.cfg.Config.ValidatorKeysConverter.Length)
	if err != nil {
		return nil, err
	}

	return factory.NewIndexer(factory.ArgsIndexerFactory{
		UseKibana:                phc.clusterCfg.Config.ElasticCluster.UseKibana,
		IndexerCacheSize:         indexerCacheSize,
		Denomination:             phc.cfg.Config.Economics.Denomination,
		BulkRequestMaxSize:       phc.clusterCfg.Config.ElasticCluster.BulkRequestMaxSizeInBytes,
		Url:                      phc.clusterCfg.Config.ElasticCluster.URL,
		UserName:                 phc.clusterCfg.Config.ElasticCluster.UserName,
		Password:                 phc.clusterCfg.Config.ElasticCluster.Password,
		EnabledIndexes:           prepareIndices(phc.cfg.Config.AvailableIndices, phc.clusterCfg.Config.DisabledIndices),
		Marshalizer:              marshaller,
		Hasher:                   hasher,
		AddressPubkeyConverter:   addressPubkeyConverter,
		ValidatorPubkeyConverter: validatorPubkeyConverter,
		HeaderMarshaller:         phc.marshaller,
		ImportDB:                 phc.importDB,
	})
}

func prepareIndices(availableIndices, disabledIndices []string) []string {
	indices := make([]string, 0)

	mapDisabledIndices := make(map[string]struct{})
	for _, index := range disabledIndices {
		mapDisabledIndices[index] = struct{}{}
	}

	for _, availableIndex := range availableIndices {
		_, shouldSkip := mapDisabledIndices[availableIndex]
		if shouldSkip {
			continue
		}
		indices = append(indices, availableIndex)
	}

	return indices
}

func (phc *payloadHandlerCreator) IsInterfaceNil() bool {
	return phc == nil
}
