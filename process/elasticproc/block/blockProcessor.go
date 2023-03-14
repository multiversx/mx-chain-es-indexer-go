package block

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
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

var log = logger.GetOrCreate("indexer/process/block")

type blockProcessor struct {
	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
}

// NewBlockProcessor will create a new instance of block processor
func NewBlockProcessor(hasher hashing.Hasher, marshalizer marshal.Marshalizer) (*blockProcessor, error) {
	if check.IfNil(hasher) {
		return nil, indexer.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, indexer.ErrNilMarshalizer
	}

	return &blockProcessor{
		hasher:      hasher,
		marshalizer: marshalizer,
	}, nil
}

// PrepareBlockForDB will prepare a database block and serialize it for database
func (bp *blockProcessor) PrepareBlockForDB(obh *outport.OutportBlockWithHeader) (*data.Block, error) {
	if check.IfNil(obh.Header) {
		return nil, indexer.ErrNilHeaderHandler
	}
	if obh.BlockData.Body == nil {
		return nil, indexer.ErrNilBlockBody
	}

	blockSizeInBytes, err := bp.computeBlockSize(obh.Header, obh.BlockData.Body)
	if err != nil {
		return nil, err
	}

	// TODO compute size of transactions
	sizeTxs := obh.OutportBlock.TransactionPool.Size()

	miniblocksHashes := bp.getEncodedMBSHashes(obh.BlockData.Body)
	leaderIndex := bp.getLeaderIndex(obh.SignersIndexes)

	numTxs, notarizedTxs := getTxsCount(obh.Header)
	elasticBlock := &data.Block{
		Nonce:                 obh.Header.GetNonce(),
		Round:                 obh.Header.GetRound(),
		Epoch:                 obh.Header.GetEpoch(),
		ShardID:               obh.Header.GetShardID(),
		Hash:                  hex.EncodeToString(obh.BlockData.HeaderHash),
		MiniBlocksHashes:      miniblocksHashes,
		NotarizedBlocksHashes: obh.NotarizedHeadersHashes,
		Proposer:              leaderIndex,
		Validators:            obh.SignersIndexes,
		PubKeyBitmap:          hex.EncodeToString(obh.Header.GetPubKeysBitmap()),
		Size:                  int64(blockSizeInBytes),
		SizeTxs:               int64(sizeTxs),
		Timestamp:             time.Duration(obh.Header.GetTimeStamp()),
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

	bp.addEpochStartInfoForMeta(obh.Header, elasticBlock)
	putMiniblocksDetailsInBlock(obh.Header, elasticBlock, obh.TransactionPool, obh.BlockData.Body)

	return elasticBlock, nil
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
		if mb.Type == block.PeerBlock {
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

func (bp *blockProcessor) getEncodedMBSHashes(body *block.Body) []string {
	miniblocksHashes := make([]string, 0)
	for _, miniblock := range body.MiniBlocks {
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

func putMiniblocksDetailsInBlock(header coreData.HeaderHandler, block *data.Block, pool *outport.TransactionPool, body *block.Body) {
	mbHeaders := header.GetMiniBlockHeaderHandlers()

	for idx, mbHeader := range mbHeaders {
		mbType := nodeBlock.Type(mbHeader.GetTypeInt32())
		if mbType == nodeBlock.PeerBlock {
			continue
		}

		txsHashes := body.MiniBlocks[idx].TxHashes
		block.MiniBlocksDetails = append(block.MiniBlocksDetails, &data.MiniBlocksDetails{
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

func (bp *blockProcessor) computeBlockSize(header coreData.HeaderHandler, body *block.Body) (int, error) {
	headerBytes, err := bp.marshalizer.Marshal(header)
	if err != nil {
		return 0, err
	}
	bodyBytes, err := bp.marshalizer.Marshal(body)
	if err != nil {
		return 0, err
	}

	blockSize := len(headerBytes) + len(bodyBytes)

	return blockSize, nil
}

func (bp *blockProcessor) getLeaderIndex(signersIndexes []uint64) uint64 {
	if len(signersIndexes) > 0 {
		return signersIndexes[0]
	}

	return 0
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
