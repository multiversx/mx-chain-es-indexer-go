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
	BulkRequestMaxSize int
	UseKibana          bool
	SelfShardID        uint32
	IndexTemplates     map[string]*bytes.Buffer
	IndexPolicies      map[string]*bytes.Buffer
	EnabledIndexes     map[string]struct{}
	TransactionsProc   DBTransactionsHandler
	AccountsProc       DBAccountHandler
	BlockProc          DBBlockHandler
	MiniblocksProc     DBMiniblocksHandler
	StatisticsProc     DBStatisticsHandler
	ValidatorsProc     DBValidatorsHandler
	DBClient           PostgresClientHandler
	LogsAndEventsProc  DBLogsAndEventsHandler
	OperationsProc     OperationsHandler
}

type postgresProcessor struct {
	bulkRequestMaxSize int
	selfShardID        uint32
	enabledIndexes     map[string]struct{}
	postgresClient     PostgresClientHandler
	accountsProc       DBAccountHandler
	blockProc          DBBlockHandler
	transactionsProc   DBTransactionsHandler
	miniblocksProc     DBMiniblocksHandler
	statisticsProc     DBStatisticsHandler
	validatorsProc     DBValidatorsHandler
	logsAndEventsProc  DBLogsAndEventsHandler
	operationsProc     OperationsHandler
}

// NewPostgresProcessor
func NewPostgresProcessor(arguments *ArgPostgresProcessor) (*postgresProcessor, error) {
	err := checkPostgresProcessorArgs(arguments)
	if err != nil {
		return nil, err
	}

	ei := &postgresProcessor{
		postgresClient:     arguments.DBClient,
		enabledIndexes:     arguments.EnabledIndexes,
		accountsProc:       arguments.AccountsProc,
		blockProc:          arguments.BlockProc,
		miniblocksProc:     arguments.MiniblocksProc,
		transactionsProc:   arguments.TransactionsProc,
		selfShardID:        arguments.SelfShardID,
		statisticsProc:     arguments.StatisticsProc,
		validatorsProc:     arguments.ValidatorsProc,
		logsAndEventsProc:  arguments.LogsAndEventsProc,
		operationsProc:     arguments.OperationsProc,
		bulkRequestMaxSize: arguments.BulkRequestMaxSize,
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

func (psp *postgresProcessor) createTables() error {
	err := psp.postgresClient.AutoMigrateTables(
		// Accounts
		// &data.AccountInfo{},
		// &data.AccountBalanceHistory{},
		//&data.AccountESDT{},
		//&data.Account{},

		// Block
		&data.Block{},
		&data.Miniblock{},

		// Data
		&data.ValidatorsPublicKeys{},
		&data.RoundInfo{},
		&data.EpochInfo{},

		// Delegators
		&data.Delegator{},

		// Logs
		&data.Logs{},
		&data.Event{},

		// ScDeploy
		&data.ScDeployInfo{},
		&data.Upgrade{},

		// Transactions
		&data.Transaction{},
		&data.ScResult{},
		&data.Receipt{},

		// Tokens
		&data.TokenInfo{},
		&data.OwnerData{},
	)
	if err != nil {
		return err
	}

	err = psp.postgresClient.CreateTables()
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

	err = psp.indexBlock(elasticBlock)
	if err != nil {
		return err
	}

	err = psp.indexEpochInfoData(header)
	if err != nil {
		return err
	}

	return nil
}

func (psp *postgresProcessor) indexBlock(block *data.Block) error {
	err := psp.postgresClient.InsertBlock(block)
	if err != nil {
		return err
	}

	return psp.indexEpochStartInfo(block)
}

func (psp *postgresProcessor) indexEpochStartInfo(block *data.Block) error {
	if psp.selfShardID != core.MetachainShardId {
		return nil
	}

	if !block.EpochStartBlock {
		return nil
	}

	return psp.postgresClient.InsertEpochStartInfo(block)
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
		return fmt.Errorf("%w in indexEpochInfoData", elasticIndexer.ErrHeaderTypeAssertion)
	}

	return psp.postgresClient.InsertEpochInfo(metablock)
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

	// err = psp.indexTransactionsWithRefund(preparedResults.TxHashRefund)
	// if err != nil {
	// 	return err
	// }

	err = psp.prepareAndIndexTagsCount(logsData.TagsCount)
	if err != nil {
		return err
	}

	err = psp.indexNFTCreateInfo(logsData.Tokens)
	if err != nil {
		return err
	}

	err = psp.prepareAndIndexLogs(pool.Logs, headerTimestamp)
	if err != nil {
		return err
	}

	err = psp.indexScResults(preparedResults.ScResults)
	if err != nil {
		return err
	}

	err = psp.indexReceipts(preparedResults.Receipts)
	if err != nil {
		return err
	}

	err = psp.indexAlteredAccounts(headerTimestamp, preparedResults.AlteredAccts, logsData.NFTsDataUpdates)
	if err != nil {
		return err
	}

	err = psp.indexTokens(logsData.TokensInfo)
	if err != nil {
		return err
	}

	err = psp.indexDelegators(logsData.Delegators)
	if err != nil {
		return err
	}

	err = psp.prepareAndIndexOperations(preparedResults.Transactions, preparedResults.TxHashStatus, header, preparedResults.ScResults)
	if err != nil {
		return err
	}

	err = psp.indexNFTBurnInfo(logsData.TokensSupply)
	if err != nil {
		return err
	}

	// err = ei.prepareAndIndexRolesData(logsData.RolesData, buffers)
	// if err != nil {
	// 	return err
	// }

	err = psp.indexScDeploys(logsData.ScDeploys)
	if err != nil {
		return err
	}

	return nil
}

func (psp *postgresProcessor) indexTransactions(txs []*data.Transaction) error {
	if len(txs) == 0 {
		return nil
	}

	return psp.postgresClient.Insert(txs)
}

func (psp *postgresProcessor) prepareAndIndexTagsCount(tagsCount data.CountTags) error {
	if tagsCount.Len() == 0 {
		return nil
	}

	tagsMap, err := tagsCount.TagsCountToPostgres()
	if err != nil {
		return err
	}

	err = psp.postgresClient.InsertTags(tagsMap)
	if err != nil {
		return err
	}

	return nil
}

func (psp *postgresProcessor) indexNFTCreateInfo(tokensData data.TokensHandler) error {
	if tokensData.Len() == 0 {
		return nil
	}

	// TODO: handle get type from response

	tokens := tokensData.GetAll()
	psp.accountsProc.PutTokenMedataDataInTokens(tokens)

	err := psp.postgresClient.Insert(tokens)
	if err != nil {
		return err
	}

	return nil
}

func (psp *postgresProcessor) indexNFTBurnInfo(tokensData data.TokensHandler) error {
	if tokensData.Len() == 0 {
		return nil
	}

	// TODO: handle get type from response

	tokensInfo := make([]*data.TokenInfo, 0)
	for _, supplyData := range tokensData.GetAll() {
		if supplyData.Type != core.NonFungibleESDT {
			continue
		}

		tokensInfo = append(tokensInfo, supplyData)
	}

	err := psp.postgresClient.Insert(tokensInfo)
	if err != nil {
		return err
	}

	return nil
}

func (psp *postgresProcessor) prepareAndIndexLogs(logsAndEvents []*coreData.LogData, timestamp uint64) error {
	if len(logsAndEvents) == 0 {
		return nil
	}

	logsDB := psp.logsAndEventsProc.PrepareLogsForDB(logsAndEvents, timestamp)

	err := psp.postgresClient.Insert(logsDB)
	if err != nil {
		return err
	}

	return nil
}

func (psp *postgresProcessor) indexScResults(scrs []*data.ScResult) error {
	if len(scrs) == 0 {
		return nil
	}

	err := psp.postgresClient.Insert(scrs)
	if err != nil {
		return err
	}

	return nil
}

func (psp *postgresProcessor) indexReceipts(receipts []*data.Receipt) error {
	if len(receipts) == 0 {
		return nil
	}

	return psp.postgresClient.Insert(receipts)
}

func (psp *postgresProcessor) indexTokens(tokensData []*data.TokenInfo) error {
	if len(tokensData) == 0 {
		return nil
	}

	return psp.postgresClient.Insert(tokensData)
}

func (psp *postgresProcessor) indexDelegators(delegators map[string]*data.Delegator) error {
	if len(delegators) == 0 {
		return nil
	}

	var err error
	for _, delegator := range delegators {
		err = psp.postgresClient.Insert(delegator)
		if err != nil {
			return err
		}
	}

	return nil
}

func (psp *postgresProcessor) prepareAndIndexOperations(
	txs []*data.Transaction,
	txHashStatus map[string]string,
	header coreData.HeaderHandler,
	scrs []*data.ScResult,
) error {
	processedTxs, processedSCRs := psp.operationsProc.ProcessTransactionsAndSCRs(txs, scrs)

	err := psp.postgresClient.InsertTxsOperation(processedTxs)
	if err != nil {
		return err
	}

	err = psp.postgresClient.InsertScrsOperation(processedSCRs)
	if err != nil {
		return err
	}

	return nil
}

func (psp *postgresProcessor) indexAlteredAccounts(
	timestamp uint64,
	alteredAccounts data.AlteredAccountsHandler,
	updatesNFTsData []*data.NFTDataUpdate,
) error {
	regularAccountsToIndex, accountsToIndexESDT := psp.accountsProc.GetAccounts(alteredAccounts)

	err := psp.SaveAccounts(timestamp, regularAccountsToIndex)
	if err != nil {
		return err
	}

	accountsESDTMap, tokensData := psp.accountsProc.PrepareAccountsMapESDT(timestamp, accountsToIndexESDT)

	// TODO: get and update data from response
	tokensData.PutTypeAndOwnerInAccountsESDT(accountsESDTMap)

	err = psp.indexAccountsESDT(accountsESDTMap)
	if err != nil {
		return err
	}

	return psp.saveAccountsESDTHistory(timestamp, accountsESDTMap)
}

func (psp *postgresProcessor) indexAccountsESDT(accountsESDTMap map[string]*data.AccountInfo) error {
	for _, acc := range accountsESDTMap {

		// handle delete account

		id := acc.Address
		hexEncodedNonce := converters.EncodeNonceToHex(acc.TokenNonce)
		id += fmt.Sprintf("-%s-%s", acc.TokenName, hexEncodedNonce)

		err := psp.postgresClient.InsertAccountESDT(id, acc)
		if err != nil {
			return err
		}

		if acc.Data != nil {
			err = psp.postgresClient.InsertESDTMetaData(acc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (psp *postgresProcessor) saveAccountsESDTHistory(timestamp uint64, accountsInfoMap map[string]*data.AccountInfo) error {
	accountsESDTMap := psp.accountsProc.PrepareAccountsHistory(timestamp, accountsInfoMap)

	for _, acc := range accountsESDTMap {
		// handle delete account

		id := acc.Address

		isESDT := acc.Token != ""
		if isESDT {
			hexEncodedNonce := converters.EncodeNonceToHex(acc.TokenNonce)
			id += fmt.Sprintf("-%s-%s", acc.Token, hexEncodedNonce)
		}

		id += fmt.Sprintf("-%d", acc.Timestamp)

		err := psp.postgresClient.InsertAccountESDTHistory(acc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (psp *postgresProcessor) indexScDeploys(deployData map[string]*data.ScDeployInfo) error {
	if len(deployData) == 0 {
		return nil
	}

	var err error
	for _, scDeploy := range deployData {
		err = psp.postgresClient.Insert(scDeploy)
		if err != nil {
			return err
		}
	}

	return nil
}

// SaveValidatorsRating will save validators rating
func (ei *postgresProcessor) SaveValidatorsRating(index string, validatorsRatingInfo []*data.ValidatorRatingInfo) error {
	// if !ei.isIndexEnabled(elasticIndexer.RatingIndex) {
	// 	return nil
	// }

	err := ei.validatorsProc.ValidatorsRatingToPostgres(ei.postgresClient, index, validatorsRatingInfo)
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

	id := fmt.Sprintf("%d_%d", shardID, epoch)
	err := ei.postgresClient.InsertValidatorsPubKeys(id, validatorsPubKeys)
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
	accountsMap := psp.accountsProc.PrepareRegularAccountsMap(timestamp, accts)

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
		err = psp.postgresClient.InsertAccount(acc)
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
		err = psp.postgresClient.InsertAccountHistory(acc)
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
