package elasticproc

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/core/request"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	elasticIndexer "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/tags"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/tokeninfo"
	"github.com/multiversx/mx-chain-es-indexer-go/templates"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var (
	log = logger.GetOrCreate("indexer/process")

	indexes = []string{
		elasticIndexer.TransactionsIndex, elasticIndexer.BlockIndex, elasticIndexer.MiniblocksIndex, elasticIndexer.RatingIndex, elasticIndexer.RoundsIndex, elasticIndexer.ValidatorsIndex,
		elasticIndexer.AccountsIndex, elasticIndexer.AccountsHistoryIndex, elasticIndexer.ReceiptsIndex, elasticIndexer.ScResultsIndex, elasticIndexer.AccountsESDTHistoryIndex, elasticIndexer.AccountsESDTIndex,
		elasticIndexer.EpochInfoIndex, elasticIndexer.SCDeploysIndex, elasticIndexer.TokensIndex, elasticIndexer.TagsIndex, elasticIndexer.LogsIndex, elasticIndexer.DelegatorsIndex, elasticIndexer.OperationsIndex,
		elasticIndexer.ESDTsIndex, elasticIndexer.ValuesIndex, elasticIndexer.EventsIndex, elasticIndexer.ExecutionResultsIndex,
	}
)

const (
	versionStr             = "indexer-version"
	minNumWritesInParallel = 1
)

// ArgElasticProcessor holds all dependencies required by the elasticProcessor in order to create
// new instances
type ArgElasticProcessor struct {
	NumWritesInParallel int
	BulkRequestMaxSize  int
	UseKibana           bool
	ImportDB            bool
	EnabledIndexes      map[string]struct{}
	TransactionsProc    DBTransactionsHandler
	AccountsProc        DBAccountHandler
	BlockProc           DBBlockHandler
	MiniblocksProc      DBMiniblocksHandler
	StatisticsProc      DBStatisticsHandler
	ValidatorsProc      DBValidatorsHandler
	DBClient            DatabaseClientHandler
	LogsAndEventsProc   DBLogsAndEventsHandler
	OperationsProc      OperationsHandler
	MappingsHandler     TemplatesAndPoliciesHandler
	Version             string
}

type elasticProcessor struct {
	numWritesInParallel int
	bulkRequestMaxSize  int
	importDB            bool
	enabledIndexes      map[string]struct{}
	mutex               sync.RWMutex
	elasticClient       DatabaseClientHandler
	accountsProc        DBAccountHandler
	blockProc           DBBlockHandler
	transactionsProc    DBTransactionsHandler
	miniblocksProc      DBMiniblocksHandler
	statisticsProc      DBStatisticsHandler
	validatorsProc      DBValidatorsHandler
	logsAndEventsProc   DBLogsAndEventsHandler
	operationsProc      OperationsHandler
	mappingsHandler     TemplatesAndPoliciesHandler
}

// NewElasticProcessor handles Elasticsearch operations such as initialization, adding, modifying or removing data
func NewElasticProcessor(arguments *ArgElasticProcessor) (*elasticProcessor, error) {
	err := checkArguments(arguments)
	if err != nil {
		return nil, err
	}
	numWritesInParallel := arguments.NumWritesInParallel
	if numWritesInParallel < minNumWritesInParallel {
		log.Warn("elasticProcessor.NewElasticProcessor: provided num writes in parallel is invalid, the minimum value will be set", "min value", minNumWritesInParallel)
		numWritesInParallel = minNumWritesInParallel
	}

	ei := &elasticProcessor{
		elasticClient:       arguments.DBClient,
		enabledIndexes:      arguments.EnabledIndexes,
		accountsProc:        arguments.AccountsProc,
		blockProc:           arguments.BlockProc,
		miniblocksProc:      arguments.MiniblocksProc,
		transactionsProc:    arguments.TransactionsProc,
		statisticsProc:      arguments.StatisticsProc,
		validatorsProc:      arguments.ValidatorsProc,
		logsAndEventsProc:   arguments.LogsAndEventsProc,
		operationsProc:      arguments.OperationsProc,
		bulkRequestMaxSize:  arguments.BulkRequestMaxSize,
		mappingsHandler:     arguments.MappingsHandler,
		numWritesInParallel: numWritesInParallel,
	}

	err = ei.init()
	if err != nil {
		return nil, err
	}

	err = ei.indexVersion(arguments.Version)

	return ei, err
}

// TODO move all the index create part in a new component
func (ei *elasticProcessor) init() error {
	indexTemplates, _, err := ei.mappingsHandler.GetElasticTemplatesAndPolicies()
	if err != nil {
		return err
	}

	err = ei.createIndexTemplates(indexTemplates)
	if err != nil {
		return err
	}

	err = ei.createIndexes()
	if err != nil {
		return err
	}

	err = ei.createAliases()
	if err != nil {
		return err
	}

	extraMappings, err := ei.mappingsHandler.GetTimestampMsMappings()
	if err != nil {
		return err
	}

	return ei.addExtraMappings(extraMappings)
}

func (ei *elasticProcessor) addExtraMappings(extraMappings []templates.ExtraMapping) error {
	for _, mappingsTuple := range extraMappings {
		err := ei.elasticClient.PutMappings(mappingsTuple.Index, mappingsTuple.Mappings)
		if err != nil {
			log.Warn("cannot add extra mappings", "index", mappingsTuple.Index, "error", err)
		}
	}

	return nil
}

func (ei *elasticProcessor) indexVersion(version string) error {
	if version == "" {
		log.Debug("ei.elasticProcessor indexer version is empty")
		return nil
	}

	keyValueObj := &data.KeyValueObj{
		Key:   versionStr,
		Value: version,
	}

	meta := []byte(fmt.Sprintf(`{ "index" : { "_index":"%s", "_id" : "%s" } }%s`, elasticIndexer.ValuesIndex, versionStr, "\n"))
	keyValueObjBytes, err := json.Marshal(keyValueObj)
	if err != nil {
		return err
	}

	buffSlice := data.NewBufferSlice(0)
	err = buffSlice.PutData(meta, keyValueObjBytes)
	if err != nil {
		return err
	}

	return ei.elasticClient.DoBulkRequest(context.Background(), buffSlice.Buffers()[0], "")
}

func (ei *elasticProcessor) createIndexTemplates(indexTemplates map[string]*bytes.Buffer) error {
	for _, index := range indexes {
		indexTemplate := getTemplateByName(index, indexTemplates)
		if indexTemplate != nil {
			err := ei.elasticClient.CheckAndCreateTemplate(index, indexTemplate)
			if err != nil {
				return fmt.Errorf("index: %s, error: %w", index, err)
			}
		}
	}
	return nil
}

func (ei *elasticProcessor) createIndexes() error {

	for _, index := range indexes {
		indexName := fmt.Sprintf("%s-%s", index, elasticIndexer.IndexSuffix)
		err := ei.elasticClient.CheckAndCreateIndex(indexName)
		if err != nil {
			return fmt.Errorf("index: %s, error: %w", index, err)
		}
	}
	return nil
}

func (ei *elasticProcessor) createAliases() error {
	for _, index := range indexes {
		indexName := fmt.Sprintf("%s-%s", index, elasticIndexer.IndexSuffix)
		err := ei.elasticClient.CheckAndCreateAlias(index, indexName)
		if err != nil {
			return err
		}
	}

	return nil
}

func getTemplateByName(templateName string, templateList map[string]*bytes.Buffer) *bytes.Buffer {
	if template, ok := templateList[templateName]; ok {
		return template
	}

	log.Debug("elasticProcessor.getTemplateByName", "could not find template", templateName)
	return nil
}

// SaveHeader will prepare and save information about a header in elasticsearch server
func (ei *elasticProcessor) SaveHeader(outportBlockWithHeader *outport.OutportBlockWithHeader) error {
	blockResults, err := ei.blockProc.PrepareBlockForDB(outportBlockWithHeader)
	if err != nil {
		return err
	}

	buffSlice := data.NewBufferSlice(ei.bulkRequestMaxSize)
	err = ei.indexBlock(blockResults.Block, buffSlice)
	if err != nil {
		return err
	}

	err = ei.indexEpochInfoData(outportBlockWithHeader.Header, buffSlice)
	if err != nil {
		return err
	}

	err = ei.indexExecutionResults(blockResults.ExecutionResults, buffSlice)
	if err != nil {
		return err
	}

	return ei.doBulkRequests("", buffSlice.Buffers(), outportBlockWithHeader.ShardID)
}

func (ei *elasticProcessor) indexBlock(esBlock *data.Block, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(elasticIndexer.BlockIndex) {
		return nil
	}

	return ei.blockProc.SerializeBlock(esBlock, buffSlice, elasticIndexer.BlockIndex)
}

func (ei *elasticProcessor) indexExecutionResults(executionResults []*data.ExecutionResult, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(elasticIndexer.ExecutionResultsIndex) {
		return nil
	}

	return ei.blockProc.SerializeExecutionResults(executionResults, buffSlice, elasticIndexer.ExecutionResultsIndex)
}

func (ei *elasticProcessor) indexEpochInfoData(header coreData.HeaderHandler, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(elasticIndexer.EpochInfoIndex) ||
		header.GetShardID() != core.MetachainShardId {
		return nil
	}

	return ei.blockProc.SerializeEpochInfoData(header, buffSlice, elasticIndexer.EpochInfoIndex)
}

// RemoveHeader will remove a block from elasticsearch server
func (ei *elasticProcessor) RemoveHeader(header coreData.HeaderHandler) error {
	headerHash, err := ei.blockProc.ComputeHeaderHash(header)
	if err != nil {
		return err
	}

	ctxWithValue := context.WithValue(context.Background(), request.ContextKey, request.ExtendTopicWithShardID(request.RemoveTopic, header.GetShardID()))
	err = ei.elasticClient.DoQueryRemove(
		ctxWithValue,
		elasticIndexer.BlockIndex,
		converters.PrepareHashesForQueryRemove([]string{hex.EncodeToString(headerHash)}),
	)
	if err != nil {
		return err
	}

	if len(header.GetExecutionResultsHandlers()) == 0 {
		return nil
	}

	executionResultsHashes := make([]string, 0)
	for _, executonResult := range header.GetExecutionResultsHandlers() {
		executionResultsHashes = append(executionResultsHashes, hex.EncodeToString(executonResult.GetHeaderHash()))
	}

	return ei.elasticClient.DoQueryRemove(
		ctxWithValue,
		elasticIndexer.ExecutionResultsIndex,
		converters.PrepareHashesForQueryRemove(executionResultsHashes),
	)
}

// RemoveMiniblocks will remove all miniblocks that are in header from elasticsearch server
func (ei *elasticProcessor) RemoveMiniblocks(header coreData.HeaderHandler, body *block.Body) error {
	encodedMiniblocksHashes := ei.miniblocksProc.GetMiniblocksHashesHexEncoded(header, body)
	if len(encodedMiniblocksHashes) == 0 {
		return nil
	}

	ctxWithValue := context.WithValue(context.Background(), request.ContextKey, request.ExtendTopicWithShardID(request.RemoveTopic, header.GetShardID()))
	return ei.elasticClient.DoQueryRemove(
		ctxWithValue,
		elasticIndexer.MiniblocksIndex,
		converters.PrepareHashesForQueryRemove(encodedMiniblocksHashes),
	)
}

// RemoveTransactions will remove transaction that are in miniblock from the elasticsearch server
func (ei *elasticProcessor) RemoveTransactions(header coreData.HeaderHandler, body *block.Body, timestampMs uint64) error {
	headerData := &data.HeaderData{
		Timestamp:        header.GetTimeStamp(),
		TimestampMs:      timestampMs,
		Round:            header.GetRound(),
		ShardID:          header.GetShardID(),
		Epoch:            header.GetEpoch(),
		MiniBlockHeaders: header.GetMiniBlockHeaderHandlers(),
	}
	encodedTxsHashes, encodedScrsHashes := ei.transactionsProc.GetHexEncodedHashesForRemove(headerData, body)
	shardID := header.GetShardID()

	err := ei.removeIfHashesNotEmpty(elasticIndexer.TransactionsIndex, encodedTxsHashes, shardID)
	if err != nil {
		return err
	}

	err = ei.removeIfHashesNotEmpty(elasticIndexer.ScResultsIndex, encodedScrsHashes, shardID)
	if err != nil {
		return err
	}

	err = ei.removeIfHashesNotEmpty(elasticIndexer.OperationsIndex, append(encodedTxsHashes, encodedScrsHashes...), shardID)
	if err != nil {
		return err
	}

	err = ei.removeIfHashesNotEmpty(elasticIndexer.LogsIndex, append(encodedTxsHashes, encodedScrsHashes...), shardID)
	if err != nil {
		return err
	}

	err = ei.removeFromIndexByTimestampAndShardID(header.GetShardID(), elasticIndexer.EventsIndex, timestampMs)
	if err != nil {
		return err
	}

	return ei.updateDelegatorsInCaseOfRevert(header, body, timestampMs)
}

func (ei *elasticProcessor) updateDelegatorsInCaseOfRevert(header coreData.HeaderHandler, body *block.Body, timestampMs uint64) error {
	// delegators index should be updated in case of revert only if the observer is in Metachain and the reverted block has miniblocks
	isMeta := header.GetShardID() == core.MetachainShardId
	hasMiniblocks := len(body.MiniBlocks) > 0
	shouldUpdate := isMeta && hasMiniblocks
	if !shouldUpdate {
		return nil
	}

	ctxWithValue := context.WithValue(context.Background(), request.ContextKey, request.ExtendTopicWithShardID(request.UpdateTopic, header.GetShardID()))

	delegatorsQuery := ei.logsAndEventsProc.PrepareDelegatorsQueryInCaseOfRevert(timestampMs)
	return ei.elasticClient.UpdateByQuery(ctxWithValue, elasticIndexer.DelegatorsIndex, delegatorsQuery)
}

func (ei *elasticProcessor) removeIfHashesNotEmpty(index string, hashes []string, shardID uint32) error {
	if len(hashes) == 0 {
		return nil
	}

	ctxWithValue := context.WithValue(context.Background(), request.ContextKey, request.ExtendTopicWithShardID(request.RemoveTopic, shardID))
	return ei.elasticClient.DoQueryRemove(
		ctxWithValue,
		index,
		converters.PrepareHashesForQueryRemove(hashes),
	)
}

// RemoveAccountsESDT will remove data from accountsesdt index and accountsesdthistory
func (ei *elasticProcessor) RemoveAccountsESDT(shardID uint32, timestampMs uint64) error {
	err := ei.removeFromIndexByTimestampAndShardID(shardID, elasticIndexer.AccountsESDTIndex, timestampMs)
	if err != nil {
		return err
	}

	return ei.removeFromIndexByTimestampAndShardID(shardID, elasticIndexer.AccountsESDTHistoryIndex, timestampMs)
}

func (ei *elasticProcessor) removeFromIndexByTimestampAndShardID(shardID uint32, index string, timestampMs uint64) error {
	ctxWithValue := context.WithValue(context.Background(), request.ContextKey, request.ExtendTopicWithShardID(request.RemoveTopic, shardID))
	query := fmt.Sprintf(`{"query": {"bool": {"must": [{"match": {"shardID": {"query": %d,"operator": "AND"}}},{"match": {"timestampMs": {"query": "%d","operator": "AND"}}}]}}}`, shardID, timestampMs)

	return ei.elasticClient.DoQueryRemove(
		ctxWithValue,
		index,
		bytes.NewBuffer([]byte(query)),
	)
}

// SaveMiniblocks will prepare and save information about miniblocks in elasticsearch server
func (ei *elasticProcessor) SaveMiniblocks(header coreData.HeaderHandler, miniBlocks []*block.MiniBlock, timestampMS uint64) error {
	if !ei.isIndexEnabled(elasticIndexer.MiniblocksIndex) {
		return nil
	}

	mbs := ei.miniblocksProc.PrepareDBMiniblocks(header, miniBlocks, timestampMS)
	if len(mbs) == 0 {
		return nil
	}

	buffSlice := data.NewBufferSlice(ei.bulkRequestMaxSize)
	ei.miniblocksProc.SerializeBulkMiniBlocks(mbs, buffSlice, elasticIndexer.MiniblocksIndex, header.GetShardID())

	return ei.doBulkRequests("", buffSlice.Buffers(), header.GetShardID())
}

// SaveTransactions will prepare and save information about a transactions in elasticsearch server
func (ei *elasticProcessor) SaveTransactions(obh *outport.OutportBlockWithHeader) error {
	miniBlocks := append(obh.BlockData.Body.MiniBlocks, obh.BlockData.IntraShardMiniBlocks...)
	headerData := &data.HeaderData{
		Timestamp:        converters.MillisecondsToSeconds(obh.BlockData.TimestampMs),
		TimestampMs:      obh.BlockData.TimestampMs,
		Round:            obh.Header.GetRound(),
		ShardID:          obh.Header.GetShardID(),
		Epoch:            obh.Header.GetEpoch(),
		MiniBlockHeaders: obh.Header.GetMiniBlockHeaderHandlers(),
		NumberOfShards:   obh.NumberOfShards,
	}

	buffers := data.NewBufferSlice(ei.bulkRequestMaxSize)
	err := ei.prepareAndSaveTransactionsData(headerData, miniBlocks, obh.TransactionPool, obh.AlteredAccounts, buffers)
	if err != nil {
		return err
	}

	for _, executionResult := range obh.Header.GetExecutionResultsHandlers() {
		executionResulData, found := obh.BlockData.Results[hex.EncodeToString(executionResult.GetHeaderHash())]
		if !found {
			log.Warn("elasticProcessor.SaveTransactions: cannot find execution result data", "hash", executionResult.GetHeaderHash())
			continue
		}

		headerData = &data.HeaderData{
			Timestamp:        converters.MillisecondsToSeconds(executionResulData.TimestampMs),
			TimestampMs:      executionResulData.TimestampMs,
			Round:            executionResult.GetHeaderRound(),
			ShardID:          obh.Header.GetShardID(),
			NumberOfShards:   obh.NumberOfShards,
			Epoch:            executionResult.GetHeaderEpoch(),
			MiniBlockHeaders: converters.GetMiniBlocksHeaderHandlersFromExecResult(executionResult),
		}

		miniBlocks = append(executionResulData.Body.MiniBlocks, executionResulData.IntraShardMiniBlocks...)
		err = ei.prepareAndSaveTransactionsData(headerData, miniBlocks, executionResulData.TransactionPool, executionResulData.AlteredAccounts, buffers)
		if err != nil {
			return err
		}
	}

	return ei.doBulkRequests("", buffers.Buffers(), headerData.ShardID)
}

func (ei *elasticProcessor) prepareAndSaveTransactionsData(
	headerData *data.HeaderData,
	miniBlocks []*block.MiniBlock,
	pool *outport.TransactionPool,
	alteredAccounts map[string]*alteredAccount.AlteredAccount,
	buffers *data.BufferSlice,
) error {

	preparedResults := ei.transactionsProc.PrepareTransactionsForDatabase(miniBlocks, headerData, pool, ei.isImportDB(), headerData.NumberOfShards)
	logsData := ei.logsAndEventsProc.ExtractDataFromLogs(pool.Logs, preparedResults, headerData.ShardID, headerData.NumberOfShards, headerData.TimestampMs)

	err := ei.indexTransactions(preparedResults.Transactions, logsData.TxHashStatusInfo, headerData.ShardID, buffers)
	if err != nil {
		return err
	}

	err = ei.prepareAndIndexOperations(preparedResults.Transactions, logsData.TxHashStatusInfo, headerData.ShardID, preparedResults.ScResults, buffers, ei.isImportDB())
	if err != nil {
		return err
	}

	err = ei.indexTransactionsFeeData(preparedResults.TxHashFee, buffers)
	if err != nil {
		return err
	}

	err = ei.indexNFTCreateInfo(logsData.Tokens, alteredAccounts, buffers, headerData.ShardID)
	if err != nil {
		return err
	}

	err = ei.indexLogs(logsData.DBLogs, buffers)
	if err != nil {
		return err
	}

	err = ei.indexEvents(logsData.DBEvents, buffers)
	if err != nil {
		return err
	}

	err = ei.indexScResults(preparedResults.ScResults, buffers)
	if err != nil {
		return err
	}

	err = ei.indexReceipts(preparedResults.Receipts, buffers)
	if err != nil {
		return err
	}

	tagsCount := tags.NewTagsCount()
	err = ei.indexAlteredAccounts(logsData.NFTsDataUpdates, alteredAccounts, buffers, tagsCount, headerData.ShardID, headerData.TimestampMs)
	if err != nil {
		return err
	}

	err = ei.prepareAndIndexTagsCount(tagsCount, buffers)
	if err != nil {
		return err
	}

	err = ei.indexTokens(logsData.TokensInfo, logsData.NFTsDataUpdates, buffers, headerData.ShardID)
	if err != nil {
		return err
	}

	err = ei.prepareAndIndexDelegators(logsData.Delegators, buffers)
	if err != nil {
		return err
	}

	err = ei.indexNFTBurnInfo(logsData.TokensSupply, buffers, headerData.ShardID)
	if err != nil {
		return err
	}

	err = ei.prepareAndIndexRolesData(logsData.TokenRolesAndProperties, buffers, elasticIndexer.TokensIndex)
	if err != nil {
		return err
	}
	err = ei.prepareAndIndexRolesData(logsData.TokenRolesAndProperties, buffers, elasticIndexer.ESDTsIndex)
	if err != nil {
		return err
	}

	return ei.indexScDeploys(logsData.ScDeploys, logsData.ChangeOwnerOperations, buffers)
}

func (ei *elasticProcessor) prepareAndIndexRolesData(tokenRolesAndProperties *tokeninfo.TokenRolesAndProperties, buffSlice *data.BufferSlice, index string) error {
	if !ei.isIndexEnabled(index) {
		return nil
	}

	return ei.logsAndEventsProc.SerializeRolesData(tokenRolesAndProperties, buffSlice, index)
}

func (ei *elasticProcessor) prepareAndIndexDelegators(delegators map[string]*data.Delegator, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(elasticIndexer.DelegatorsIndex) {
		return nil
	}

	return ei.logsAndEventsProc.SerializeDelegators(delegators, buffSlice, elasticIndexer.DelegatorsIndex)
}

func (ei *elasticProcessor) indexTransactionsFeeData(txsHashFeeData map[string]*data.FeeData, buffSlice *data.BufferSlice) error {
	if len(txsHashFeeData) == 0 {
		return nil
	}

	err := ei.transactionsProc.SerializeTransactionsFeeData(txsHashFeeData, buffSlice, elasticIndexer.TransactionsIndex)
	if err != nil {
		return nil
	}

	return ei.transactionsProc.SerializeTransactionsFeeData(txsHashFeeData, buffSlice, elasticIndexer.OperationsIndex)
}

func (ei *elasticProcessor) indexLogs(logsDB []*data.Logs, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(elasticIndexer.LogsIndex) {
		return nil
	}

	return ei.logsAndEventsProc.SerializeLogs(logsDB, buffSlice, elasticIndexer.LogsIndex)
}

func (ei *elasticProcessor) indexEvents(eventsDB []*data.LogEvent, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(elasticIndexer.EventsIndex) {
		return nil
	}

	return ei.logsAndEventsProc.SerializeEvents(eventsDB, buffSlice, elasticIndexer.EventsIndex)
}

func (ei *elasticProcessor) indexScDeploys(deployData map[string]*data.ScDeployInfo, changeOwnerOperation map[string]*data.OwnerData, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(elasticIndexer.SCDeploysIndex) {
		return nil
	}

	err := ei.logsAndEventsProc.SerializeSCDeploys(deployData, buffSlice, elasticIndexer.SCDeploysIndex)
	if err != nil {
		return err
	}

	return ei.logsAndEventsProc.SerializeChangeOwnerOperations(changeOwnerOperation, buffSlice, elasticIndexer.SCDeploysIndex)
}

func (ei *elasticProcessor) indexTransactions(txs []*data.Transaction, txHashStatusInfo map[string]*outport.StatusInfo, shardID uint32, bytesBuff *data.BufferSlice) error {
	if !ei.isIndexEnabled(elasticIndexer.TransactionsIndex) {
		return nil
	}

	return ei.transactionsProc.SerializeTransactions(txs, txHashStatusInfo, shardID, bytesBuff, elasticIndexer.TransactionsIndex)
}

func (ei *elasticProcessor) prepareAndIndexOperations(
	txs []*data.Transaction,
	txHashStatusInfo map[string]*outport.StatusInfo,
	shardID uint32,
	scrs []*data.ScResult,
	buffSlice *data.BufferSlice,
	isImportDB bool,
) error {
	if !ei.isIndexEnabled(elasticIndexer.OperationsIndex) {
		return nil
	}

	processedTxs, processedSCRs := ei.operationsProc.ProcessTransactionsAndSCRs(txs, scrs, isImportDB, shardID)

	err := ei.transactionsProc.SerializeTransactions(processedTxs, txHashStatusInfo, shardID, buffSlice, elasticIndexer.OperationsIndex)
	if err != nil {
		return err
	}

	return ei.operationsProc.SerializeSCRs(processedSCRs, buffSlice, elasticIndexer.OperationsIndex, shardID)
}

// SaveValidatorsRating will save validators rating
func (ei *elasticProcessor) SaveValidatorsRating(ratingData *outport.ValidatorsRating) error {
	if !ei.isIndexEnabled(elasticIndexer.RatingIndex) {
		return nil
	}

	buffSlice, err := ei.validatorsProc.SerializeValidatorsRating(ratingData)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(elasticIndexer.RatingIndex, buffSlice, ratingData.ShardID)
}

// SaveShardValidatorsPubKeys will prepare and save information about a shard validators public keys in elasticsearch server
func (ei *elasticProcessor) SaveShardValidatorsPubKeys(validatorsPubKeys *outport.ValidatorsPubKeys) error {
	if !ei.isIndexEnabled(elasticIndexer.ValidatorsIndex) {
		return nil
	}

	buffSlice, err := ei.validatorsProc.PrepareAnSerializeValidatorsPubKeys(validatorsPubKeys)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(elasticIndexer.ValidatorsIndex, buffSlice, validatorsPubKeys.ShardID)
}

// SaveRoundsInfo will prepare and save information about a slice of rounds in elasticsearch server
func (ei *elasticProcessor) SaveRoundsInfo(rounds *outport.RoundsInfo) error {
	if !ei.isIndexEnabled(elasticIndexer.RoundsIndex) {
		return nil
	}

	buff := ei.statisticsProc.SerializeRoundsInfo(rounds)

	ctxWithValue := context.WithValue(context.Background(), request.ContextKey, request.ExtendTopicWithShardID(request.BulkTopic, rounds.ShardID))
	return ei.elasticClient.DoBulkRequest(ctxWithValue, buff, elasticIndexer.RoundsIndex)
}

func (ei *elasticProcessor) indexAlteredAccounts(
	updatesNFTsData []*data.NFTDataUpdate,
	coreAlteredAccounts map[string]*alteredAccount.AlteredAccount,
	buffSlice *data.BufferSlice,
	tagsCount data.CountTags,
	shardID uint32,
	timestampMs uint64,
) error {
	regularAccountsToIndex, accountsToIndexESDT := ei.accountsProc.GetAccounts(coreAlteredAccounts)

	err := ei.saveAccounts(regularAccountsToIndex, buffSlice, shardID, timestampMs)
	if err != nil {
		return err
	}

	return ei.saveAccountsESDT(accountsToIndexESDT, updatesNFTsData, buffSlice, tagsCount, shardID, timestampMs)
}

func (ei *elasticProcessor) saveAccountsESDT(
	wrappedAccounts []*data.AccountESDT,
	updatesNFTsData []*data.NFTDataUpdate,
	buffSlice *data.BufferSlice,
	tagsCount data.CountTags,
	shardID uint32,
	timestampMs uint64,
) error {
	accountsESDTMap, tokensData := ei.accountsProc.PrepareAccountsMapESDT(wrappedAccounts, tagsCount, shardID, timestampMs)
	err := ei.addTokenTypeAndCurrentOwnerInAccountsESDT(tokensData, accountsESDTMap, shardID)
	if err != nil {
		return err
	}

	err = ei.indexAccountsESDT(accountsESDTMap, updatesNFTsData, buffSlice)
	if err != nil {
		return err
	}

	return ei.saveAccountsESDTHistory(accountsESDTMap, buffSlice, shardID, timestampMs)
}

func (ei *elasticProcessor) addTokenTypeAndCurrentOwnerInAccountsESDT(tokensData data.TokensHandler, accountsESDTMap map[string]*data.AccountInfo, shardID uint32) error {
	if check.IfNil(tokensData) || tokensData.Len() == 0 {
		return nil
	}

	responseTokens := &data.ResponseTokens{}
	ctxWithValue := context.WithValue(context.Background(), request.ContextKey, request.ExtendTopicWithShardID(request.GetTopic, shardID))
	err := ei.elasticClient.DoMultiGet(ctxWithValue, tokensData.GetAllTokens(), elasticIndexer.TokensIndex, true, responseTokens)
	if err != nil {
		return err
	}

	tokensData.AddTypeAndOwnerFromResponse(responseTokens)
	tokensData.PutTypeAndOwnerInAccountsESDT(accountsESDTMap)

	return nil
}

func (ei *elasticProcessor) prepareAndIndexTagsCount(tagsCount data.CountTags, buffSlice *data.BufferSlice) error {
	shouldSkipIndex := !ei.isIndexEnabled(elasticIndexer.TagsIndex) || tagsCount.Len() == 0
	if shouldSkipIndex {
		return nil
	}

	return tagsCount.Serialize(buffSlice, elasticIndexer.TagsIndex)
}

func (ei *elasticProcessor) indexAccountsESDT(
	accountsESDTMap map[string]*data.AccountInfo,
	updatesNFTsData []*data.NFTDataUpdate,
	buffSlice *data.BufferSlice,
) error {
	if !ei.isIndexEnabled(elasticIndexer.AccountsESDTIndex) {
		return nil
	}

	return ei.accountsProc.SerializeAccountsESDT(accountsESDTMap, updatesNFTsData, buffSlice, elasticIndexer.AccountsESDTIndex)
}

func (ei *elasticProcessor) indexNFTCreateInfo(tokensData data.TokensHandler, coreAlteredAccounts map[string]*alteredAccount.AlteredAccount, buffSlice *data.BufferSlice, shardID uint32) error {
	shouldSkipIndex := !ei.isIndexEnabled(elasticIndexer.TokensIndex) || tokensData.Len() == 0
	if shouldSkipIndex {
		return nil
	}

	ctxWithValue := context.WithValue(context.Background(), request.ContextKey, request.ExtendTopicWithShardID(request.GetTopic, shardID))
	responseTokens := &data.ResponseTokens{}
	err := ei.elasticClient.DoMultiGet(ctxWithValue, tokensData.GetAllTokens(), elasticIndexer.TokensIndex, true, responseTokens)
	if err != nil {
		return err
	}

	tokensData.AddTypeAndOwnerFromResponse(responseTokens)

	tokens := tokensData.GetAllWithoutMetaESDT()
	ei.accountsProc.PutTokenMedataDataInTokens(tokens, coreAlteredAccounts)

	return ei.accountsProc.SerializeNFTCreateInfo(tokens, buffSlice, elasticIndexer.TokensIndex)
}

func (ei *elasticProcessor) indexNFTBurnInfo(tokensData data.TokensHandler, buffSlice *data.BufferSlice, shardID uint32) error {
	shouldSkipIndex := !ei.isIndexEnabled(elasticIndexer.TokensIndex) || tokensData.Len() == 0
	if shouldSkipIndex {
		return nil
	}

	ctxWithValue := context.WithValue(context.Background(), request.ContextKey, request.ExtendTopicWithShardID(request.GetTopic, shardID))
	responseTokens := &data.ResponseTokens{}
	err := ei.elasticClient.DoMultiGet(ctxWithValue, tokensData.GetAllTokens(), elasticIndexer.TokensIndex, true, responseTokens)
	if err != nil {
		return err
	}

	// TODO implement to keep in tokens also the supply
	tokensData.AddTypeAndOwnerFromResponse(responseTokens)
	return ei.logsAndEventsProc.SerializeSupplyData(tokensData, buffSlice, elasticIndexer.TokensIndex)
}

// SaveAccounts will prepare and save information about provided accounts in elasticsearch server
func (ei *elasticProcessor) SaveAccounts(accountsData *outport.Accounts) error {
	buffSlice := data.NewBufferSlice(ei.bulkRequestMaxSize)

	accounts := make([]*data.Account, 0, len(accountsData.AlteredAccounts))
	for _, account := range accountsData.AlteredAccounts {
		accounts = append(accounts, &data.Account{
			UserAccount: account,
			IsSender:    false,
		})
	}

	return ei.saveAccounts(accounts, buffSlice, accountsData.ShardID, accountsData.BlockTimestampMs)
}

func (ei *elasticProcessor) saveAccounts(accts []*data.Account, buffSlice *data.BufferSlice, shardID uint32, timestampMs uint64) error {
	accountsMap := ei.accountsProc.PrepareRegularAccountsMap(accts, shardID, timestampMs)
	err := ei.indexAccounts(accountsMap, elasticIndexer.AccountsIndex, buffSlice)
	if err != nil {
		return err
	}

	return ei.saveAccountsHistory(accountsMap, buffSlice, shardID, timestampMs)
}

func (ei *elasticProcessor) indexAccounts(accountsMap map[string]*data.AccountInfo, index string, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(index) {
		return nil
	}

	return ei.serializeAndIndexAccounts(accountsMap, index, buffSlice)
}

func (ei *elasticProcessor) serializeAndIndexAccounts(accountsMap map[string]*data.AccountInfo, index string, buffSlice *data.BufferSlice) error {
	return ei.accountsProc.SerializeAccounts(accountsMap, buffSlice, index)
}

func (ei *elasticProcessor) saveAccountsESDTHistory(accountsInfoMap map[string]*data.AccountInfo, buffSlice *data.BufferSlice, shardID uint32, timestampMs uint64) error {
	if !ei.isIndexEnabled(elasticIndexer.AccountsESDTHistoryIndex) {
		return nil
	}

	accountsMap := ei.accountsProc.PrepareAccountsHistory(accountsInfoMap, shardID, timestampMs)

	return ei.serializeAndIndexAccountsHistory(accountsMap, elasticIndexer.AccountsESDTHistoryIndex, buffSlice)
}

func (ei *elasticProcessor) saveAccountsHistory(accountsInfoMap map[string]*data.AccountInfo, buffSlice *data.BufferSlice, shardID uint32, timestampMS uint64) error {
	if !ei.isIndexEnabled(elasticIndexer.AccountsHistoryIndex) {
		return nil
	}

	accountsMap := ei.accountsProc.PrepareAccountsHistory(accountsInfoMap, shardID, timestampMS)

	return ei.serializeAndIndexAccountsHistory(accountsMap, elasticIndexer.AccountsHistoryIndex, buffSlice)
}

func (ei *elasticProcessor) serializeAndIndexAccountsHistory(accountsMap map[string]*data.AccountBalanceHistory, index string, buffSlice *data.BufferSlice) error {
	return ei.accountsProc.SerializeAccountsHistory(accountsMap, buffSlice, index)
}

func (ei *elasticProcessor) indexScResults(scrs []*data.ScResult, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(elasticIndexer.ScResultsIndex) {
		return nil
	}

	return ei.transactionsProc.SerializeScResults(scrs, buffSlice, elasticIndexer.ScResultsIndex)
}

func (ei *elasticProcessor) indexReceipts(receipts []*data.Receipt, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(elasticIndexer.ReceiptsIndex) {
		return nil
	}

	return ei.transactionsProc.SerializeReceipts(receipts, buffSlice, elasticIndexer.ReceiptsIndex)
}

func (ei *elasticProcessor) isIndexEnabled(index string) bool {
	_, isEnabled := ei.enabledIndexes[index]
	return isEnabled
}

func (ei *elasticProcessor) doBulkRequests(index string, buffSlice []*bytes.Buffer, shardID uint32) error {
	jobs := make(chan *bytes.Buffer)
	errCh := make(chan error, len(buffSlice))

	var wg sync.WaitGroup

	for i := 0; i < ei.numWritesInParallel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for buf := range jobs {
				ctx := context.WithValue(context.Background(), request.ContextKey, request.ExtendTopicWithShardID(request.BulkTopic, shardID))
				err := ei.elasticClient.DoBulkRequest(ctx, buf, index)
				if err != nil {
					errCh <- err
				}
			}
		}()
	}

	go func() {
		for _, buf := range buffSlice {
			jobs <- buf
		}
		close(jobs)
	}()

	wg.Wait()
	close(errCh)

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// SetOutportConfig will set the outport config
func (ei *elasticProcessor) SetOutportConfig(cfg outport.OutportConfig) error {
	ei.mutex.Lock()
	defer ei.mutex.Unlock()

	ei.importDB = cfg.IsInImportDBMode

	return nil
}

func (ei *elasticProcessor) isImportDB() bool {
	ei.mutex.RLock()
	defer ei.mutex.RUnlock()

	return ei.importDB
}

// IsInterfaceNil returns true if there is no value under the interface
func (ei *elasticProcessor) IsInterfaceNil() bool {
	return ei == nil
}
