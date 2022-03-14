package process

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
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
	SelfShardID       uint32
	EnabledIndexes    map[string]struct{}
	TransactionsProc  DBTransactionsHandler
	AccountsProc      DBAccountHandler
	BlockProc         DBBlockHandler
	MiniblocksProc    DBMiniblocksHandler
	StatisticsProc    DBStatisticsHandler
	ValidatorsProc    DBValidatorsHandler
	DBClient          DatabaseClientRequestsHandler
	LogsAndEventsProc DBLogsAndEventsHandler
	OperationsProc    OperationsHandler
	IndicesCreator    IndicesCreatorHandler
}

type elasticProcessor struct {
	selfShardID       uint32
	enabledIndexes    map[string]struct{}
	elasticClient     DatabaseClientRequestsHandler
	accountsProc      DBAccountHandler
	blockProc         DBBlockHandler
	transactionsProc  DBTransactionsHandler
	miniblocksProc    DBMiniblocksHandler
	statisticsProc    DBStatisticsHandler
	validatorsProc    DBValidatorsHandler
	logsAndEventsProc DBLogsAndEventsHandler
	operationsProc    OperationsHandler
	indicesCreator    IndicesCreatorHandler
}

// NewElasticProcessor handles Elasticsearch operations such as initialization, adding, modifying or removing data
func NewElasticProcessor(arguments *ArgElasticProcessor) (*elasticProcessor, error) {
	err := checkArguments(arguments)
	if err != nil {
		return nil, err
	}

	ei := &elasticProcessor{
		elasticClient:     arguments.DBClient,
		enabledIndexes:    arguments.EnabledIndexes,
		accountsProc:      arguments.AccountsProc,
		blockProc:         arguments.BlockProc,
		miniblocksProc:    arguments.MiniblocksProc,
		transactionsProc:  arguments.TransactionsProc,
		selfShardID:       arguments.SelfShardID,
		statisticsProc:    arguments.StatisticsProc,
		validatorsProc:    arguments.ValidatorsProc,
		logsAndEventsProc: arguments.LogsAndEventsProc,
		operationsProc:    arguments.OperationsProc,
		indicesCreator:    arguments.IndicesCreator,
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

	buff, err := ei.blockProc.SerializeBlock(elasticBlock)
	if err != nil {
		return err
	}

	req := &esapi.IndexRequest{
		Index:      data.BlockIndex,
		DocumentID: elasticBlock.Hash,
		Body:       bytes.NewReader(buff.Bytes()),
	}

	err = ei.elasticClient.DoRequest(req)
	if err != nil {
		return err
	}

	return ei.indexEpochInfoData(header)
}

func (ei *elasticProcessor) indexEpochInfoData(header coreData.HeaderHandler) error {
	if !ei.isIndexEnabled(data.EpochInfoIndex) ||
		ei.selfShardID != core.MetachainShardId {
		return nil
	}

	buff, err := ei.blockProc.SerializeEpochInfoData(header)
	if err != nil {
		return err
	}

	req := &esapi.IndexRequest{
		Index:      data.EpochInfoIndex,
		DocumentID: fmt.Sprintf("%d", header.GetEpoch()),
		Body:       bytes.NewReader(buff.Bytes()),
	}

	return ei.elasticClient.DoRequest(req)
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

	buff := ei.miniblocksProc.SerializeBulkMiniBlocks(mbs, miniblocksInDBMap)
	return ei.elasticClient.DoBulkRequest(buff, data.MiniblocksIndex)
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

	err := ei.indexTransactions(preparedResults.Transactions, preparedResults.TxHashStatus, header)
	if err != nil {
		return err
	}

	err = ei.indexTransactionsWithRefund(preparedResults.TxHashRefund)
	if err != nil {
		return err
	}

	err = ei.prepareAndIndexTagsCount(logsData.TagsCount)
	if err != nil {
		return err
	}

	err = ei.indexNFTCreateInfo(logsData.Tokens)
	if err != nil {
		return err
	}

	err = ei.prepareAndIndexLogs(pool.Logs, headerTimestamp)
	if err != nil {
		return err
	}

	err = ei.indexScResults(preparedResults.ScResults)
	if err != nil {
		return err
	}

	err = ei.indexReceipts(preparedResults.Receipts)
	if err != nil {
		return err
	}

	err = ei.indexAlteredAccounts(headerTimestamp, preparedResults.AlteredAccts, logsData.NFTsDataUpdates)
	if err != nil {
		return err
	}

	err = ei.indexTokens(logsData.TokensInfo, logsData.NFTsDataUpdates)
	if err != nil {
		return err
	}

	err = ei.prepareAndIndexDelegators(logsData.Delegators)
	if err != nil {
		return err
	}

	err = ei.prepareAndIndexOperations(preparedResults.Transactions, preparedResults.TxHashStatus, header, preparedResults.ScResults)
	if err != nil {
		return err
	}

	err = ei.indexNFTBurnInfo(logsData.TokensSupply)
	if err != nil {
		return err
	}

	err = ei.prepareAndIndexRolesData(logsData.RolesData)
	if err != nil {
		return err
	}

	return ei.indexScDeploys(logsData.ScDeploys)
}

func (ei *elasticProcessor) prepareAndIndexRolesData(rolesData data.RolesData) error {
	buffSlice, err := ei.logsAndEventsProc.SerializeRolesData(rolesData)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(data.TokensIndex, buffSlice)
}

func (ei *elasticProcessor) prepareAndIndexDelegators(delegators map[string]*data.Delegator) error {
	if !ei.isIndexEnabled(data.DelegatorsIndex) {
		return nil
	}

	buffSlice, err := ei.logsAndEventsProc.SerializeDelegators(delegators)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(data.DelegatorsIndex, buffSlice)
}

func (ei *elasticProcessor) indexTransactionsWithRefund(txsHashRefund map[string]*data.RefundData) error {
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

	buffSlice, err := ei.transactionsProc.SerializeTransactionWithRefund(txsFromDB, txsHashRefund)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(data.TransactionsIndex, buffSlice)
}

func (ei *elasticProcessor) prepareAndIndexLogs(logsAndEvents []*coreData.LogData, timestamp uint64) error {
	if !ei.isIndexEnabled(data.LogsIndex) {
		return nil
	}

	logsDB := ei.logsAndEventsProc.PrepareLogsForDB(logsAndEvents, timestamp)
	buffSlice, err := ei.logsAndEventsProc.SerializeLogs(logsDB)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(data.LogsIndex, buffSlice)
}

func (ei *elasticProcessor) indexScDeploys(deployData map[string]*data.ScDeployInfo) error {
	if !ei.isIndexEnabled(data.SCDeploysIndex) {
		return nil
	}

	buffSlice, err := ei.logsAndEventsProc.SerializeSCDeploys(deployData)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(data.SCDeploysIndex, buffSlice)
}

func (ei *elasticProcessor) indexTransactions(txs []*data.Transaction, txHashStatus map[string]string, header coreData.HeaderHandler) error {
	if !ei.isIndexEnabled(data.TransactionsIndex) {
		return nil
	}

	buffSlice, err := ei.transactionsProc.SerializeTransactions(txs, txHashStatus, header.GetShardID())
	if err != nil {
		return err
	}

	return ei.doBulkRequests(data.TransactionsIndex, buffSlice)
}

func (ei *elasticProcessor) prepareAndIndexOperations(txs []*data.Transaction, txHashStatus map[string]string, header coreData.HeaderHandler, scrs []*data.ScResult) error {
	if !ei.isIndexEnabled(data.OperationsIndex) {
		return nil
	}

	processedTxs, processedSCRs := ei.operationsProc.ProcessTransactionsAndSCRs(txs, scrs)

	buffSlice, err := ei.transactionsProc.SerializeTransactions(processedTxs, txHashStatus, header.GetShardID())
	if err != nil {
		return err
	}

	err = ei.doBulkRequests(data.OperationsIndex, buffSlice)
	if err != nil {
		return err
	}

	buffSliceSCRs, err := ei.operationsProc.SerializeSCRs(processedSCRs)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(data.OperationsIndex, buffSliceSCRs)
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
) error {
	regularAccountsToIndex, accountsToIndexESDT := ei.accountsProc.GetAccounts(alteredAccounts)

	err := ei.SaveAccounts(timestamp, regularAccountsToIndex)
	if err != nil {
		return err
	}

	return ei.saveAccountsESDT(timestamp, accountsToIndexESDT, updatesNFTsData)
}

func (ei *elasticProcessor) saveAccountsESDT(
	timestamp uint64,
	wrappedAccounts []*data.AccountESDT,
	updatesNFTsData []*data.NFTDataUpdate,
) error {
	accountsESDTMap, tokensData := ei.accountsProc.PrepareAccountsMapESDT(timestamp, wrappedAccounts)
	err := ei.addTokenTypeInAccountsESDT(tokensData, accountsESDTMap)
	if err != nil {
		return err
	}

	err = ei.indexAccountsESDT(accountsESDTMap, updatesNFTsData)
	if err != nil {
		return err
	}

	return ei.saveAccountsESDTHistory(timestamp, accountsESDTMap)
}

func (ei *elasticProcessor) addTokenTypeInAccountsESDT(tokensData data.TokensHandler, accountsESDTMap map[string]*data.AccountInfo) error {
	if check.IfNil(tokensData) || tokensData.Len() == 0 {
		return nil
	}

	responseTokens := &data.ResponseTokens{}
	err := ei.elasticClient.DoMultiGet(tokensData.GetAllTokens(), data.TokensIndex, true, responseTokens)
	if err != nil {
		return err
	}

	tokensData.AddTypeFromResponse(responseTokens)
	tokensData.PutTypeInAccountsESDT(accountsESDTMap)

	return nil
}

func (ei *elasticProcessor) prepareAndIndexTagsCount(tagsCount data.CountTags) error {
	shouldSkipIndex := !ei.isIndexEnabled(data.TagsIndex) || tagsCount.Len() == 0
	if shouldSkipIndex {
		return nil
	}

	serializedTags, err := tagsCount.Serialize()
	if err != nil {
		return err
	}

	return ei.doBulkRequests(data.TagsIndex, serializedTags)
}

func (ei *elasticProcessor) indexAccountsESDT(
	accountsESDTMap map[string]*data.AccountInfo,
	updatesNFTsData []*data.NFTDataUpdate,
) error {
	if !ei.isIndexEnabled(data.AccountsESDTIndex) {
		return nil
	}

	buffSlice, err := ei.accountsProc.SerializeAccountsESDT(accountsESDTMap, updatesNFTsData)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(data.AccountsESDTIndex, buffSlice)
}

func (ei *elasticProcessor) indexNFTCreateInfo(tokensData data.TokensHandler) error {
	shouldSkipIndex := !ei.isIndexEnabled(data.TokensIndex) || tokensData.Len() == 0
	if shouldSkipIndex {
		return nil
	}

	responseTokens := &data.ResponseTokens{}
	err := ei.elasticClient.DoMultiGet(tokensData.GetAllTokens(), data.TokensIndex, true, responseTokens)
	if err != nil {
		return err
	}

	tokensData.AddTypeFromResponse(responseTokens)

	tokens := tokensData.GetAll()
	ei.accountsProc.PutTokenMedataDataInTokens(tokens)

	buffSlice, err := ei.accountsProc.SerializeNFTCreateInfo(tokens)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(data.TokensIndex, buffSlice)
}

func (ei *elasticProcessor) indexNFTBurnInfo(tokensData data.TokensHandler) error {
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
	tokensData.AddTypeFromResponse(responseTokens)
	buffSlice, err := ei.logsAndEventsProc.SerializeSupplyData(tokensData)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(data.TokensIndex, buffSlice)
}

// SaveAccounts will prepare and save information about provided accounts in elasticsearch server
func (ei *elasticProcessor) SaveAccounts(timestamp uint64, accts []*data.Account) error {
	accountsMap := ei.accountsProc.PrepareRegularAccountsMap(accts)
	err := ei.indexAccounts(accountsMap, data.AccountsIndex)
	if err != nil {
		return err
	}

	return ei.saveAccountsHistory(timestamp, accountsMap)
}

func (ei *elasticProcessor) indexAccounts(accountsMap map[string]*data.AccountInfo, index string) error {
	if !ei.isIndexEnabled(index) {
		return nil
	}

	return ei.serializeAndIndexAccounts(accountsMap, index)
}

func (ei *elasticProcessor) serializeAndIndexAccounts(accountsMap map[string]*data.AccountInfo, index string) error {
	buffSlice, err := ei.accountsProc.SerializeAccounts(accountsMap)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(index, buffSlice)
}

func (ei *elasticProcessor) saveAccountsESDTHistory(timestamp uint64, accountsInfoMap map[string]*data.AccountInfo) error {
	if !ei.isIndexEnabled(data.AccountsESDTHistoryIndex) {
		return nil
	}

	accountsMap := ei.accountsProc.PrepareAccountsHistory(timestamp, accountsInfoMap)

	return ei.serializeAndIndexAccountsHistory(accountsMap, data.AccountsESDTHistoryIndex)
}

func (ei *elasticProcessor) saveAccountsHistory(timestamp uint64, accountsInfoMap map[string]*data.AccountInfo) error {
	if !ei.isIndexEnabled(data.AccountsHistoryIndex) {
		return nil
	}

	accountsMap := ei.accountsProc.PrepareAccountsHistory(timestamp, accountsInfoMap)

	return ei.serializeAndIndexAccountsHistory(accountsMap, data.AccountsHistoryIndex)
}

func (ei *elasticProcessor) serializeAndIndexAccountsHistory(accountsMap map[string]*data.AccountBalanceHistory, index string) error {
	buffSlice, err := ei.accountsProc.SerializeAccountsHistory(accountsMap)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(index, buffSlice)
}

func (ei *elasticProcessor) indexScResults(scrs []*data.ScResult) error {
	if !ei.isIndexEnabled(data.ScResultsIndex) {
		return nil
	}

	buffSlice, err := ei.transactionsProc.SerializeScResults(scrs)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(data.ScResultsIndex, buffSlice)
}

func (ei *elasticProcessor) indexReceipts(receipts []*data.Receipt) error {
	if !ei.isIndexEnabled(data.ReceiptsIndex) {
		return nil
	}

	buffSlice, err := ei.transactionsProc.SerializeReceipts(receipts)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(data.ReceiptsIndex, buffSlice)
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
