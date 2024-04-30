package process

type migrationProcessor struct {
}

func NewMigrationProcessor() (*migrationProcessor, error) {
	return &migrationProcessor{}, nil
}

func (mp *migrationProcessor) StartMigrations(migrations []string) error {
	return nil
}
