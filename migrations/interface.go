package migrations

import "github.com/multiversx/mx-chain-es-indexer-go/config"

type MigrationHandler interface {
	DoMigration(migrationInfo config.Migration) error
	IsInterfaceNil() bool
}

type MigrationProcessor interface {
	StartMigrations() error
	AddMigrationHandler(handler MigrationHandler, migrationInfo config.Migration) error
}
