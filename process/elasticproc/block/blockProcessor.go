package block

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/api"
	nodeBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	indexer "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const (
	notExecutedInCurrentBlock = -1
	notFound                  = -2
)

var (
	log                     = logger.GetOrCreate("indexer/process/block")
	errNilBlockData         = errors.New("nil block data")
	errNilHeaderGasConsumed = errors.New("nil header gas consumed data")
)

type blockProcessor struct {
	hasher                    hashing.Hasher
	marshalizer               marshal.Marshalizer
	validatorsPubKeyConverter core.PubkeyConverter
}

// NewBlockProcessor will create a new instance of block processor
func NewBlockProcessor(hasher hashing.Hasher, marshalizer marshal.Marshalizer, validatorsPubKeyConverter core.PubkeyConverter) (*blockProcessor, error) {
	if check.IfNil(hasher) {
		return nil, indexer.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, indexer.ErrNilMarshalizer
	}
	if check.IfNil(validatorsPubKeyConverter) {
		return nil, indexer.ErrNilPubkeyConverter
	}

	return &blockProcessor{
		hasher:                    hasher,
		marshalizer:               marshalizer,
		validatorsPubKeyConverter: validatorsPubKeyConverter,
	}, nil
}

// PrepareBlockForDB will prepare a database block and serialize it for database
func (bp *blockProcessor) PrepareBlockForDB(obh *outport.OutportBlockWithHeader) (*data.PreparedBlockResults, error) {
	if check.IfNil(obh.Header) {
		return nil, indexer.ErrNilHeaderHandler
	}
	if obh.BlockData == nil {
		return nil, errNilBlockData
	}
	if obh.BlockData.Body == nil {
		return nil, indexer.ErrNilBlockBody
	}
	if obh.HeaderGasConsumption == nil {
		return nil, errNilHeaderGasConsumed
	}

	blockSizeInBytes, err := bp.computeBlockSize(obh.BlockData.HeaderBytes, obh.BlockData.Body)
	if err != nil {
		return nil, err
	}

	sizeTxs := computeSizeOfTransactions(obh.TransactionPool)
	miniblocksHashes := bp.getEncodedMBSHashes(obh.BlockData.Body, obh.BlockData.IntraShardMiniBlocks)

	numTxs, notarizedTxs := getTxsCount(obh.Header)
	elasticBlock := &data.Block{
		Nonce:                 obh.Header.GetNonce(),
		Round:                 obh.Header.GetRound(),
		Epoch:                 obh.Header.GetEpoch(),
		ShardID:               obh.Header.GetShardID(),
		Hash:                  hex.EncodeToString(obh.BlockData.HeaderHash),
		MiniBlocksHashes:      miniblocksHashes,
		NotarizedBlocksHashes: obh.NotarizedHeadersHashes,
		Proposer:              getLeaderIndex(obh),
		ProposerBlsKey:        hex.EncodeToString(obh.LeaderBLSKey),
		Validators:            obh.SignersIndexes,
		PubKeyBitmap:          hex.EncodeToString(obh.Header.GetPubKeysBitmap()),
		Size:                  int64(blockSizeInBytes),
		SizeTxs:               int64(sizeTxs),
		Timestamp:             obh.Header.GetTimeStamp(),
		TimestampMs:           obh.OutportBlock.BlockData.GetTimestampMs(),
		TxCount:               numTxs,
		NotarizedTxsCount:     notarizedTxs,
		StateRootHash:         hex.EncodeToString(obh.Header.GetRootHash()),
		PrevHash:              hex.EncodeToString(obh.Header.GetPrevHash()),
		SearchOrder:           computeBlockSearchOrder(obh.Header),
		EpochStartBlock:       obh.Header.IsStartOfEpochBlock(),
		GasProvided:           obh.HeaderGasConsumption.GasProvided,
		GasRefunded:           obh.HeaderGasConsumption.GasRefunded,
		GasPenalized:          obh.HeaderGasConsumption.GasPenalized,
		MaxGasLimit:           obh.HeaderGasConsumption.MaxGasPerBlock,
		AccumulatedFees:       converters.BigIntToString(obh.Header.GetAccumulatedFees()),
		DeveloperFees:         converters.BigIntToString(obh.Header.GetDeveloperFees()),
		RandSeed:              hex.EncodeToString(obh.Header.GetRandSeed()),
		PrevRandSeed:          hex.EncodeToString(obh.Header.GetPrevRandSeed()),
		Signature:             hex.EncodeToString(obh.Header.GetSignature()),
		LeaderSignature:       hex.EncodeToString(obh.Header.GetLeaderSignature()),
		ChainID:               string(obh.Header.GetChainID()),
		SoftwareVersion:       hex.EncodeToString(obh.Header.GetSoftwareVersion()),
		ReceiptsHash:          hex.EncodeToString(obh.Header.GetReceiptsHash()),
		Reserved:              obh.Header.GetReserved(),
		UUID:                  converters.GenerateBase64UUID(),
	}

	additionalData := obh.Header.GetAdditionalData()
	if obh.Header.GetAdditionalData() != nil {
		elasticBlock.ScheduledData = &data.ScheduledData{
			ScheduledRootHash:        hex.EncodeToString(additionalData.GetScheduledRootHash()),
			ScheduledAccumulatedFees: converters.BigIntToString(additionalData.GetScheduledAccumulatedFees()),
			ScheduledDeveloperFees:   converters.BigIntToString(additionalData.GetScheduledDeveloperFees()),
			ScheduledGasProvided:     additionalData.GetScheduledGasProvided(),
			ScheduledGasPenalized:    additionalData.GetScheduledGasPenalized(),
			ScheduledGasRefunded:     additionalData.GetScheduledGasRefunded(),
		}
	}

	elasticBlock.ExecutionResultBlockHashes = make([]string, 0, len(obh.Header.GetExecutionResultsHandlers()))
	for _, executionResult := range obh.Header.GetExecutionResultsHandlers() {
		elasticBlock.ExecutionResultBlockHashes = append(elasticBlock.ExecutionResultBlockHashes, hex.EncodeToString(executionResult.GetHeaderHash()))
	}

	bp.addEpochStartInfoForMeta(obh.Header, elasticBlock)

	elasticBlock.MiniBlocksDetails = prepareMiniBlockDetails(obh.Header.GetMiniBlockHeaderHandlers(), obh.BlockData.Body, obh.TransactionPool)

	appendBlockDetailsFromIntraShardMbs(elasticBlock, obh.BlockData.IntraShardMiniBlocks, obh.TransactionPool, len(obh.Header.GetMiniBlockHeaderHandlers()))

	addProofs(elasticBlock, obh)

	executionResultData, err := bp.prepareExecutionResults(obh)
	if err != nil {
		return nil, err
	}

	return &data.PreparedBlockResults{
		Block:            elasticBlock,
		ExecutionResults: executionResultData,
	}, nil
}

func (bp *blockProcessor) prepareExecutionResults(obh *outport.OutportBlockWithHeader) ([]*data.ExecutionResult, error) {
	if !obh.Header.IsHeaderV3() {
		return []*data.ExecutionResult{}, nil
	}

	executionResults := make([]*data.ExecutionResult, 0)
	for _, executionResultHandler := range obh.Header.GetExecutionResultsHandlers() {
		executionResult := bp.prepareExecutionResult(executionResultHandler, obh)

		executionResults = append(executionResults, executionResult)
	}

	return executionResults, nil
}

func (bp *blockProcessor) prepareExecutionResult(baseExecutionResult coreData.BaseExecutionResultHandler, obh *outport.OutportBlockWithHeader) *data.ExecutionResult {
	miniblocksHashes := bp.getEncodedMBSHashes(obh.BlockData.Body, obh.BlockData.IntraShardMiniBlocks)

	executionResultsHash := hex.EncodeToString(baseExecutionResult.GetHeaderHash())
	executionResult := &data.ExecutionResult{
		UUID:                 converters.GenerateBase64UUID(),
		Hash:                 executionResultsHash,
		RootHash:             hex.EncodeToString(baseExecutionResult.GetRootHash()),
		NotarizedInBlockHash: hex.EncodeToString(obh.BlockData.GetHeaderHash()),
		Nonce:                baseExecutionResult.GetHeaderNonce(),
		Round:                baseExecutionResult.GetHeaderRound(),
		Epoch:                baseExecutionResult.GetHeaderEpoch(),
		MiniBlocksHashes:     miniblocksHashes,
		GasUsed:              baseExecutionResult.GetGasUsed(),
	}

	executionResultData, found := obh.BlockData.Results[executionResultsHash]
	if !found {
		log.Warn("cannot find execution result data for execution result", "hash", executionResultsHash)
		return executionResult
	}

	switch t := baseExecutionResult.(type) {
	case *nodeBlock.MetaExecutionResult:
		executionResult.MiniBlocksDetails = prepareMiniBlockDetails(t.GetMiniBlockHeadersHandlers(), executionResultData.Body, obh.TransactionPool)
		executionResult.AccumulatedFees = t.AccumulatedFees.String()
		executionResult.DeveloperFees = t.DeveloperFees.String()
		executionResult.TxCount = t.ExecutedTxCount
	case *nodeBlock.ExecutionResult:
		executionResult.MiniBlocksDetails = prepareMiniBlockDetails(t.GetMiniBlockHeadersHandlers(), executionResultData.Body, obh.TransactionPool)
		executionResult.AccumulatedFees = t.AccumulatedFees.String()
		executionResult.DeveloperFees = t.DeveloperFees.String()
		executionResult.TxCount = t.ExecutedTxCount
	default:
		return executionResult
	}

	return executionResult
}

func getLeaderIndex(obh *outport.OutportBlockWithHeader) uint64 {
	if obh.BlockData.HeaderProof != nil {
		return obh.LeaderIndex
	}

	if len(obh.SignersIndexes) > 0 {
		return obh.SignersIndexes[0]
	}

	return 0
}

func addProofs(elasticBlock *data.Block, obh *outport.OutportBlockWithHeader) {
	if obh.BlockData.HeaderProof != nil {
		elasticBlock.Proof = proofToAPIProof(obh.BlockData.HeaderProof)
		elasticBlock.PubKeyBitmap = elasticBlock.Proof.PubKeysBitmap
	}
}

func proofToAPIProof(headerProof coreData.HeaderProofHandler) *api.HeaderProof {
	return &api.HeaderProof{
		PubKeysBitmap:       hex.EncodeToString(headerProof.GetPubKeysBitmap()),
		AggregatedSignature: hex.EncodeToString(headerProof.GetAggregatedSignature()),
		HeaderHash:          hex.EncodeToString(headerProof.GetHeaderHash()),
		HeaderEpoch:         headerProof.GetHeaderEpoch(),
		HeaderNonce:         headerProof.GetHeaderNonce(),
		HeaderShardId:       headerProof.GetHeaderShardId(),
		HeaderRound:         headerProof.GetHeaderRound(),
		IsStartOfEpoch:      headerProof.GetIsStartOfEpoch(),
	}
}

func getTxsCount(header coreData.HeaderHandler) (numTxs, notarizedTxs uint32) {
	numTxs = header.GetTxCount()

	if core.MetachainShardId != header.GetShardID() {
		return numTxs, notarizedTxs
	}

	metaHeader, ok := header.(*nodeBlock.MetaBlock)
	if !ok {
		return 0, 0
	}

	notarizedTxs = metaHeader.TxCount
	numTxs = 0
	for _, mb := range metaHeader.MiniBlockHeaders {
		if mb.Type == nodeBlock.PeerBlock {
			continue
		}

		numTxs += mb.TxCount
	}

	notarizedTxs = notarizedTxs - numTxs

	return numTxs, notarizedTxs
}

func (bp *blockProcessor) addEpochStartInfoForMeta(header coreData.HeaderHandler, block *data.Block) {
	if header.GetShardID() != core.MetachainShardId {
		return
	}

	metaHeader, ok := header.(*nodeBlock.MetaBlock)
	if !ok {
		return
	}

	if !metaHeader.IsStartOfEpochBlock() {
		return
	}

	metaHeaderEconomics := metaHeader.EpochStart.Economics

	block.EpochStartInfo = &data.EpochStartInfo{
		TotalSupply:                      metaHeaderEconomics.TotalSupply.String(),
		TotalToDistribute:                metaHeaderEconomics.TotalToDistribute.String(),
		TotalNewlyMinted:                 metaHeaderEconomics.TotalNewlyMinted.String(),
		RewardsPerBlock:                  metaHeaderEconomics.RewardsPerBlock.String(),
		RewardsForProtocolSustainability: metaHeaderEconomics.RewardsForProtocolSustainability.String(),
		NodePrice:                        metaHeaderEconomics.NodePrice.String(),
		PrevEpochStartRound:              metaHeaderEconomics.PrevEpochStartRound,
		PrevEpochStartHash:               hex.EncodeToString(metaHeaderEconomics.PrevEpochStartHash),
	}
	if len(metaHeader.EpochStart.LastFinalizedHeaders) == 0 {
		return
	}

	epochStartShardsData := metaHeader.EpochStart.LastFinalizedHeaders
	block.EpochStartShardsData = make([]*data.EpochStartShardData, 0, len(metaHeader.EpochStart.LastFinalizedHeaders))
	for _, epochStartShardData := range epochStartShardsData {
		bp.addEpochStartShardDataForMeta(epochStartShardData, block)
	}
}

func (bp *blockProcessor) addEpochStartShardDataForMeta(epochStartShardData nodeBlock.EpochStartShardData, block *data.Block) {
	shardData := &data.EpochStartShardData{
		ShardID:               epochStartShardData.ShardID,
		Epoch:                 epochStartShardData.Epoch,
		Round:                 epochStartShardData.Round,
		Nonce:                 epochStartShardData.Nonce,
		HeaderHash:            hex.EncodeToString(epochStartShardData.HeaderHash),
		RootHash:              hex.EncodeToString(epochStartShardData.RootHash),
		ScheduledRootHash:     hex.EncodeToString(epochStartShardData.ScheduledRootHash),
		FirstPendingMetaBlock: hex.EncodeToString(epochStartShardData.FirstPendingMetaBlock),
		LastFinishedMetaBlock: hex.EncodeToString(epochStartShardData.LastFinishedMetaBlock),
	}

	if len(epochStartShardData.PendingMiniBlockHeaders) == 0 {
		block.EpochStartShardsData = append(block.EpochStartShardsData, shardData)
		return
	}

	shardData.PendingMiniBlockHeaders = make([]*data.Miniblock, 0, len(epochStartShardData.PendingMiniBlockHeaders))
	for _, pendingMb := range epochStartShardData.PendingMiniBlockHeaders {
		shardData.PendingMiniBlockHeaders = append(shardData.PendingMiniBlockHeaders, &data.Miniblock{
			Hash:            hex.EncodeToString(pendingMb.Hash),
			SenderShardID:   pendingMb.SenderShardID,
			ReceiverShardID: pendingMb.ReceiverShardID,
			Type:            pendingMb.Type.String(),
			Reserved:        pendingMb.Reserved,
		})
	}

	block.EpochStartShardsData = append(block.EpochStartShardsData, shardData)
}

func (bp *blockProcessor) getEncodedMBSHashes(body *nodeBlock.Body, intraShardMbs []*nodeBlock.MiniBlock) []string {
	miniblocksHashes := make([]string, 0)
	mbs := append(body.MiniBlocks, intraShardMbs...)
	for _, miniblock := range mbs {
		mbHash, errComputeHash := core.CalculateHash(bp.marshalizer, bp.hasher, miniblock)
		if errComputeHash != nil {
			log.Warn("internal error computing hash", "error", errComputeHash)

			continue
		}

		encodedMbHash := hex.EncodeToString(mbHash)
		miniblocksHashes = append(miniblocksHashes, encodedMbHash)
	}

	return miniblocksHashes
}

func prepareMiniBlockDetails(mbHeaders []coreData.MiniBlockHeaderHandler, body *nodeBlock.Body, pool *outport.TransactionPool) []*data.MiniBlocksDetails {
	mbsDetails := make([]*data.MiniBlocksDetails, 0, len(mbHeaders))
	for idx, mbHeader := range mbHeaders {
		mbType := nodeBlock.Type(mbHeader.GetTypeInt32())
		if mbType == nodeBlock.PeerBlock {
			continue
		}

		txsHashes := body.MiniBlocks[idx].TxHashes
		mbsDetails = append(mbsDetails, &data.MiniBlocksDetails{
			IndexFirstProcessedTx:    mbHeader.GetIndexOfFirstTxProcessed(),
			IndexLastProcessedTx:     mbHeader.GetIndexOfLastTxProcessed(),
			MBIndex:                  idx,
			ProcessingType:           nodeBlock.ProcessingType(mbHeader.GetProcessingType()).String(),
			Type:                     mbType.String(),
			SenderShardID:            mbHeader.GetSenderShardID(),
			ReceiverShardID:          mbHeader.GetReceiverShardID(),
			TxsHashes:                hexEncodeSlice(txsHashes),
			ExecutionOrderTxsIndices: extractExecutionOrderIndicesFromPool(mbHeader, txsHashes, pool),
		})
	}

	return mbsDetails
}

func appendBlockDetailsFromIntraShardMbs(block *data.Block, intraShardMbs []*nodeBlock.MiniBlock, pool *outport.TransactionPool, offset int) {
	for idx, intraMB := range intraShardMbs {
		if intraMB.Type == nodeBlock.PeerBlock || intraMB.Type == nodeBlock.ReceiptBlock {
			continue
		}

		block.MiniBlocksDetails = append(block.MiniBlocksDetails, &data.MiniBlocksDetails{
			IndexFirstProcessedTx:    0,
			IndexLastProcessedTx:     int32(len(intraMB.GetTxHashes()) - 1),
			SenderShardID:            intraMB.GetSenderShardID(),
			ReceiverShardID:          intraMB.GetReceiverShardID(),
			MBIndex:                  idx + offset,
			Type:                     intraMB.Type.String(),
			ProcessingType:           nodeBlock.Normal.String(),
			TxsHashes:                hexEncodeSlice(intraMB.TxHashes),
			ExecutionOrderTxsIndices: extractExecutionOrderIntraShardMBUnsigned(intraMB, pool),
		})
	}
}

func extractExecutionOrderIntraShardMBUnsigned(mb *nodeBlock.MiniBlock, pool *outport.TransactionPool) []int {
	executionOrderTxsIndices := make([]int, len(mb.TxHashes))
	for idx, txHash := range mb.TxHashes {
		executionOrder, found := getExecutionOrderForTx(txHash, int32(mb.Type), pool)
		if !found {
			log.Warn("blockProcessor.extractExecutionOrderIntraShardMBUnsigned cannot find tx in pool", "txHash", hex.EncodeToString(txHash))
			executionOrderTxsIndices[idx] = notFound
			continue
		}

		executionOrderTxsIndices[idx] = int(executionOrder)
	}

	return executionOrderTxsIndices
}

func extractExecutionOrderIndicesFromPool(mbHeader coreData.MiniBlockHeaderHandler, txsHashes [][]byte, pool *outport.TransactionPool) []int {
	mbType := mbHeader.GetTypeInt32()
	executionOrderTxsIndices := make([]int, len(txsHashes))
	indexOfFirstTxProcessed, indexOfLastTxProcessed := mbHeader.GetIndexOfFirstTxProcessed(), mbHeader.GetIndexOfLastTxProcessed()
	for idx, txHash := range txsHashes {
		isExecutedInCurrentBlock := int32(idx) >= indexOfFirstTxProcessed && int32(idx) <= indexOfLastTxProcessed
		if !isExecutedInCurrentBlock {
			executionOrderTxsIndices[idx] = notExecutedInCurrentBlock
			continue
		}

		executionOrder, found := getExecutionOrderForTx(txHash, mbType, pool)
		if !found {
			log.Warn("blockProcessor.extractExecutionOrderIndicesFromPool cannot find tx in pool", "txHash", hex.EncodeToString(txHash))
			executionOrderTxsIndices[idx] = notFound
			continue
		}

		executionOrderTxsIndices[idx] = int(executionOrder)
	}

	return executionOrderTxsIndices
}

type executionOrderHandler interface {
	GetExecutionOrder() uint32
}

func getExecutionOrderForTx(txHash []byte, mbType int32, pool *outport.TransactionPool) (uint32, bool) {
	var tx executionOrderHandler
	var found bool

	switch nodeBlock.Type(mbType) {
	case nodeBlock.TxBlock:
		tx, found = pool.Transactions[hex.EncodeToString(txHash)]
	case nodeBlock.InvalidBlock:
		tx, found = pool.InvalidTxs[hex.EncodeToString(txHash)]
	case nodeBlock.RewardsBlock:
		tx, found = pool.Rewards[hex.EncodeToString(txHash)]
	case nodeBlock.SmartContractResultBlock:
		tx, found = pool.SmartContractResults[hex.EncodeToString(txHash)]
	default:
		return 0, false
	}

	if !found {
		return 0, false
	}
	return tx.GetExecutionOrder(), true
}

func (bp *blockProcessor) computeBlockSize(headerBytes []byte, body *nodeBlock.Body) (int, error) {
	bodyBytes, err := bp.marshalizer.Marshal(body)
	if err != nil {
		return 0, err
	}

	blockSize := len(headerBytes) + len(bodyBytes)

	return blockSize, nil
}

func computeBlockSearchOrder(header coreData.HeaderHandler) uint64 {
	shardIdentifier := createShardIdentifier(header.GetShardID())
	stringOrder := fmt.Sprintf("1%02d%d", shardIdentifier, header.GetNonce())

	order, err := strconv.ParseUint(stringOrder, 10, 64)
	if err != nil {
		log.Debug("elasticsearchDatabase.computeBlockSearchOrder",
			"could not set uint32 search order", err.Error())
		return 0
	}

	return order
}

func createShardIdentifier(shardID uint32) uint32 {
	shardIdentifier := shardID + 2
	if shardID == core.MetachainShardId {
		shardIdentifier = 1
	}

	return shardIdentifier
}

// ComputeHeaderHash will compute the hash of a provided header
func (bp *blockProcessor) ComputeHeaderHash(header coreData.HeaderHandler) ([]byte, error) {
	return core.CalculateHash(bp.marshalizer, bp.hasher, header)
}

func hexEncodeSlice(slice [][]byte) []string {
	res := make([]string, 0, len(slice))
	for _, s := range slice {
		res = append(res, hex.EncodeToString(s))
	}
	return res
}

func computeSizeOfTransactions(pool *outport.TransactionPool) int {
	if pool == nil {
		return 0
	}

	txsSize := 0
	for _, txInfo := range pool.Transactions {
		txsSize += txInfo.Transaction.Size()
	}
	for _, rewardInfo := range pool.Rewards {
		txsSize += rewardInfo.Reward.Size()
	}
	for _, invalidTxInfo := range pool.InvalidTxs {
		txsSize += invalidTxInfo.Transaction.Size()
	}
	for _, scrInfo := range pool.SmartContractResults {
		txsSize += scrInfo.SmartContractResult.Size()
	}

	return txsSize
}
