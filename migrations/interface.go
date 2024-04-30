package migrations

type MigrationHandler interface {
	DoMigration(migrationName string) error
}

type MigrationProcessor interface {
	StartMigrations(migrationNames []string) error
}
