package migrations

import "github.com/multiversx/mx-chain-es-indexer-go/config"

type MigrationHandler interface {
	DoMigration(migrationInfo config.Migration) error
}

type MigrationProcessor interface {
	StartMigrations(migrationNames []string) error
}
