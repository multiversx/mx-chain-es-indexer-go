package process

import (
	"bytes"
	"encoding/hex"
	"fmt"

	elasticIndexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
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

	indexes = []string{
		elasticIndexer.TransactionsIndex, elasticIndexer.BlockIndex, elasticIndexer.MiniblocksIndex, elasticIndexer.RatingIndex, elasticIndexer.RoundsIndex, elasticIndexer.ValidatorsIndex,
		elasticIndexer.AccountsIndex, elasticIndexer.AccountsHistoryIndex, elasticIndexer.ReceiptsIndex, elasticIndexer.ScResultsIndex, elasticIndexer.AccountsESDTHistoryIndex, elasticIndexer.AccountsESDTIndex,
		elasticIndexer.EpochInfoIndex, elasticIndexer.SCDeploysIndex, elasticIndexer.TokensIndex, elasticIndexer.TagsIndex, elasticIndexer.LogsIndex,
	}
)

type objectsMap = map[string]interface{}

// ArgElasticProcessor holds all dependencies required by the elasticProcessor in order to create
// new instances
type ArgElasticProcessor struct {
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
	DBClient          DatabaseClientHandler
	LogsAndEventsProc DBLogsAndEventsHandler
}

type elasticProcessor struct {
	selfShardID       uint32
	enabledIndexes    map[string]struct{}
	elasticClient     DatabaseClientHandler
	accountsProc      DBAccountHandler
	blockProc         DBBlockHandler
	transactionsProc  DBTransactionsHandler
	miniblocksProc    DBMiniblocksHandler
	statisticsProc    DBStatisticsHandler
	validatorsProc    DBValidatorsHandler
	logsAndEventsProc DBLogsAndEventsHandler
}

// NewElasticProcessor handles Elastic Search operations such as initialization, adding, modifying or removing data
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
	}

	err = ei.init(arguments.UseKibana, arguments.IndexTemplates, arguments.IndexPolicies)
	if err != nil {
		return nil, err
	}

	return ei, nil
}

func checkArguments(arguments *ArgElasticProcessor) error {
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

func (ei *elasticProcessor) init(useKibana bool, indexTemplates, _ map[string]*bytes.Buffer) error {
	err := ei.createOpenDistroTemplates(indexTemplates)
	if err != nil {
		return err
	}

	if useKibana {
		// TODO: Re-activate after we think of a solid way to handle forks+rotating indexes
		// err = ei.createIndexPolicies(indexPolicies)
		// if err != nil {
		//	return err
		// }
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

	return nil
}

// nolint
func (ei *elasticProcessor) createIndexPolicies(indexPolicies map[string]*bytes.Buffer) error {
	indexesPolicies := []string{elasticIndexer.TransactionsPolicy, elasticIndexer.BlockPolicy, elasticIndexer.MiniblocksPolicy, elasticIndexer.RatingPolicy, elasticIndexer.RoundsPolicy, elasticIndexer.ValidatorsPolicy,
		elasticIndexer.AccountsPolicy, elasticIndexer.AccountsESDTPolicy, elasticIndexer.AccountsHistoryPolicy, elasticIndexer.AccountsESDTHistoryPolicy, elasticIndexer.AccountsESDTIndex, elasticIndexer.ReceiptsPolicy, elasticIndexer.ScResultsPolicy}
	for _, indexPolicyName := range indexesPolicies {
		indexPolicy := getTemplateByName(indexPolicyName, indexPolicies)
		if indexPolicy != nil {
			err := ei.elasticClient.CheckAndCreatePolicy(indexPolicyName, indexPolicy)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (ei *elasticProcessor) createOpenDistroTemplates(indexTemplates map[string]*bytes.Buffer) error {
	opendistroTemplate := getTemplateByName(elasticIndexer.OpenDistroIndex, indexTemplates)
	if opendistroTemplate != nil {
		err := ei.elasticClient.CheckAndCreateTemplate(elasticIndexer.OpenDistroIndex, opendistroTemplate)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ei *elasticProcessor) createIndexTemplates(indexTemplates map[string]*bytes.Buffer) error {
	for _, index := range indexes {
		indexTemplate := getTemplateByName(index, indexTemplates)
		if indexTemplate != nil {
			err := ei.elasticClient.CheckAndCreateTemplate(index, indexTemplate)
			if err != nil {
				return err
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
			return err
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

func getTemplateByName(templateName string, templateList map[string]*bytes.Buffer) *bytes.Buffer {
	if template, ok := templateList[templateName]; ok {
		return template
	}

	log.Debug("elasticProcessor.getTemplateByName", "could not find template", templateName)
	return nil
}

// SaveHeader will prepare and save information about a header in elasticsearch server
func (ei *elasticProcessor) SaveHeader(
	header coreData.HeaderHandler,
	signersIndexes []uint64,
	body *block.Body,
	notarizedHeadersHashes []string,
	txsSize int,
) error {
	if !ei.isIndexEnabled(elasticIndexer.BlockIndex) {
		return nil
	}

	elasticBlock, err := ei.blockProc.PrepareBlockForDB(header, signersIndexes, body, notarizedHeadersHashes, txsSize)
	if err != nil {
		return err
	}

	buff, err := ei.blockProc.SerializeBlock(elasticBlock)
	if err != nil {
		return err
	}

	req := &esapi.IndexRequest{
		Index:      elasticIndexer.BlockIndex,
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
	if !ei.isIndexEnabled(elasticIndexer.EpochInfoIndex) ||
		ei.selfShardID != core.MetachainShardId {
		return nil
	}

	buff, err := ei.blockProc.SerializeEpochInfoData(header)
	if err != nil {
		return err
	}

	req := &esapi.IndexRequest{
		Index:      elasticIndexer.EpochInfoIndex,
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

	return ei.elasticClient.DoBulkRemove(elasticIndexer.BlockIndex, []string{hex.EncodeToString(headerHash)})
}

// RemoveMiniblocks will remove all miniblocks that are in header from elasticsearch server
func (ei *elasticProcessor) RemoveMiniblocks(header coreData.HeaderHandler, body *block.Body) error {
	encodedMiniblocksHashes := ei.miniblocksProc.GetMiniblocksHashesHexEncoded(header, body)
	if len(encodedMiniblocksHashes) == 0 {
		return nil
	}

	return ei.elasticClient.DoBulkRemove(elasticIndexer.MiniblocksIndex, encodedMiniblocksHashes)
}

// RemoveTransactions will remove transaction that are in miniblock from the elasticsearch server
func (ei *elasticProcessor) RemoveTransactions(header coreData.HeaderHandler, body *block.Body) error {
	encodedTxsHashes := ei.transactionsProc.GetRewardsTxsHashesHexEncoded(header, body)
	if len(encodedTxsHashes) == 0 {
		return nil
	}

	return ei.elasticClient.DoBulkRemove(elasticIndexer.TransactionsIndex, encodedTxsHashes)
}

// SaveMiniblocks will prepare and save information about miniblocks in elasticsearch server
func (ei *elasticProcessor) SaveMiniblocks(header coreData.HeaderHandler, body *block.Body) error {
	if !ei.isIndexEnabled(elasticIndexer.MiniblocksIndex) {
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
	return ei.elasticClient.DoBulkRequest(buff, elasticIndexer.MiniblocksIndex)
}

func (ei *elasticProcessor) miniblocksInDBMap(mbs []*data.Miniblock) (map[string]bool, error) {
	mbsHashes := make([]string, len(mbs))
	for idx := range mbs {
		mbsHashes[idx] = mbs[idx].Hash
	}

	return ei.getExistingObjMap(mbsHashes, elasticIndexer.MiniblocksIndex)
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

	err = ei.indexAlteredAccounts(headerTimestamp, preparedResults.AlteredAccts, logsData.PendingBalances)
	if err != nil {
		return err
	}

	err = ei.indexTokens(preparedResults.Tokens)
	if err != nil {
		return err
	}

	return ei.indexScDeploys(logsData.ScDeploys)
}

func (ei *elasticProcessor) prepareAndIndexLogs(logsAndEvents map[string]coreData.LogHandler, timestamp uint64) error {
	if !ei.isIndexEnabled(elasticIndexer.LogsIndex) {
		return nil
	}

	logsDB := ei.logsAndEventsProc.PrepareLogsForDB(logsAndEvents, timestamp)
	buffSlice, err := ei.logsAndEventsProc.SerializeLogs(logsDB)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(elasticIndexer.LogsIndex, buffSlice)
}

func (ei *elasticProcessor) indexTokens(tokensData []*data.TokenInfo) error {
	if !ei.isIndexEnabled(elasticIndexer.TokensIndex) {
		return nil
	}

	buffSlice, err := ei.transactionsProc.SerializeTokens(tokensData)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(elasticIndexer.TokensIndex, buffSlice)
}

func (ei *elasticProcessor) indexScDeploys(deployData map[string]*data.ScDeployInfo) error {
	if !ei.isIndexEnabled(elasticIndexer.SCDeploysIndex) {
		return nil
	}

	buffSlice, err := ei.logsAndEventsProc.SerializeSCDeploys(deployData)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(elasticIndexer.SCDeploysIndex, buffSlice)
}

func (ei *elasticProcessor) indexTransactions(txs []*data.Transaction, txHashStatus map[string]string, header coreData.HeaderHandler) error {
	if !ei.isIndexEnabled(elasticIndexer.TransactionsIndex) {
		return nil
	}

	buffSlice, err := ei.transactionsProc.SerializeTransactions(txs, txHashStatus, header.GetShardID())
	if err != nil {
		return err
	}

	return ei.doBulkRequests(elasticIndexer.TransactionsIndex, buffSlice)
}

// SaveValidatorsRating will save validators rating
func (ei *elasticProcessor) SaveValidatorsRating(index string, validatorsRatingInfo []*data.ValidatorRatingInfo) error {
	if !ei.isIndexEnabled(elasticIndexer.RatingIndex) {
		return nil
	}

	buffSlice, err := ei.validatorsProc.SerializeValidatorsRating(index, validatorsRatingInfo)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(elasticIndexer.RatingIndex, buffSlice)
}

// SaveShardValidatorsPubKeys will prepare and save information about a shard validators public keys in elasticsearch server
func (ei *elasticProcessor) SaveShardValidatorsPubKeys(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error {
	if !ei.isIndexEnabled(elasticIndexer.ValidatorsIndex) {
		return nil
	}

	validatorsPubKeys := ei.validatorsProc.PrepareValidatorsPublicKeys(shardValidatorsPubKeys)
	buff, err := ei.validatorsProc.SerializeValidatorsPubKeys(validatorsPubKeys)
	if err != nil {
		return err
	}

	req := &esapi.IndexRequest{
		Index:      elasticIndexer.ValidatorsIndex,
		DocumentID: fmt.Sprintf("%d_%d", shardID, epoch),
		Body:       bytes.NewReader(buff.Bytes()),
	}

	return ei.elasticClient.DoRequest(req)
}

// SaveRoundsInfo will prepare and save information about a slice of rounds in elasticsearch server
func (ei *elasticProcessor) SaveRoundsInfo(info []*data.RoundInfo) error {
	if !ei.isIndexEnabled(elasticIndexer.RoundsIndex) {
		return nil
	}

	buff := ei.statisticsProc.SerializeRoundsInfo(info)

	return ei.elasticClient.DoBulkRequest(buff, elasticIndexer.RoundsIndex)
}

func (ei *elasticProcessor) indexAlteredAccounts(
	timestamp uint64,
	alteredAccounts data.AlteredAccountsHandler,
	pendingBalances map[string]*data.AccountInfo,
) error {
	regularAccountsToIndex, accountsToIndexESDT := ei.accountsProc.GetAccounts(alteredAccounts)

	err := ei.SaveAccounts(timestamp, regularAccountsToIndex)
	if err != nil {
		return err
	}

	return ei.saveAccountsESDT(timestamp, accountsToIndexESDT, pendingBalances)
}

func (ei *elasticProcessor) saveAccountsESDT(
	timestamp uint64,
	wrappedAccounts []*data.AccountESDT,
	pendingBalances map[string]*data.AccountInfo,
) error {
	accountsESDTMap := ei.accountsProc.PrepareAccountsMapESDT(wrappedAccounts)

	resAccountsMap := converters.MergeAccountsInfoMaps(accountsESDTMap, pendingBalances)
	err := ei.indexAccountsESDT(resAccountsMap)
	if err != nil {
		return err
	}

	return ei.saveAccountsESDTHistory(timestamp, accountsESDTMap)
}

func (ei *elasticProcessor) prepareAndIndexTagsCount(tagsCount data.CountTags) error {
	shouldSkipIndex := !ei.isIndexEnabled(elasticIndexer.TagsIndex) || tagsCount.Len() == 0
	if shouldSkipIndex {
		return nil
	}

	serializedTags, err := tagsCount.Serialize()
	if err != nil {
		return err
	}

	return ei.doBulkRequests(elasticIndexer.TagsIndex, serializedTags)
}

func (ei *elasticProcessor) indexAccountsESDT(accountsESDTMap map[string]*data.AccountInfo) error {
	if !ei.isIndexEnabled(elasticIndexer.AccountsESDTIndex) {
		return nil
	}

	return ei.serializeAndIndexAccounts(accountsESDTMap, elasticIndexer.AccountsESDTIndex, true)
}

func (ei *elasticProcessor) indexNFTCreateInfo(tokensData data.TokensHandler) error {
	shouldSkipIndex := !ei.isIndexEnabled(elasticIndexer.TokensIndex) || tokensData.Len() == 0
	if shouldSkipIndex {
		return nil
	}

	responseTokens := &data.ResponseTokens{}
	err := ei.elasticClient.DoMultiGet(tokensData.GetAllTokens(), elasticIndexer.TokensIndex, true, responseTokens)
	if err != nil {
		return err
	}

	tokensData.AddTypeFromResponse(responseTokens)

	buffSlice, err := ei.accountsProc.SerializeNFTCreateInfo(tokensData.GetAll())
	if err != nil {
		return err
	}

	return ei.doBulkRequests(elasticIndexer.TokensIndex, buffSlice)
}

// SaveAccounts will prepare and save information about provided accounts in elasticsearch server
func (ei *elasticProcessor) SaveAccounts(timestamp uint64, accts []*data.Account) error {
	accountsMap := ei.accountsProc.PrepareRegularAccountsMap(accts)
	err := ei.indexAccounts(accountsMap, elasticIndexer.AccountsIndex, false)
	if err != nil {
		return err
	}

	return ei.saveAccountsHistory(timestamp, accountsMap)
}

func (ei *elasticProcessor) indexAccounts(accountsMap map[string]*data.AccountInfo, index string, areESDTAccounts bool) error {
	if !ei.isIndexEnabled(elasticIndexer.AccountsIndex) {
		return nil
	}

	return ei.serializeAndIndexAccounts(accountsMap, index, areESDTAccounts)
}

func (ei *elasticProcessor) serializeAndIndexAccounts(accountsMap map[string]*data.AccountInfo, index string, areESDTAccounts bool) error {
	buffSlice, err := ei.accountsProc.SerializeAccounts(accountsMap, areESDTAccounts)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(index, buffSlice)
}

func (ei *elasticProcessor) saveAccountsESDTHistory(timestamp uint64, accountsInfoMap map[string]*data.AccountInfo) error {
	if !ei.isIndexEnabled(elasticIndexer.AccountsESDTHistoryIndex) {
		return nil
	}

	accountsMap := ei.accountsProc.PrepareAccountsHistory(timestamp, accountsInfoMap)

	return ei.serializeAndIndexAccountsHistory(accountsMap, elasticIndexer.AccountsESDTHistoryIndex)
}

func (ei *elasticProcessor) saveAccountsHistory(timestamp uint64, accountsInfoMap map[string]*data.AccountInfo) error {
	if !ei.isIndexEnabled(elasticIndexer.AccountsHistoryIndex) {
		return nil
	}

	accountsMap := ei.accountsProc.PrepareAccountsHistory(timestamp, accountsInfoMap)

	return ei.serializeAndIndexAccountsHistory(accountsMap, elasticIndexer.AccountsHistoryIndex)
}

func (ei *elasticProcessor) serializeAndIndexAccountsHistory(accountsMap map[string]*data.AccountBalanceHistory, index string) error {
	buffSlice, err := ei.accountsProc.SerializeAccountsHistory(accountsMap)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(index, buffSlice)
}

func (ei *elasticProcessor) indexScResults(scrs []*data.ScResult) error {
	if !ei.isIndexEnabled(elasticIndexer.ScResultsIndex) {
		return nil
	}

	buffSlice, err := ei.transactionsProc.SerializeScResults(scrs)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(elasticIndexer.ScResultsIndex, buffSlice)
}

func (ei *elasticProcessor) indexReceipts(receipts []*data.Receipt) error {
	if !ei.isIndexEnabled(elasticIndexer.ReceiptsIndex) {
		return nil
	}

	buffSlice, err := ei.transactionsProc.SerializeReceipts(receipts)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(elasticIndexer.ReceiptsIndex, buffSlice)
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
			return err
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ei *elasticProcessor) IsInterfaceNil() bool {
	return ei == nil
}
