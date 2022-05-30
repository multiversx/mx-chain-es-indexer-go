package postgres

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

const dsn = "host=localhost user=postgres password=mysecretpassword dbname=elrondv3 port=5432 sslmode=disable"

const batchSize = 1000

type ArgsPostgresClient struct {
	Hostname string
	Port     int
	Username string
	Password string
	DBName   string
}

type postgresClient struct {
	dsn string
	ps  *gorm.DB
}

func NewPostgresClient(args *ArgsPostgresClient) (*postgresClient, error) {
	// TODO: check args
	dsn := fmt.Sprintf("port=%d host=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		args.Port, args.Hostname, args.Username, args.Password, args.DBName,
	)

	postgresLogger, _ := newPostgresLogger()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: postgresLogger,
	})
	if err != nil {
		return nil, err
	}

	schema.RegisterSerializer("base64", Base64Serializer{})

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

	err = pc.createAccountsTable()
	if err != nil {
		return err
	}

	err = pc.createESDTAccountsTable()
	if err != nil {
		return err
	}

	err = pc.createTokenMetaDataTable()
	if err != nil {
		return err
	}

	err = pc.createAccountsHistoryTable()
	if err != nil {
		return err
	}

	err = pc.createAccountsESDTHistoryTable()
	if err != nil {
		return err
	}

	err = pc.createTagsTable()
	if err != nil {
		return err
	}

	err = pc.createScrsOperationsTable()
	if err != nil {
		return err
	}

	err = pc.createTxsOperationsTable()
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

func (pc *postgresClient) createAccountsTable() error {
	sql := `CREATE TABLE IF NOT EXISTS accounts (
		address text NOT NULL UNIQUE,
		nonce bigint,
		balance text,
		balance_num decimal,
		token_name text,
		token_identifier text,
		token_nonce bigint,
		properties text,
		total_balance_with_stake text,
		total_balance_with_stake_num decimal,
		PRIMARY KEY (address)
	)`

	return pc.CreateRawTable(sql)
}

func (pc *postgresClient) createESDTAccountsTable() error {
	sql := `CREATE TABLE IF NOT EXISTS accounts_esdt (
		address text NOT NULL,
		nonce bigint,
		balance text,
		balance_num decimal,
		token_name text NOT NULL,
		token_identifier text,
		token_nonce bigint NOT NULL,
		properties text,
		total_balance_with_stake text,
		total_balance_with_stake_num decimal,
		PRIMARY KEY (address, token_name, token_nonce)
	)`

	return pc.CreateRawTable(sql)
}

func (pc *postgresClient) createTokenMetaDataTable() error {
	sql := `CREATE TABLE IF NOT EXISTS token_meta_data (
		name text NOT NULL UNIQUE,
		creator text,
		royalties bigint,
		hash text,
		uris text,
		tags text[],
		attributes text,
		meta_data text,
		non_empty_uris boolean,
		white_listed_storage boolean,
		address text,
		token_name text,
		token_nonce bigint,
		PRIMARY KEY (name),
		FOREIGN KEY (address, token_name, token_nonce) REFERENCES accounts_esdt(address, token_name, token_nonce)
	)`

	return pc.CreateRawTable(sql)
}

func (pc *postgresClient) createAccountsHistoryTable() error {
	sql := `CREATE TABLE IF NOT EXISTS accounts_history (
		address text,
		timestamp int8,
		balance text,
		token text,
		identifier text,
		token_nonce int8,
		is_sender bool,
		is_smart_contract bool,
		PRIMARY KEY (address, timestamp)
	)`

	return pc.CreateRawTable(sql)
}

func (pc *postgresClient) createAccountsESDTHistoryTable() error {
	sql := `CREATE TABLE IF NOT EXISTS accounts_esdt_history (
		address text,
		timestamp int8,
		balance text,
		token text,
		identifier text,
		token_nonce int8,
		is_sender bool,
		is_smart_contract bool,
		PRIMARY KEY (address, token, token_nonce, timestamp)
	)`

	return pc.CreateRawTable(sql)
}

func (pc *postgresClient) createTagsTable() error {
	sql := `CREATE TABLE IF NOT EXISTS tags (
		tag text NOT NULL UNIQUE,
		count integer,
		PRIMARY KEY (tag)
	)`

	return pc.CreateRawTable(sql)
}

func (pc *postgresClient) createTxsOperationsTable() error {
	err := pc.ps.Table("txs_operations").AutoMigrate(&data.Transaction{})
	if err != nil {
		return err
	}

	return nil
}

func (pc *postgresClient) createScrsOperationsTable() error {
	err := pc.ps.Table("scrs_operations").AutoMigrate(&data.ScResult{})
	if err != nil {
		return err
	}

	return nil
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
	log.Info("Insert", "type", reflect.TypeOf(entity))
	result := pc.ps.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(entity, batchSize)
	if result.Error != nil {
		return result.Error
	}

	log.Info("Insert", "rows affected", result.RowsAffected)

	return nil
}

func (pc *postgresClient) InsertBlock(block *data.Block) error {
	result := pc.ps.Model(&data.Block{}).Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(block, batchSize)
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

func (pc *postgresClient) InsertAccount(account *data.AccountInfo) error {
	sql := `INSERT INTO accounts_esdt (
		address, nonce, balance, balance_num, token_name, token_identifier, token_nonce, properties, total_balance_with_stake, total_balance_with_stake_num
	) VALUES(
		?,?,?,?,?,?,?,?,?,?
	) ON CONFLICT DO NOTHING`

	result := pc.ps.Exec(sql,
		account.Address,
		account.Nonce,
		account.Balance,
		account.BalanceNum,
		account.TokenName,
		account.TokenIdentifier,
		account.TokenNonce,
		account.Properties,
		account.TotalBalanceWithStake,
		account.TotalBalanceWithStakeNum,
	)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (pc *postgresClient) InsertAccountESDT(id string, account *data.AccountInfo) error {
	sql := `INSERT INTO accounts_esdt (
		address,
		nonce,
		balance,
		balance_num,
		token_name,
		token_identifier,
		token_nonce,
		properties,
		total_balance_with_stake,
		total_balance_with_stake_num
	) VALUES(
		?,?,?,?,?,?,?,?,?,?
	) ON CONFLICT DO NOTHING`

	result := pc.ps.Exec(sql,
		account.Address,
		account.Nonce,
		account.Balance,
		account.BalanceNum,
		account.TokenName,
		account.TokenIdentifier,
		account.TokenNonce,
		account.Properties,
		account.TotalBalanceWithStake,
		account.TotalBalanceWithStakeNum,
	)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (pc *postgresClient) InsertESDTMetaData(account *data.AccountInfo) error {
	sql := `INSERT INTO token_meta_data (
		name,
		creator,
		royalties,
		hash,
		uris,
		tags,
		attributes,
		meta_data,
		non_empty_uris,
		white_listed_storage,
		address,
		token_name,
		token_nonce
	) VALUES(
		?,?,?,?,?,?,?,?,?,?,?,?,?
	) ON CONFLICT DO NOTHING`

	uris, err := json.Marshal(account.Data.URIs)
	if err != nil {
		return err
	}

	tags := make([]string, 0)
	for _, tag := range account.Data.Tags {
		tags = append(tags, base64.StdEncoding.EncodeToString([]byte(tag)))
	}

	result := pc.ps.Exec(sql,
		account.Data.Name,
		account.Data.Creator,
		account.Data.Royalties,
		base64.StdEncoding.EncodeToString(account.Data.Hash),
		base64.StdEncoding.EncodeToString(uris),
		pq.StringArray(tags),
		base64.StdEncoding.EncodeToString(account.Data.Attributes),
		base64.StdEncoding.EncodeToString([]byte(account.Data.MetaData)),
		account.Data.NonEmptyURIs,
		account.Data.WhiteListedStorage,
		account.Address,
		account.TokenName,
		account.TokenNonce,
	)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (pc *postgresClient) InsertAccountHistory(account *data.AccountBalanceHistory) error {
	sql := `INSERT INTO accounts_history (
		address,
		timestamp,
		balance,
		token,
		identifier,
		token_nonce,
		is_sender,
		is_smart_contract
	) VALUES(
		?,?,?,?,?,?,?,?
	) ON CONFLICT DO NOTHING`

	result := pc.ps.Exec(sql,
		account.Address,
		account.Timestamp,
		account.Balance,
		account.Token,
		account.Identifier,
		account.TokenNonce,
		account.IsSender,
		account.IsSmartContract,
	)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (pc *postgresClient) InsertAccountESDTHistory(account *data.AccountBalanceHistory) error {
	sql := `INSERT INTO accounts_esdt_history (
		address,
		timestamp,
		balance,
		token,
		identifier,
		token_nonce,
		is_sender,
		is_smart_contract
	) VALUES (
		?,?,?,?,?,?,?,?
	) ON CONFLICT DO NOTHING`

	result := pc.ps.Exec(sql,
		account.Address,
		account.Timestamp,
		account.Balance,
		account.Token,
		account.Identifier,
		account.TokenNonce,
		account.IsSender,
		account.IsSmartContract,
	)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (pc *postgresClient) InsertTags(tags map[string]int) error {
	sql := `INSERT INTO tags (
		tag, count
	) VALUES`

	vals := []interface{}{}
	for tag, count := range tags {
		sql += "(?, ?),"
		vals = append(vals, tag, count)
	}

	// trim the last ,
	sql = sql[0 : len(sql)-1]

	sql += " ON CONFLICT (tag) DO UPDATE SET count = tags.count + 1"

	result := pc.ps.Exec(sql, vals...)
	if result.Error != nil {
		return result.Error
	}

	log.Info("Insert tags", "rows affected", result.RowsAffected)

	return nil
}

func (pc *postgresClient) InsertScrsOperation(scResults []*data.ScResult) error {
	result := pc.ps.Table("scrs_operations").Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(scResults, batchSize)
	if result.Error != nil {
		return result.Error
	}

	log.Info("Insert", "rows affected", result.RowsAffected)

	return nil
}

func (pc *postgresClient) InsertTxsOperation(txs []*data.Transaction) error {
	result := pc.ps.Table("txs_operations").Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(txs, batchSize)
	if result.Error != nil {
		return result.Error
	}

	log.Info("Insert", "rows affected", result.RowsAffected)

	return nil
}

func (pc *postgresClient) IsInterfaceNil() bool {
	return pc == nil
}
