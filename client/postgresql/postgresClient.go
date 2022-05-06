package postgres

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var log = logger.GetOrCreate("indexer/postgres")

const dsn = "host=localhost user=postgres password=mysecretpassword dbname=postgres port=5432 sslmode=disable"

type postgresClient struct {
	dsn string
	ps  *gorm.DB
}

func NewPostgresClient() (*postgresClient, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	})
	if err != nil {
		return nil, err
	}

	return &postgresClient{
		ps: db,
	}, nil
}

func (pc *postgresClient) CreateTable(entity interface{}) error {
	if pc.ps.Migrator().HasTable(entity) {
		return nil
	}

	err := pc.ps.Migrator().CreateTable(entity)
	if err != nil {
		return err
	}

	log.Info("table has been created", "name")

	return nil
}

func (pc *postgresClient) AutoMigrateTables(tables ...interface{}) error {
	err := pc.ps.AutoMigrate(tables...)
	if err != nil {
		return err
	}

	log.Info("tables have been migrated")

	return nil
}

func (pc *postgresClient) Insert(entity interface{}) error {
	result := pc.ps.Create(entity)
	if result.Error != nil {
		return result.Error
	}

	log.Info("Insert", "rows affected", result.RowsAffected)

	return nil
}

func (pc *postgresClient) Delete(entity interface{}) error {
	result := pc.ps.Delete(entity)
	if result.Error != nil {
		return result.Error
	}

	log.Info("Delete", "rows affected", result.RowsAffected)

	return nil
}

func (pc *postgresClient) IsInterfaceNil() bool {
	return pc == nil
}
