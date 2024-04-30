package process

import (
	"errors"
	"fmt"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-es-indexer-go/config"
	"github.com/multiversx/mx-chain-es-indexer-go/migrations"
)

var errNilMigrationHandler = errors.New("migration handler is nil")

type migrationData struct {
	handler       migrations.MigrationHandler
	migrationInfo config.Migration
}

type migrationProcessor struct {
	migrations []migrationData
}

// NewMigrationProcessor will create a new instance of migration processor
func NewMigrationProcessor() (*migrationProcessor, error) {
	return &migrationProcessor{
		migrations: make([]migrationData, 0),
	}, nil
}

// StartMigrations will start the migrations
func (mp *migrationProcessor) StartMigrations() error {
	for _, m := range mp.migrations {
		err := m.handler.DoMigration(m.migrationInfo)
		if err != nil {
			return fmt.Errorf("error processing migration %s: %s", m.migrationInfo.Name, err)
		}
	}

	return nil
}

// AddMigrationHandler will add a new migration handler
func (mp *migrationProcessor) AddMigrationHandler(handler migrations.MigrationHandler, migrationInfo config.Migration) error {
	if check.IfNil(handler) {
		return errNilMigrationHandler
	}

	mp.migrations = append(mp.migrations, migrationData{
		handler:       handler,
		migrationInfo: migrationInfo,
	})

	return nil
}
