package process

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/process/tags"
	"github.com/ElrondNetwork/elastic-indexer-go/process/tokeninfo"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

const (
	docsKey  = "docs"
	errorKey = "error"
	idKey    = "_id"
	foundKey = "found"
)

var (
	log = logger.GetOrCreate("indexer/process")
)

type objectsMap = map[string]interface{}

// ArgElasticProcessor holds all dependencies required by the elasticProcessor in order to create
// new instances
type ArgElasticProcessor struct {
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
	DBClient           DatabaseClientRequestsHandler
	LogsAndEventsProc  DBLogsAndEventsHandler
	OperationsProc     OperationsHandler
	IndicesCreator     IndicesCreatorHandler
}

type elasticProcessor struct {
	bulkRequestMaxSize int
	selfShardID        uint32
	enabledIndexes     map[string]struct{}
	elasticClient      DatabaseClientRequestsHandler
	accountsProc       DBAccountHandler
	blockProc          DBBlockHandler
	transactionsProc   DBTransactionsHandler
	miniblocksProc     DBMiniblocksHandler
	statisticsProc     DBStatisticsHandler
	validatorsProc     DBValidatorsHandler
	logsAndEventsProc  DBLogsAndEventsHandler
	operationsProc     OperationsHandler
	indicesCreator     IndicesCreatorHandler
}

// NewElasticProcessor handles Elasticsearch operations such as initialization, adding, modifying or removing data
func NewElasticProcessor(arguments *ArgElasticProcessor) (*elasticProcessor, error) {
	err := checkArguments(arguments)
	if err != nil {
		return nil, err
	}

	ei := &elasticProcessor{
		elasticClient:      arguments.DBClient,
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
		indicesCreator:     arguments.IndicesCreator,
		bulkRequestMaxSize: arguments.BulkRequestMaxSize,
	}

	return ei, nil
}

// CreateIndices will create indices, templates and policies if needed
func (ei *elasticProcessor) CreateIndices(indexTemplates, indexPolicies map[string]*bytes.Buffer, useKibana bool) error {
	return ei.indicesCreator.CreateIndicesIfNeeded(indexTemplates, indexPolicies, useKibana)
}

func (ei *elasticProcessor) getExistingObjMap(hashes []string, index string) (map[string]bool, error) {
	if len(hashes) == 0 {
		return make(map[string]bool), nil
	}

	response := make(objectsMap)
	err := ei.elasticClient.DoMultiGet(hashes, index, false, &response)
	if err != nil {
		return make(map[string]bool), err
	}

	return getDecodedResponseMultiGet(response), nil
}

func getDecodedResponseMultiGet(response objectsMap) map[string]bool {
	founded := make(map[string]bool)
	interfaceSlice, ok := response[docsKey].([]interface{})
	if !ok {
		return founded
	}

	for _, element := range interfaceSlice {
		obj := element.(objectsMap)
		_, ok = obj[errorKey]
		if ok {
			continue
		}
		founded[obj[idKey].(string)] = obj[foundKey].(bool)
	}

	return founded
}

// SaveHeader will prepare and save information about a header in elasticsearch server
func (ei *elasticProcessor) SaveHeader(
	header coreData.HeaderHandler,
	signersIndexes []uint64,
	body *block.Body,
	notarizedHeadersHashes []string,
	gasConsumptionData indexer.HeaderGasConsumption,
	txsSize int,
) error {
	if !ei.isIndexEnabled(data.BlockIndex) {
		return nil
	}

	elasticBlock, err := ei.blockProc.PrepareBlockForDB(header, signersIndexes, body, notarizedHeadersHashes, gasConsumptionData, txsSize)
	if err != nil {
		return err
	}

	buffSlice := data.NewBufferSlice(ei.bulkRequestMaxSize)
	err = ei.blockProc.SerializeBlock(elasticBlock, buffSlice, data.BlockIndex)
	if err != nil {
		return err
	}

	err = ei.indexEpochInfoData(header, buffSlice)
	if err != nil {
		return err
	}

	return ei.doBulkRequests("", buffSlice.Buffers())
}

func (ei *elasticProcessor) indexEpochInfoData(header coreData.HeaderHandler, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(data.EpochInfoIndex) ||
		ei.selfShardID != core.MetachainShardId {
		return nil
	}

	return ei.blockProc.SerializeEpochInfoData(header, buffSlice, data.EpochInfoIndex)
}

// RemoveHeader will remove a block from elasticsearch server
func (ei *elasticProcessor) RemoveHeader(header coreData.HeaderHandler) error {
	headerHash, err := ei.blockProc.ComputeHeaderHash(header)
	if err != nil {
		return err
	}

	return ei.elasticClient.DoBulkRemove(data.BlockIndex, []string{hex.EncodeToString(headerHash)})
}

// RemoveMiniblocks will remove all miniblocks that are in header from elasticsearch server
func (ei *elasticProcessor) RemoveMiniblocks(header coreData.HeaderHandler, body *block.Body) error {
	encodedMiniblocksHashes := ei.miniblocksProc.GetMiniblocksHashesHexEncoded(header, body)
	if len(encodedMiniblocksHashes) == 0 {
		return nil
	}

	return ei.elasticClient.DoBulkRemove(data.MiniblocksIndex, encodedMiniblocksHashes)
}

// RemoveTransactions will remove transaction that are in miniblock from the elasticsearch server
func (ei *elasticProcessor) RemoveTransactions(header coreData.HeaderHandler, body *block.Body) error {
	encodedTxsHashes := ei.transactionsProc.GetRewardsTxsHashesHexEncoded(header, body)
	if len(encodedTxsHashes) == 0 {
		return nil
	}

	return ei.elasticClient.DoBulkRemove(data.TransactionsIndex, encodedTxsHashes)
}

// SaveMiniblocks will prepare and save information about miniblocks in elasticsearch server
func (ei *elasticProcessor) SaveMiniblocks(header coreData.HeaderHandler, body *block.Body) error {
	if !ei.isIndexEnabled(data.MiniblocksIndex) {
		return nil
	}

	mbs := ei.miniblocksProc.PrepareDBMiniblocks(header, body)
	if len(mbs) == 0 {
		return nil
	}

	miniblocksInDBMap, err := ei.miniblocksInDBMap(mbs)
	if err != nil {
		log.Warn("elasticProcessor.SaveMiniblocks cannot get indexed miniblocks", "error", err)
	}

	buffSlice := data.NewBufferSlice(ei.bulkRequestMaxSize)
	ei.miniblocksProc.SerializeBulkMiniBlocks(mbs, miniblocksInDBMap, buffSlice, data.MiniblocksIndex)

	return ei.doBulkRequests("", buffSlice.Buffers())
}

func (ei *elasticProcessor) miniblocksInDBMap(mbs []*data.Miniblock) (map[string]bool, error) {
	mbsHashes := make([]string, len(mbs))
	for idx := range mbs {
		mbsHashes[idx] = mbs[idx].Hash
	}

	return ei.getExistingObjMap(mbsHashes, data.MiniblocksIndex)
}

// SaveTransactions will prepare and save information about a transactions in elasticsearch server
func (ei *elasticProcessor) SaveTransactions(
	body *block.Body,
	header coreData.HeaderHandler,
	pool *indexer.Pool,
) error {
	headerTimestamp := header.GetTimeStamp()

	preparedResults := ei.transactionsProc.PrepareTransactionsForDatabase(body, header, pool)
	logsData := ei.logsAndEventsProc.ExtractDataFromLogs(pool.Logs, preparedResults, headerTimestamp)

	buffers := data.NewBufferSlice(ei.bulkRequestMaxSize)
	err := ei.indexTransactions(preparedResults.Transactions, preparedResults.TxHashStatus, header, buffers)
	if err != nil {
		return err
	}

	err = ei.indexTransactionsWithRefund(preparedResults.TxHashRefund, buffers)
	if err != nil {
		return err
	}

	err = ei.indexNFTCreateInfo(logsData.Tokens, buffers)
	if err != nil {
		return err
	}

	err = ei.prepareAndIndexLogs(pool.Logs, headerTimestamp, buffers)
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
	err = ei.indexAlteredAccounts(headerTimestamp, preparedResults.AlteredAccts, logsData.NFTsDataUpdates, buffers, tagsCount)
	if err != nil {
		return err
	}

	err = ei.prepareAndIndexTagsCount(tagsCount, buffers)
	if err != nil {
		return err
	}

	err = ei.indexTokens(logsData.TokensInfo, logsData.NFTsDataUpdates, buffers)
	if err != nil {
		return err
	}

	err = ei.prepareAndIndexDelegators(logsData.Delegators, buffers)
	if err != nil {
		return err
	}

	err = ei.prepareAndIndexOperations(preparedResults.Transactions, preparedResults.TxHashStatus, header, preparedResults.ScResults, buffers)
	if err != nil {
		return err
	}

	err = ei.indexNFTBurnInfo(logsData.TokensSupply, buffers)
	if err != nil {
		return err
	}

	err = ei.prepareAndIndexRolesData(logsData.TokenRolesAndProperties, buffers)
	if err != nil {
		return err
	}

	err = ei.indexScDeploys(logsData.ScDeploys, buffers)
	if err != nil {
		return err
	}

	return ei.doBulkRequests("", buffers.Buffers())
}

func (ei *elasticProcessor) prepareAndIndexRolesData(tokenRolesAndProperties *tokeninfo.TokenRolesAndProperties, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(data.TokensIndex) {
		return nil
	}

	return ei.logsAndEventsProc.SerializeRolesData(tokenRolesAndProperties, buffSlice, data.TokensIndex)
}

func (ei *elasticProcessor) prepareAndIndexDelegators(delegators map[string]*data.Delegator, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(data.DelegatorsIndex) {
		return nil
	}

	return ei.logsAndEventsProc.SerializeDelegators(delegators, buffSlice, data.DelegatorsIndex)
}

func (ei *elasticProcessor) indexTransactionsWithRefund(txsHashRefund map[string]*data.RefundData, buffSlice *data.BufferSlice) error {
	if len(txsHashRefund) == 0 {
		return nil
	}
	txsHashes := make([]string, len(txsHashRefund))
	for txHash := range txsHashRefund {
		txsHashes = append(txsHashes, txHash)
	}

	responseTransactions := &data.ResponseTransactions{}
	err := ei.elasticClient.DoMultiGet(txsHashes, data.TransactionsIndex, true, responseTransactions)
	if err != nil {
		return err
	}

	txsFromDB := make(map[string]*data.Transaction)
	for idx := 0; idx < len(responseTransactions.Docs); idx++ {
		txRes := responseTransactions.Docs[idx]
		if !txRes.Found {
			continue
		}

		txsFromDB[txRes.ID] = &txRes.Source
	}

	return ei.transactionsProc.SerializeTransactionWithRefund(txsFromDB, txsHashRefund, buffSlice, data.TransactionsIndex)
}

func (ei *elasticProcessor) prepareAndIndexLogs(logsAndEvents []*coreData.LogData, timestamp uint64, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(data.LogsIndex) {
		return nil
	}

	logsDB := ei.logsAndEventsProc.PrepareLogsForDB(logsAndEvents, timestamp)

	return ei.logsAndEventsProc.SerializeLogs(logsDB, buffSlice, data.LogsIndex)
}

func (ei *elasticProcessor) indexScDeploys(deployData map[string]*data.ScDeployInfo, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(data.SCDeploysIndex) {
		return nil
	}

	return ei.logsAndEventsProc.SerializeSCDeploys(deployData, buffSlice, data.SCDeploysIndex)
}

func (ei *elasticProcessor) indexTransactions(txs []*data.Transaction, txHashStatus map[string]string, header coreData.HeaderHandler, bytesBuff *data.BufferSlice) error {
	if !ei.isIndexEnabled(data.TransactionsIndex) {
		return nil
	}

	return ei.transactionsProc.SerializeTransactions(txs, txHashStatus, header.GetShardID(), bytesBuff, data.TransactionsIndex)
}

func (ei *elasticProcessor) prepareAndIndexOperations(
	txs []*data.Transaction,
	txHashStatus map[string]string,
	header coreData.HeaderHandler,
	scrs []*data.ScResult,
	buffSlice *data.BufferSlice,
) error {
	if !ei.isIndexEnabled(data.OperationsIndex) {
		return nil
	}

	processedTxs, processedSCRs := ei.operationsProc.ProcessTransactionsAndSCRs(txs, scrs)

	err := ei.transactionsProc.SerializeTransactions(processedTxs, txHashStatus, header.GetShardID(), buffSlice, data.OperationsIndex)
	if err != nil {
		return err
	}

	return ei.operationsProc.SerializeSCRs(processedSCRs, buffSlice, data.OperationsIndex)
}

// SaveValidatorsRating will save validators rating
func (ei *elasticProcessor) SaveValidatorsRating(index string, validatorsRatingInfo []*data.ValidatorRatingInfo) error {
	if !ei.isIndexEnabled(data.RatingIndex) {
		return nil
	}

	buffSlice, err := ei.validatorsProc.SerializeValidatorsRating(index, validatorsRatingInfo)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(data.RatingIndex, buffSlice)
}

// SaveShardValidatorsPubKeys will prepare and save information about a shard validators public keys in elasticsearch server
func (ei *elasticProcessor) SaveShardValidatorsPubKeys(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error {
	if !ei.isIndexEnabled(data.ValidatorsIndex) {
		return nil
	}

	validatorsPubKeys := ei.validatorsProc.PrepareValidatorsPublicKeys(shardValidatorsPubKeys)
	buff, err := ei.validatorsProc.SerializeValidatorsPubKeys(validatorsPubKeys)
	if err != nil {
		return err
	}

	req := &esapi.IndexRequest{
		Index:      data.ValidatorsIndex,
		DocumentID: fmt.Sprintf("%d_%d", shardID, epoch),
		Body:       bytes.NewReader(buff.Bytes()),
	}

	return ei.elasticClient.DoRequest(req)
}

// SaveRoundsInfo will prepare and save information about a slice of rounds in elasticsearch server
func (ei *elasticProcessor) SaveRoundsInfo(info []*data.RoundInfo) error {
	if !ei.isIndexEnabled(data.RoundsIndex) {
		return nil
	}

	buff := ei.statisticsProc.SerializeRoundsInfo(info)

	return ei.elasticClient.DoBulkRequest(buff, data.RoundsIndex)
}

func (ei *elasticProcessor) indexAlteredAccounts(
	timestamp uint64,
	alteredAccounts data.AlteredAccountsHandler,
	updatesNFTsData []*data.NFTDataUpdate,
	buffSlice *data.BufferSlice,
	tagsCount data.CountTags,
) error {
	regularAccountsToIndex, accountsToIndexESDT := ei.accountsProc.GetAccounts(alteredAccounts)

	err := ei.saveAccounts(timestamp, regularAccountsToIndex, buffSlice)
	if err != nil {
		return err
	}

	return ei.saveAccountsESDT(timestamp, accountsToIndexESDT, updatesNFTsData, buffSlice, tagsCount)
}

func (ei *elasticProcessor) saveAccountsESDT(
	timestamp uint64,
	wrappedAccounts []*data.AccountESDT,
	updatesNFTsData []*data.NFTDataUpdate,
	buffSlice *data.BufferSlice,
	tagsCount data.CountTags,
) error {
	accountsESDTMap, tokensData := ei.accountsProc.PrepareAccountsMapESDT(timestamp, wrappedAccounts, tagsCount)
	err := ei.addTokenTypeAndCurrentOwnerInAccountsESDT(tokensData, accountsESDTMap)
	if err != nil {
		return err
	}

	err = ei.indexAccountsESDT(accountsESDTMap, updatesNFTsData, buffSlice)
	if err != nil {
		return err
	}

	return ei.saveAccountsESDTHistory(timestamp, accountsESDTMap, buffSlice)
}

func (ei *elasticProcessor) addTokenTypeAndCurrentOwnerInAccountsESDT(tokensData data.TokensHandler, accountsESDTMap map[string]*data.AccountInfo) error {
	if check.IfNil(tokensData) || tokensData.Len() == 0 {
		return nil
	}

	responseTokens := &data.ResponseTokens{}
	err := ei.elasticClient.DoMultiGet(tokensData.GetAllTokens(), data.TokensIndex, true, responseTokens)
	if err != nil {
		return err
	}

	tokensData.AddTypeAndOwnerFromResponse(responseTokens)
	tokensData.PutTypeAndOwnerInAccountsESDT(accountsESDTMap)

	return nil
}

func (ei *elasticProcessor) prepareAndIndexTagsCount(tagsCount data.CountTags, buffSlice *data.BufferSlice) error {
	shouldSkipIndex := !ei.isIndexEnabled(data.TagsIndex) || tagsCount.Len() == 0
	if shouldSkipIndex {
		return nil
	}

	return tagsCount.Serialize(buffSlice, data.TagsIndex)
}

func (ei *elasticProcessor) indexAccountsESDT(
	accountsESDTMap map[string]*data.AccountInfo,
	updatesNFTsData []*data.NFTDataUpdate,
	buffSlice *data.BufferSlice,
) error {
	if !ei.isIndexEnabled(data.AccountsESDTIndex) {
		return nil
	}

	return ei.accountsProc.SerializeAccountsESDT(accountsESDTMap, updatesNFTsData, buffSlice, data.AccountsESDTIndex)
}

func (ei *elasticProcessor) indexNFTCreateInfo(tokensData data.TokensHandler, buffSlice *data.BufferSlice) error {
	shouldSkipIndex := !ei.isIndexEnabled(data.TokensIndex) || tokensData.Len() == 0
	if shouldSkipIndex {
		return nil
	}

	responseTokens := &data.ResponseTokens{}
	err := ei.elasticClient.DoMultiGet(tokensData.GetAllTokens(), data.TokensIndex, true, responseTokens)
	if err != nil {
		return err
	}

	tokensData.AddTypeAndOwnerFromResponse(responseTokens)

	tokens := tokensData.GetAllWithoutMetaESDT()
	ei.accountsProc.PutTokenMedataDataInTokens(tokens)

	return ei.accountsProc.SerializeNFTCreateInfo(tokens, buffSlice, data.TokensIndex)
}

func (ei *elasticProcessor) indexNFTBurnInfo(tokensData data.TokensHandler, buffSlice *data.BufferSlice) error {
	shouldSkipIndex := !ei.isIndexEnabled(data.TokensIndex) || tokensData.Len() == 0
	if shouldSkipIndex {
		return nil
	}

	responseTokens := &data.ResponseTokens{}
	err := ei.elasticClient.DoMultiGet(tokensData.GetAllTokens(), data.TokensIndex, true, responseTokens)
	if err != nil {
		return err
	}

	// TODO implement to keep in tokens also the supply
	tokensData.AddTypeAndOwnerFromResponse(responseTokens)
	return ei.logsAndEventsProc.SerializeSupplyData(tokensData, buffSlice, data.TokensIndex)
}

// SaveAccounts will prepare and save information about provided accounts in elasticsearch server
func (ei *elasticProcessor) SaveAccounts(timestamp uint64, accts []*data.Account) error {
	buffSlice := data.NewBufferSlice(ei.bulkRequestMaxSize)
	return ei.saveAccounts(timestamp, accts, buffSlice)
}

func (ei *elasticProcessor) saveAccounts(timestamp uint64, accts []*data.Account, buffSlice *data.BufferSlice) error {
	accountsMap := ei.accountsProc.PrepareRegularAccountsMap(timestamp, accts)
	err := ei.indexAccounts(accountsMap, data.AccountsIndex, buffSlice)
	if err != nil {
		return err
	}

	return ei.saveAccountsHistory(timestamp, accountsMap, buffSlice)
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

func (ei *elasticProcessor) saveAccountsESDTHistory(timestamp uint64, accountsInfoMap map[string]*data.AccountInfo, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(data.AccountsESDTHistoryIndex) {
		return nil
	}

	accountsMap := ei.accountsProc.PrepareAccountsHistory(timestamp, accountsInfoMap)

	return ei.serializeAndIndexAccountsHistory(accountsMap, data.AccountsESDTHistoryIndex, buffSlice)
}

func (ei *elasticProcessor) saveAccountsHistory(timestamp uint64, accountsInfoMap map[string]*data.AccountInfo, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(data.AccountsHistoryIndex) {
		return nil
	}

	accountsMap := ei.accountsProc.PrepareAccountsHistory(timestamp, accountsInfoMap)

	return ei.serializeAndIndexAccountsHistory(accountsMap, data.AccountsHistoryIndex, buffSlice)
}

func (ei *elasticProcessor) serializeAndIndexAccountsHistory(accountsMap map[string]*data.AccountBalanceHistory, index string, buffSlice *data.BufferSlice) error {
	return ei.accountsProc.SerializeAccountsHistory(accountsMap, buffSlice, index)
}

func (ei *elasticProcessor) indexScResults(scrs []*data.ScResult, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(data.ScResultsIndex) {
		return nil
	}

	return ei.transactionsProc.SerializeScResults(scrs, buffSlice, data.ScResultsIndex)
}

func (ei *elasticProcessor) indexReceipts(receipts []*data.Receipt, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(data.ReceiptsIndex) {
		return nil
	}

	return ei.transactionsProc.SerializeReceipts(receipts, buffSlice, data.ReceiptsIndex)
}

func (ei *elasticProcessor) isIndexEnabled(index string) bool {
	_, isEnabled := ei.enabledIndexes[index]
	return isEnabled
}

func (ei *elasticProcessor) doBulkRequests(index string, buffSlice []*bytes.Buffer) error {
	var err error
	for idx := range buffSlice {
		err = ei.elasticClient.DoBulkRequest(buffSlice[idx], index)
		if err != nil {
			return fmt.Errorf("index: %s, message: %s", index, err.Error())
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ei *elasticProcessor) IsInterfaceNil() bool {
	return ei == nil
}
