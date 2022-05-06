package process

import (
	"bytes"
	"fmt"

	elasticIndexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
)

// ArgPostgresProcessor
type ArgPostgresProcessor struct {
	UseKibana         bool
	SelfShardID       uint32
	IndexTemplates    map[string]*bytes.Buffer
	IndexPolicies     map[string]*bytes.Buffer
	EnabledIndexes    map[string]struct{}
	TransactionsProc  DBTransactionsHandler
	AccountsProc      DBAccountHandler
	BlockProc         DBBlockHandler
	MiniblocksProc    DBMiniblocksHandler
	StatisticsProc    DBStatisticsHandler
	ValidatorsProc    DBValidatorsHandler
	DBClient          PostgresClientHandler
	LogsAndEventsProc DBLogsAndEventsHandler
}

type postgresProcessor struct {
	selfShardID       uint32
	enabledIndexes    map[string]struct{}
	postgresClient    PostgresClientHandler
	accountsProc      DBAccountHandler
	blockProc         DBBlockHandler
	transactionsProc  DBTransactionsHandler
	miniblocksProc    DBMiniblocksHandler
	statisticsProc    DBStatisticsHandler
	validatorsProc    DBValidatorsHandler
	logsAndEventsProc DBLogsAndEventsHandler
}

// NewPostgresProcessor
func NewPostgresProcessor(arguments *ArgPostgresProcessor) (*postgresProcessor, error) {
	err := checkPostgresProcessorArgs(arguments)
	if err != nil {
		return nil, err
	}

	ei := &postgresProcessor{
		postgresClient:    arguments.DBClient,
		enabledIndexes:    arguments.EnabledIndexes,
		accountsProc:      arguments.AccountsProc,
		blockProc:         arguments.BlockProc,
		miniblocksProc:    arguments.MiniblocksProc,
		transactionsProc:  arguments.TransactionsProc,
		selfShardID:       arguments.SelfShardID,
		statisticsProc:    arguments.StatisticsProc,
		validatorsProc:    arguments.ValidatorsProc,
		logsAndEventsProc: arguments.LogsAndEventsProc,
	}

	err = ei.init(arguments.UseKibana, arguments.IndexTemplates, arguments.IndexPolicies)
	if err != nil {
		return nil, err
	}

	return ei, nil
}

func checkPostgresProcessorArgs(arguments *ArgPostgresProcessor) error {
	if arguments == nil {
		return elasticIndexer.ErrNilElasticProcessorArguments
	}
	if arguments.EnabledIndexes == nil {
		return elasticIndexer.ErrNilEnabledIndexesMap
	}
	if arguments.DBClient == nil {
		return elasticIndexer.ErrNilDatabaseClient
	}
	if arguments.StatisticsProc == nil {
		return elasticIndexer.ErrNilStatisticHandler
	}
	if arguments.BlockProc == nil {
		return elasticIndexer.ErrNilBlockHandler
	}
	if arguments.AccountsProc == nil {
		return elasticIndexer.ErrNilAccountsHandler
	}
	if arguments.MiniblocksProc == nil {
		return elasticIndexer.ErrNilMiniblocksHandler
	}

	if arguments.ValidatorsProc == nil {
		return elasticIndexer.ErrNilValidatorsHandler
	}

	if arguments.TransactionsProc == nil {
		return elasticIndexer.ErrNilTransactionsHandler
	}

	if arguments.LogsAndEventsProc == nil {
		return elasticIndexer.ErrNilLogsAndEventsHandler
	}

	return nil
}

func (ps *postgresProcessor) init(useKibana bool, indexTemplates, _ map[string]*bytes.Buffer) error {
	err := ps.createTables()
	if err != nil {
		return err
	}

	fmt.Println("init has been executed successfully")

	return nil
}

func (ps *postgresProcessor) createTables() error {
	err := ps.postgresClient.AutoMigrateTables(
		// Accounts
		&data.AccountInfo{},
		&data.TokenMetaData{},
		&data.AccountBalanceHistory{},
		//&data.AccountESDT{},
		//&data.Account{},

		// Block
		&data.Block{},
		&data.ScheduledData{},
		&data.EpochStartInfo{},
		&data.Miniblock{},

		// Data
		&data.ValidatorsPublicKeys{},
		&data.ValidatorRatingInfo{},
		&data.RoundInfo{},
		&data.EpochInfo{},

		// Delegators
		&data.Delegator{},

		// Logs
		&data.Logs{},
		&data.Event{},

		// ScDeploy
		&data.ScDeployInfo{},

		// Transactions
		&data.Transaction{},
		&data.ScResult{},
		&data.Receipt{},
	)
	if err != nil {
		return err
	}

	return nil
}

// SaveHeader will prepare and save information about a header in elasticsearch server
func (psp *postgresProcessor) SaveHeader(
	header coreData.HeaderHandler,
	signersIndexes []uint64,
	body *block.Body,
	notarizedHeadersHashes []string,
	gasConsumptionData indexer.HeaderGasConsumption,
	txsSize int,
) error {
	// if !ei.isIndexEnabled(elasticIndexer.BlockIndex) {
	// 	return nil
	// }

	elasticBlock, err := psp.blockProc.PrepareBlockForDB(header, signersIndexes, body, notarizedHeadersHashes, gasConsumptionData, txsSize)
	if err != nil {
		return err
	}

	err = psp.postgresClient.Insert(elasticBlock)
	if err != nil {
		return err
	}

	err = psp.indexEpochInfoData(header)
	if err != nil {
		return err
	}

	return nil
}

func (psp *postgresProcessor) indexEpochInfoData(header coreData.HeaderHandler) error {
	if psp.selfShardID != core.MetachainShardId {
		return nil
	}

	if check.IfNil(header) {
		return elasticIndexer.ErrNilHeaderHandler
	}

	metablock, ok := header.(*block.MetaBlock)
	if !ok {
		return fmt.Errorf("%w in blockProcessor.SerializeEpochInfoData", elasticIndexer.ErrHeaderTypeAssertion)
	}

	epochInfo := &data.EpochInfo{
		AccumulatedFees: metablock.AccumulatedFeesInEpoch.String(),
		DeveloperFees:   metablock.DevFeesInEpoch.String(),
	}

	return psp.postgresClient.Insert(epochInfo)
}

// RemoveHeader will remove a block from elasticsearch server
func (ei *postgresProcessor) RemoveHeader(header coreData.HeaderHandler) error {
	// headerHash, err := ei.blockProc.ComputeHeaderHash(header)
	// if err != nil {
	// 	return err
	// }

	return nil
}

// RemoveMiniblocks will remove all miniblocks that are in header from elasticsearch server
func (ei *postgresProcessor) RemoveMiniblocks(header coreData.HeaderHandler, body *block.Body) error {
	// encodedMiniblocksHashes := ei.miniblocksProc.GetMiniblocksHashesHexEncoded(header, body)
	// if len(encodedMiniblocksHashes) == 0 {
	// 	return nil
	// }

	return nil
}

// RemoveTransactions will remove transaction that are in miniblock from the elasticsearch server
func (ei *postgresProcessor) RemoveTransactions(header coreData.HeaderHandler, body *block.Body) error {
	// encodedTxsHashes := ei.transactionsProc.GetRewardsTxsHashesHexEncoded(header, body)
	// if len(encodedTxsHashes) == 0 {
	// 	return nil
	// }

	return nil
}

// SaveMiniblocks will prepare and save information about miniblocks in elasticsearch server
func (ei *postgresProcessor) SaveMiniblocks(header coreData.HeaderHandler, body *block.Body) error {
	// if !ei.isIndexEnabled(elasticIndexer.MiniblocksIndex) {
	// 	return nil
	// }

	mbs := ei.miniblocksProc.PrepareDBMiniblocks(header, body)
	if len(mbs) == 0 {
		return nil
	}

	err := ei.postgresClient.Insert(mbs)
	if err != nil {
		return err
	}

	return nil
}

// SaveTransactions will prepare and save information about a transactions in elasticsearch server
func (psp *postgresProcessor) SaveTransactions(
	body *block.Body,
	header coreData.HeaderHandler,
	pool *indexer.Pool,
) error {
	headerTimestamp := header.GetTimeStamp()

	preparedResults := psp.transactionsProc.PrepareTransactionsForDatabase(body, header, pool)
	logsData := psp.logsAndEventsProc.ExtractDataFromLogs(pool.Logs, preparedResults, headerTimestamp)

	err := psp.indexTransactions(preparedResults.Transactions)
	if err != nil {
		return err
	}

	err = psp.indexAlteredAccounts(headerTimestamp, preparedResults.AlteredAccts, logsData.PendingBalances)
	if err != nil {
		return err
	}

	return nil
}

func (psp *postgresProcessor) indexTransactions(txs []*data.Transaction) error {
	for _, tx := range txs {
		err := psp.postgresClient.Insert(tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (psp *postgresProcessor) indexAlteredAccounts(
	timestamp uint64,
	alteredAccounts data.AlteredAccountsHandler,
	pendingBalances map[string]*data.AccountInfo,
) error {
	regularAccountsToIndex, accountsToIndexESDT := psp.accountsProc.GetAccounts(alteredAccounts)

	err := psp.SaveAccounts(timestamp, regularAccountsToIndex)
	if err != nil {
		return err
	}

	accountsESDTMap := psp.accountsProc.PrepareAccountsMapESDT(accountsToIndexESDT)
	resAccountsMap := converters.MergeAccountsInfoMaps(accountsESDTMap, pendingBalances)

	err = psp.indexAccounts(resAccountsMap)
	if err != nil {
		return err
	}

	err = psp.saveAccountsHistory(timestamp, accountsESDTMap)
	if err != nil {
		return err
	}

	return nil
}

// SaveValidatorsRating will save validators rating
func (ei *postgresProcessor) SaveValidatorsRating(index string, validatorsRatingInfo []*data.ValidatorRatingInfo) error {
	// if !ei.isIndexEnabled(elasticIndexer.RatingIndex) {
	// 	return nil
	// }

	err := ei.postgresClient.Insert(validatorsRatingInfo)
	if err != nil {
		return err
	}

	return nil
}

// SaveShardValidatorsPubKeys will prepare and save information about a shard validators public keys in elasticsearch server
func (ei *postgresProcessor) SaveShardValidatorsPubKeys(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error {
	// if !ei.isIndexEnabled(elasticIndexer.ValidatorsIndex) {
	// 	return nil
	// }

	validatorsPubKeys := ei.validatorsProc.PrepareValidatorsPublicKeys(shardValidatorsPubKeys)

	err := ei.postgresClient.Insert(validatorsPubKeys)
	if err != nil {
		return err
	}

	return nil
}

// SaveRoundsInfo will prepare and save information about a slice of rounds in elasticsearch server
func (ei *postgresProcessor) SaveRoundsInfo(info []*data.RoundInfo) error {
	// if !ei.isIndexEnabled(elasticIndexer.RoundsIndex) {
	// 	return nil
	// }

	err := ei.postgresClient.Insert(info)
	if err != nil {
		return err
	}

	return nil
}

// SaveAccounts will prepare and save information about provided accounts in elasticsearch server
func (psp *postgresProcessor) SaveAccounts(timestamp uint64, accts []*data.Account) error {
	accountsMap := psp.accountsProc.PrepareRegularAccountsMap(accts)

	err := psp.indexAccounts(accountsMap)
	if err != nil {
		return err
	}

	err = psp.saveAccountsHistory(timestamp, accountsMap)
	if err != nil {
		return err
	}

	return nil
}

func (psp *postgresProcessor) indexAccounts(accountsMap map[string]*data.AccountInfo) error {
	var err error
	for _, acc := range accountsMap {
		err = psp.postgresClient.Insert(acc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (psp *postgresProcessor) saveAccountsHistory(timestamp uint64, accountsInfoMap map[string]*data.AccountInfo) error {
	accountsMap := psp.accountsProc.PrepareAccountsHistory(timestamp, accountsInfoMap)

	var err error
	for _, acc := range accountsMap {
		err = psp.postgresClient.Insert(acc)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ei *postgresProcessor) IsInterfaceNil() bool {
	return ei == nil
}
