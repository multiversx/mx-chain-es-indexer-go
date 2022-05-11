package postgres

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormLogger "gorm.io/gorm/logger"
)

var log = logger.GetOrCreate("indexer/postgres")

const dsn = "host=localhost user=postgres password=mysecretpassword dbname=elrondv2 port=5432 sslmode=disable"

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

func (pc *postgresClient) CreateTables() error {
	err := pc.createEpochStartInfoTable()
	if err != nil {
		return err
	}

	err = pc.createValidatorRatingInfoTable()
	if err != nil {
		return err
	}

	err = pc.createValidatorPubKeysTable()
	if err != nil {
		return err
	}

	err = pc.createEpochInfoTable()
	if err != nil {
		return err
	}

	return nil
}

func (pc *postgresClient) createEpochStartInfoTable() error {
	sql := `CREATE TABLE IF NOT EXISTS "epoch_start_infos" (
		total_supply text,
		total_to_distribute text,
		total_newly_minted text,
		rewards_per_block text,
		rewards_for_protocol_sustainability text,
		node_price text,
		prev_epoch_start_round bigint,
		prev_epoch_start_hash text,
		hash text NOT NULL UNIQUE,
		FOREIGN KEY (hash) REFERENCES blocks(hash)
	)`

	return pc.CreateRawTable(sql)
}

func (pc *postgresClient) createValidatorRatingInfoTable() error {
	sql := `CREATE TABLE IF NOT EXISTS validator_rating_infos (
		id text NOT NULL UNIQUE,
		rating numeric,
		PRIMARY KEY (id)
	)`

	return pc.CreateRawTable(sql)
}

func (pc *postgresClient) createValidatorPubKeysTable() error {
	sql := `CREATE TABLE IF NOT EXISTS validator_public_keys (
		id text NOT NULL UNIQUE,
		pub_keys text[],
		PRIMARY KEY (id)
	)`

	return pc.CreateRawTable(sql)
}

func (pc *postgresClient) createEpochInfoTable() error {
	sql := `CREATE TABLE IF NOT EXISTS epoch_info (
		epoch bigint NOT NULL UNIQUE,
		accumulated_fees text,
		developer_fees text,
		PRIMARY KEY (epoch)
	)`

	return pc.CreateRawTable(sql)
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

func (pc *postgresClient) CreateRawTable(sql string) error {
	result := pc.ps.Exec(sql)
	if result.Error != nil {
		return result.Error
	}

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
	result := pc.ps.Clauses(clause.OnConflict{DoNothing: true}).Create(entity)
	if result.Error != nil {
		return result.Error
	}

	log.Info("Insert", "rows affected", result.RowsAffected)

	return nil
}

func (pc *postgresClient) InsertBlock(block *data.Block) error {
	result := pc.ps.Model(&data.Block{}).Clauses(clause.OnConflict{DoNothing: true}).Create(block)
	if result.Error != nil {
		return result.Error
	}

	log.Info("Insert Block", "rows affected", result.RowsAffected)

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

func (pc *postgresClient) Raw(sql string, values ...interface{}) error {
	result := pc.ps.Raw(sql, values)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (pc *postgresClient) Exec(sql string, values ...interface{}) error {
	result := pc.ps.Exec(sql, values)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (pc *postgresClient) InsertEpochStartInfo(block *data.Block) error {
	sql := `INSERT INTO epoch_start_infos (
		hash,total_supply,total_to_distribute,total_newly_minted,rewards_per_block,rewards_for_protocol_sustainability,node_price,prev_epoch_start_round,prev_epoch_start_hash
	) VALUES(
		?,?,?,?,?,?,?,?,?
	) ON CONFLICT DO NOTHING`

	esi := block.EpochStartInfo

	result := pc.ps.Exec(sql, block.Hash, esi.TotalSupply, esi.TotalToDistribute, esi.TotalNewlyMinted, esi.RewardsPerBlock, esi.RewardsForProtocolSustainability, esi.NodePrice, esi.PrevEpochStartRound, esi.PrevEpochStartHash)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (pc *postgresClient) InsertValidatorsRating(id string, ratingInfo *data.ValidatorRatingInfo) error {
	sql := `INSERT INTO validator_rating_infos (
		id, rating
	) VALUES(
		?,?
	) ON CONFLICT DO NOTHING`

	result := pc.ps.Exec(sql, id, ratingInfo.Rating)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (pc *postgresClient) InsertValidatorsPubKeys(id string, pubKeys *data.ValidatorsPublicKeys) error {
	sql := `INSERT INTO validator_public_keys (
		id, pub_keys
	) VALUES(
		?,?
	) ON CONFLICT DO NOTHING`

	result := pc.ps.Exec(sql, id, pq.Array(pubKeys.PublicKeys))
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (pc *postgresClient) InsertEpochInfo(block *block.MetaBlock) error {
	sql := `INSERT INTO epoch_info (
		epoch, accumulated_fees, developer_fees
	) VALUES(
		?,?,?
	) ON CONFLICT DO NOTHING`

	result := pc.ps.Exec(sql, block.GetEpoch(), block.AccumulatedFeesInEpoch.String(), block.DevFeesInEpoch.String())
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (pc *postgresClient) IsInterfaceNil() bool {
	return pc == nil
}
