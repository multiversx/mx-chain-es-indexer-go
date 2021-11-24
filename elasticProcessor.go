package indexer

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

const numDecimalsInFloatBalance = 10

type elasticProcessor struct {
	*txDatabaseProcessor

	elasticClient          DatabaseClientHandler
	parser                 *dataParser
	enabledIndexes         map[string]struct{}
	accountsDB             AccountsAdapter
	dividerForDenomination float64
	balancePrecision       float64
}

// NewElasticProcessor creates an elasticsearch es and handles saving
func NewElasticProcessor(arguments ArgElasticProcessor) (ElasticProcessor, error) {
	err := checkArgElasticProcessor(arguments)
	if err != nil {
		return nil, err
	}

	ei := &elasticProcessor{
		elasticClient: arguments.DBClient,
		parser: &dataParser{
			hasher:      arguments.Hasher,
			marshalizer: arguments.Marshalizer,
		},
		enabledIndexes:         arguments.EnabledIndexes,
		accountsDB:             arguments.AccountsDB,
		balancePrecision:       math.Pow(10, float64(numDecimalsInFloatBalance)),
		dividerForDenomination: math.Pow(10, float64(core.MaxInt(arguments.Denomination, 0))),
	}

	ei.txDatabaseProcessor = newTxDatabaseProcessor(
		arguments.Hasher,
		arguments.Marshalizer,
		arguments.AddressPubkeyConverter,
		arguments.ValidatorPubkeyConverter,
		arguments.TransactionFeeCalculator,
		arguments.IsInImportDBMode,
		arguments.ShardCoordinator,
	)

	if arguments.IsInImportDBMode {
		log.Warn("the node is in import mode! Cross shard transactions and rewards where destination shard is " +
			"not the current node's shard won't be indexed in Elastic Search")
	}

	if arguments.UseKibana {
		err = ei.initWithKibana(arguments.IndexTemplates, arguments.IndexPolicies)
		if err != nil {
			return nil, err
		}
	} else {
		err = ei.initNoKibana(arguments.IndexTemplates)
		if err != nil {
			return nil, err
		}
	}

	return ei, nil
}

func checkArgElasticProcessor(arguments ArgElasticProcessor) error {
	if check.IfNil(arguments.DBClient) {
		return ErrNilDatabaseClient
	}
	if check.IfNil(arguments.Marshalizer) {
		return core.ErrNilMarshalizer
	}
	if check.IfNil(arguments.Hasher) {
		return core.ErrNilHasher
	}
	if check.IfNil(arguments.AddressPubkeyConverter) {
		return ErrNilPubkeyConverter
	}
	if check.IfNil(arguments.ValidatorPubkeyConverter) {
		return ErrNilPubkeyConverter
	}
	if check.IfNil(arguments.AccountsDB) {
		return ErrNilAccountsDB
	}
	if check.IfNil(arguments.ShardCoordinator) {
		return ErrNilShardCoordinator
	}

	return nil
}

func (ei *elasticProcessor) initWithKibana(indexTemplates, indexPolicies map[string]*bytes.Buffer) error {
	err := ei.createOpenDistroTemplates(indexTemplates)
	if err != nil {
		return err
	}

	//TODO: Re-activate after we think of a solid way to handle forks+rotating indexes
	// err = ei.createIndexPolicies(indexPolicies)
	// if err != nil {
	// 	return err
	// }

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

func (ei *elasticProcessor) initNoKibana(indexTemplates map[string]*bytes.Buffer) error {
	err := ei.createOpenDistroTemplates(indexTemplates)
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

	return nil
}

func (ei *elasticProcessor) createIndexPolicies(indexPolicies map[string]*bytes.Buffer) error {
	indexesPolicies := []string{txPolicy, blockPolicy, miniblocksPolicy, ratingPolicy, roundPolicy, validatorsPolicy, accountsHistoryPolicy}
	for _, indexPolicyName := range indexesPolicies {
		indexPolicy := getTemplateByName(indexPolicyName, indexPolicies)
		if indexPolicy != nil {
			err := ei.elasticClient.CheckAndCreatePolicy(indexPolicyName, indexPolicy)
			if err != nil {
				log.Error("check and create policy", "policy", indexPolicy, "err", err)
				return err
			}
		}
	}

	return nil
}

func (ei *elasticProcessor) createOpenDistroTemplates(indexTemplates map[string]*bytes.Buffer) error {
	opendistroTemplate := getTemplateByName("opendistro", indexTemplates)
	if opendistroTemplate != nil {
		err := ei.elasticClient.CheckAndCreateTemplate("opendistro", opendistroTemplate)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ei *elasticProcessor) createIndexTemplates(indexTemplates map[string]*bytes.Buffer) error {
	indexes := []string{txIndex, blockIndex, miniblocksIndex, ratingIndex, roundIndex, validatorsIndex, accountsIndex, accountsHistoryIndex}
	for _, index := range indexes {
		indexTemplate := getTemplateByName(index, indexTemplates)
		if indexTemplate != nil {
			err := ei.elasticClient.CheckAndCreateTemplate(index, indexTemplate)
			if err != nil {
				log.Error("check and create template", "err", err)
				return err
			}
		}
	}
	return nil
}

func (ei *elasticProcessor) createIndexes() error {
	indexes := []string{txIndex, blockIndex, miniblocksIndex, ratingIndex, roundIndex, validatorsIndex, accountsIndex, accountsHistoryIndex}
	for _, index := range indexes {
		indexName := fmt.Sprintf("%s-000001", index)
		err := ei.elasticClient.CheckAndCreateIndex(indexName)
		if err != nil {
			log.Error("check and create index", "err", err)
			return err
		}
	}
	return nil
}

func (ei *elasticProcessor) createAliases() error {
	indexes := []string{txIndex, blockIndex, miniblocksIndex, ratingIndex, roundIndex, validatorsIndex, accountsIndex, accountsHistoryIndex}
	for _, index := range indexes {
		indexName := fmt.Sprintf("%s-000001", index)
		err := ei.elasticClient.CheckAndCreateAlias(index, indexName)
		if err != nil {
			log.Error("check and create alias", "err", err)
			return err
		}
	}

	return nil
}

func (ei *elasticProcessor) getExistingObjMap(hashes []string, index string) (map[string]bool, error) {
	if len(hashes) == 0 {
		return make(map[string]bool), nil
	}

	response, err := ei.elasticClient.DoMultiGet(getDocumentsByIDsQuery(hashes), index)
	if err != nil {
		return nil, err
	}

	return getDecodedResponseMultiGet(response), nil
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
	gasConsumptionData indexer.HeaderGasConsumption,
	txsSize int,
) error {
	if !ei.isIndexEnabled(blockIndex) {
		return nil
	}

	var buff bytes.Buffer

	serializedBlock, headerHash, err := ei.parser.getSerializedElasticBlockAndHeaderHash(header, signersIndexes, body, notarizedHeadersHashes, gasConsumptionData, txsSize)
	if err != nil {
		return err
	}

	buff.Grow(len(serializedBlock))
	_, err = buff.Write(serializedBlock)
	if err != nil {
		return err
	}

	req := &esapi.IndexRequest{
		Index:      blockIndex,
		DocumentID: hex.EncodeToString(headerHash),
		Body:       bytes.NewReader(buff.Bytes()),
		Refresh:    "true",
	}

	return ei.elasticClient.DoRequest(req)
}

// RemoveHeader will remove a block from elasticsearch server
func (ei *elasticProcessor) RemoveHeader(header coreData.HeaderHandler) error {
	headerHash, err := core.CalculateHash(ei.marshalizer, ei.hasher, header)
	if err != nil {
		return err
	}

	return ei.elasticClient.DoBulkRemove(blockIndex, []string{hex.EncodeToString(headerHash)})
}

// RemoveMiniblocks will remove all miniblocks that are in header from elasticsearch server
func (ei *elasticProcessor) RemoveMiniblocks(header coreData.HeaderHandler, body *block.Body) error {
	if body == nil || len(header.GetMiniBlockHeadersHashes()) == 0 {
		return nil
	}

	encodedMiniblocksHashes := make([]string, 0)
	selfShardID := header.GetShardID()
	for _, miniblock := range body.MiniBlocks {
		if miniblock.Type == block.PeerBlock {
			continue
		}

		isDstMe := selfShardID == miniblock.ReceiverShardID
		isCrossShard := miniblock.ReceiverShardID != miniblock.SenderShardID
		if isDstMe && isCrossShard {
			continue
		}

		miniblockHash, err := core.CalculateHash(ei.marshalizer, ei.hasher, miniblock)
		if err != nil {
			log.Debug("indexer.RemoveMiniblocks cannot calculate miniblock hash",
				"error", err.Error())
			continue
		}
		encodedMiniblocksHashes = append(encodedMiniblocksHashes, hex.EncodeToString(miniblockHash))

	}

	return ei.elasticClient.DoBulkRemove(miniblocksIndex, encodedMiniblocksHashes)
}

// RemoveTransactions will remove transactions that are in miniblock from the elasticsearch server
func (ei *elasticProcessor) RemoveTransactions(header coreData.HeaderHandler, body *block.Body) error {
	// TODO next PR will come with the implementation of this method
	return nil
}

// SaveMiniblocks will prepare and save information about miniblocks in elasticsearch server
func (ei *elasticProcessor) SaveMiniblocks(header coreData.HeaderHandler, body *block.Body) (map[string]bool, error) {
	if !ei.isIndexEnabled(miniblocksIndex) {
		return map[string]bool{}, nil
	}

	miniblocks := ei.parser.getMiniblocks(header, body)
	if len(miniblocks) == 0 {
		return make(map[string]bool), nil
	}

	buff, mbHashDb := serializeBulkMiniBlocks(header.GetShardID(), miniblocks, ei.getExistingObjMap)
	return mbHashDb, ei.elasticClient.DoBulkRequest(&buff, miniblocksIndex)
}

// SaveTransactions will prepare and save information about a transactions in elasticsearch server
func (ei *elasticProcessor) SaveTransactions(
	body *block.Body,
	header coreData.HeaderHandler,
	pool *indexer.Pool,
	mbsInDb map[string]bool,
) error {
	if !ei.isIndexEnabled(txIndex) {
		return nil
	}

	sliceMaps := []map[string]coreData.TransactionHandler{
		pool.Txs, pool.Scrs, pool.Receipts, pool.Invalid, pool.Rewards,
	}
	allTxs := mergeSliceOfMaps(sliceMaps)

	selfShardID := ei.shardCoordinator.SelfId()
	txs, alteredAccounts := ei.prepareTransactionsForDatabase(body, header, allTxs, selfShardID, pool.Logs)
	buffSlice, err := serializeTransactions(txs, selfShardID, ei.getExistingObjMap, mbsInDb)
	if err != nil {
		return err
	}

	for idx := range buffSlice {
		err = ei.elasticClient.DoBulkRequest(&buffSlice[idx], txIndex)
		if err != nil {
			log.Warn("indexer indexing bulk of transactions",
				"error", err.Error())
			return err
		}
	}

	return ei.indexAlteredAccounts(header.GetTimeStamp(), alteredAccounts)
}

func mergeSliceOfMaps(sliceMaps []map[string]coreData.TransactionHandler) map[string]coreData.TransactionHandler {
	allTxs := make(map[string]coreData.TransactionHandler)

	for _, txsMap := range sliceMaps {
		for hash, tx := range txsMap {
			allTxs[hash] = tx
		}
	}

	return allTxs
}

// SaveValidatorsRating will save validators rating
func (ei *elasticProcessor) SaveValidatorsRating(index string, validatorsRatingInfo []*data.ValidatorRatingInfo) error {
	if !ei.isIndexEnabled(ratingIndex) {
		return nil
	}

	var buff bytes.Buffer

	infosRating := data.ValidatorsRatingInfo{ValidatorsInfos: validatorsRatingInfo}

	marshalizedInfoRating, err := json.Marshal(&infosRating)
	if err != nil {
		log.Debug("indexer: marshal", "error", "could not marshal validators rating")
		return err
	}

	buff.Grow(len(marshalizedInfoRating))
	_, err = buff.Write(marshalizedInfoRating)
	if err != nil {
		log.Warn("elastic search: save validators rating, write", "error", err.Error())
	}

	req := &esapi.IndexRequest{
		Index:      ratingIndex,
		DocumentID: index,
		Body:       bytes.NewReader(buff.Bytes()),
		Refresh:    "true",
	}

	return ei.elasticClient.DoRequest(req)
}

// SaveShardValidatorsPubKeys will prepare and save information about a shard validators public keys in elasticsearch server
func (ei *elasticProcessor) SaveShardValidatorsPubKeys(shardID, epoch uint32, shardValidatorsPubKeys [][]byte) error {
	if !ei.isIndexEnabled(validatorsIndex) {
		return nil
	}

	var buff bytes.Buffer

	shardValPubKeys := data.ValidatorsPublicKeys{
		PublicKeys: make([]string, 0, len(shardValidatorsPubKeys)),
	}
	for _, validatorPk := range shardValidatorsPubKeys {
		strValidatorPk := ei.validatorPubkeyConverter.Encode(validatorPk)
		shardValPubKeys.PublicKeys = append(shardValPubKeys.PublicKeys, strValidatorPk)
	}

	marshalizedValidatorPubKeys, err := json.Marshal(shardValPubKeys)
	if err != nil {
		log.Debug("indexer: marshal", "error", "could not marshal validators public keys")
		return err
	}

	buff.Grow(len(marshalizedValidatorPubKeys))
	_, err = buff.Write(marshalizedValidatorPubKeys)
	if err != nil {
		log.Warn("elastic search: save shard validators pub keys, write", "error", err.Error())
	}

	req := &esapi.IndexRequest{
		Index:      validatorsIndex,
		DocumentID: fmt.Sprintf("%d_%d", shardID, epoch),
		Body:       bytes.NewReader(buff.Bytes()),
		Refresh:    "true",
	}

	return ei.elasticClient.DoRequest(req)
}

// SaveRoundsInfo will prepare and save information about a slice of rounds in elasticsearch server
func (ei *elasticProcessor) SaveRoundsInfo(infos []*data.RoundInfo) error {
	if !ei.isIndexEnabled(roundIndex) {
		return nil
	}

	var buff bytes.Buffer

	for _, info := range infos {
		serializedRoundInfo, meta := serializeRoundInfo(*info)

		buff.Grow(len(meta) + len(serializedRoundInfo))
		_, err := buff.Write(meta)
		if err != nil {
			log.Warn("indexer: cannot write meta", "error", err.Error())
		}

		_, err = buff.Write(serializedRoundInfo)
		if err != nil {
			log.Warn("indexer: cannot write serialized round info", "error", err.Error())
		}
	}

	return ei.elasticClient.DoBulkRequest(&buff, roundIndex)
}

func (ei *elasticProcessor) indexAlteredAccounts(blockTimestamp uint64, accounts map[string]struct{}) error {
	if !ei.isIndexEnabled(accountsIndex) {
		return nil
	}

	accountsToIndex := make([]coreData.UserAccountHandler, 0)
	for address := range accounts {
		addressBytes, err := ei.addressPubkeyConverter.Decode(address)
		if err != nil {
			log.Warn("cannot decode address", "address", address, "error", err)
			continue
		}

		if ei.shardCoordinator.ComputeId(addressBytes) != ei.shardCoordinator.SelfId() {
			continue
		}

		account, err := ei.accountsDB.LoadAccount(addressBytes)
		if err != nil {
			log.Warn("cannot load account", "address bytes", addressBytes, "error", err)
			continue
		}

		userAccount, ok := account.(coreData.UserAccountHandler)
		if !ok {
			log.Warn("cannot cast AccountHandler to type UserAccountHandler")
			continue
		}

		accountsToIndex = append(accountsToIndex, userAccount)
	}

	if len(accountsToIndex) == 0 {
		log.Debug("no account to index from provided transactions")
		return nil
	}

	accountsSlice := make([]*data.Account, len(accountsToIndex))
	for idx, account := range accountsToIndex {
		accountsSlice[idx] = &data.Account{
			UserAccount: account,
			IsSender:    false,
		}
	}

	return ei.SaveAccounts(blockTimestamp, accountsSlice)
}

// SaveAccounts will prepare and save information about provided accounts in elasticsearch server
func (ei *elasticProcessor) SaveAccounts(blockTimestamp uint64, accountsSlice []*data.Account) error {
	if !ei.isIndexEnabled(accountsIndex) {
		return nil
	}

	accountsMap := make(map[string]*data.AccountInfo)
	for _, userAccount := range accountsSlice {
		balanceAsFloat := ei.computeBalanceAsFloat(userAccount.UserAccount.GetBalance())
		acc := &data.AccountInfo{
			Nonce:      userAccount.UserAccount.GetNonce(),
			Balance:    userAccount.UserAccount.GetBalance().String(),
			BalanceNum: balanceAsFloat,
		}
		address := ei.addressPubkeyConverter.Encode(userAccount.UserAccount.AddressBytes())
		accountsMap[address] = acc
	}

	buffSlice, err := serializeAccounts(accountsMap)
	if err != nil {
		return err
	}
	for idx := range buffSlice {
		err = ei.elasticClient.DoBulkRequest(&buffSlice[idx], accountsIndex)
		if err != nil {
			log.Warn("indexer: indexing bulk of accounts",
				"error", err.Error())
			return err
		}
	}

	return ei.saveAccountsHistory(blockTimestamp, accountsMap)
}

func (ei *elasticProcessor) saveAccountsHistory(blockTimestamp uint64, accountsInfoMap map[string]*data.AccountInfo) error {
	if !ei.isIndexEnabled(accountsHistoryIndex) {
		return nil
	}

	accountsMap := make(map[string]*data.AccountBalanceHistory)
	for address, userAccount := range accountsInfoMap {
		acc := &data.AccountBalanceHistory{
			Address:   address,
			Balance:   userAccount.Balance,
			Timestamp: time.Duration(blockTimestamp),
		}
		addressKey := fmt.Sprintf("%s_%d", address, blockTimestamp)
		accountsMap[addressKey] = acc
	}

	buffSlice, err := serializeAccountsHistory(accountsMap)
	if err != nil {
		return err
	}
	for idx := range buffSlice {
		err = ei.elasticClient.DoBulkRequest(&buffSlice[idx], accountsHistoryIndex)
		if err != nil {
			log.Warn("indexer: indexing bulk of accounts history",
				"error", err.Error())
			return err
		}
	}

	return nil
}

func (ei *elasticProcessor) isIndexEnabled(index string) bool {
	_, isEnabled := ei.enabledIndexes[index]
	return isEnabled
}

func (ei *elasticProcessor) computeBalanceAsFloat(balance *big.Int) float64 {
	balanceBigFloat := big.NewFloat(0).SetInt(balance)
	balanceFloat64, _ := balanceBigFloat.Float64()

	bal := balanceFloat64 / ei.dividerForDenomination
	balanceFloatWithDecimals := math.Round(bal*ei.balancePrecision) / ei.balancePrecision

	return core.MaxFloat64(balanceFloatWithDecimals, 0)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ei *elasticProcessor) IsInterfaceNil() bool {
	return ei == nil
}
