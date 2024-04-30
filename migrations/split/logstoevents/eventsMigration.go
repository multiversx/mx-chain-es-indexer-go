package logstoevents

import (
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-es-indexer-go/client"
	"github.com/multiversx/mx-chain-es-indexer-go/config"
	"github.com/multiversx/mx-chain-es-indexer-go/migrations"
)

type eventsMigration struct {
	eventsProc *eventsProcessor
}

func NewEventsMigration(sourceCluster, destinationCluster config.ClusterInfo) (migrations.MigrationHandler, error) {
	sourceClient, err := client.NewElasticClient(elasticsearch.Config{
		Addresses: []string{sourceCluster.URL},
		Username:  sourceCluster.UserName,
		Password:  sourceCluster.Password,
	})
	if err != nil {
		return nil, err
	}
	destinationClient, err := client.NewElasticClient(elasticsearch.Config{
		Addresses: []string{destinationCluster.URL},
		Username:  destinationCluster.UserName,
		Password:  destinationCluster.Password,
	})

	pubKeyConverter, err := pubkeyConverter.NewBech32PubkeyConverter(32, "erd")
	if err != nil {
		return nil, err
	}

	eventsProc, err := NewEventsProcessor(sourceClient, destinationClient, pubKeyConverter)
	if err != nil {
		return nil, err
	}

	return &eventsMigration{
		eventsProc: eventsProc,
	}, nil
}

func (e *eventsMigration) DoMigration(migration config.Migration) error {
	log.Info("starting migration", "name", migration.Name)

	return e.eventsProc.SplitLogIndexInEvents(migration.Name, migration.SourceIndex, migration.DestinationIndex)
}
