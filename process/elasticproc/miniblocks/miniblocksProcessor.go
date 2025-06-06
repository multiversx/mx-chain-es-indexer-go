package miniblocks

import (
	"encoding/hex"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("indexer/process/miniblocks")

type miniblocksProcessor struct {
	hasher     hashing.Hasher
	marshalier marshal.Marshalizer
}

// NewMiniblocksProcessor will create a new instance of miniblocksProcessor
func NewMiniblocksProcessor(
	hasher hashing.Hasher,
	marshalier marshal.Marshalizer,
) (*miniblocksProcessor, error) {
	if check.IfNil(marshalier) {
		return nil, dataindexer.ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, dataindexer.ErrNilHasher
	}

	return &miniblocksProcessor{
		hasher:     hasher,
		marshalier: marshalier,
	}, nil
}

// PrepareDBMiniblocks will prepare miniblocks
func (mp *miniblocksProcessor) PrepareDBMiniblocks(header coreData.HeaderHandler, miniBlocks []*block.MiniBlock, timestampMS uint64) []*data.Miniblock {
	headerHash, err := mp.calculateHash(header)
	if err != nil {
		log.Warn("indexer: could not calculate header hash", "error", err)
		return nil
	}

	selfShard := header.GetShardID()
	dbMiniblocks := make([]*data.Miniblock, 0)
	for mbIndex, miniBlock := range miniBlocks {
		if miniBlock.ReceiverShardID == core.AllShardId && selfShard != core.MetachainShardId {
			// will not index the miniblock on the destination if is for all shards
			continue
		}

		dbMiniBlock, errPrepareMiniBlock := mp.prepareMiniblockForDB(mbIndex, miniBlock, header, headerHash, timestampMS)
		if errPrepareMiniBlock != nil {
			log.Warn("miniblocksProcessor.PrepareDBMiniBlocks cannot prepare miniblock", "error", errPrepareMiniBlock)
			continue
		}

		dbMiniblocks = append(dbMiniblocks, dbMiniBlock)
	}

	return dbMiniblocks
}

func (mp *miniblocksProcessor) prepareMiniblockForDB(
	mbIndex int,
	miniblock *block.MiniBlock,
	header coreData.HeaderHandler,
	headerHash []byte,
	timestampMS uint64,
) (*data.Miniblock, error) {
	mbHash, err := mp.calculateHash(miniblock)
	if err != nil {
		return nil, err
	}

	encodedMbHash := hex.EncodeToString(mbHash)

	dbMiniblock := &data.Miniblock{
		Hash:            encodedMbHash,
		SenderShardID:   miniblock.SenderShardID,
		ReceiverShardID: miniblock.ReceiverShardID,
		Type:            miniblock.Type.String(),
		Timestamp:       time.Duration(header.GetTimeStamp()),
		TimestampMs:     time.Duration(timestampMS),
		Reserved:        miniblock.Reserved,
	}

	encodedHeaderHash := hex.EncodeToString(headerHash)
	isIntraShard := dbMiniblock.SenderShardID == dbMiniblock.ReceiverShardID
	isCrossOnSource := !isIntraShard && dbMiniblock.SenderShardID == header.GetShardID()
	if isIntraShard || isCrossOnSource {
		mp.setFieldsMBIntraShardAndCrossFromMe(mbIndex, header, encodedHeaderHash, dbMiniblock, isIntraShard)

		return dbMiniblock, nil
	}

	processingType, _ := mp.computeProcessingTypeAndConstructionState(mbIndex, header)
	dbMiniblock.ProcessingTypeOnDestination = processingType
	dbMiniblock.ReceiverBlockHash = encodedHeaderHash

	return dbMiniblock, nil
}

func (mp *miniblocksProcessor) setFieldsMBIntraShardAndCrossFromMe(
	mbIndex int,
	header coreData.HeaderHandler,
	headerHash string,
	dbMiniblock *data.Miniblock,
	isIntraShard bool,
) {
	processingType, constructionState := mp.computeProcessingTypeAndConstructionState(mbIndex, header)

	dbMiniblock.ProcessingTypeOnSource = processingType
	switch {
	case constructionState == int32(block.Final) && processingType == block.Normal.String():
		dbMiniblock.SenderBlockHash = headerHash
		dbMiniblock.ProcessingTypeOnSource = processingType
		if isIntraShard {
			dbMiniblock.ReceiverBlockHash = headerHash
			dbMiniblock.ProcessingTypeOnDestination = processingType
		}
	case constructionState == int32(block.Proposed) && processingType == block.Scheduled.String():
		dbMiniblock.SenderBlockHash = headerHash
		dbMiniblock.ProcessingTypeOnSource = processingType
	case constructionState == int32(block.Final) && processingType == block.Processed.String():
		dbMiniblock.ReceiverBlockHash = headerHash
		dbMiniblock.ProcessingTypeOnDestination = processingType
	}
}

func (mp *miniblocksProcessor) computeProcessingTypeAndConstructionState(mbIndex int, header coreData.HeaderHandler) (string, int32) {
	miniblockHeaders := header.GetMiniBlockHeaderHandlers()
	if len(miniblockHeaders) <= mbIndex {
		return block.Normal.String(), int32(block.Final)
	}

	processingType := miniblockHeaders[mbIndex].GetProcessingType()
	constructionState := miniblockHeaders[mbIndex].GetConstructionState()

	switch processingType {
	case int32(block.Scheduled):
		return block.Scheduled.String(), constructionState
	case int32(block.Processed):
		return block.Processed.String(), constructionState
	default:
		return block.Normal.String(), constructionState
	}
}

// GetMiniblocksHashesHexEncoded will compute miniblocks hashes in a hexadecimal encoding
func (mp *miniblocksProcessor) GetMiniblocksHashesHexEncoded(header coreData.HeaderHandler, body *block.Body) []string {
	if body == nil || len(header.GetMiniBlockHeadersHashes()) == 0 {
		return nil
	}

	encodedMiniblocksHashes := make([]string, 0)
	selfShardID := header.GetShardID()
	for _, miniblock := range body.MiniBlocks {
		if miniblock.Type == block.PeerBlock {
			continue
		}

		isCrossShard := miniblock.ReceiverShardID != miniblock.SenderShardID
		isCrossShardOnDestination := selfShardID == miniblock.ReceiverShardID && isCrossShard
		if isCrossShardOnDestination {
			continue
		}

		miniblockHash, err := mp.calculateHash(miniblock)
		if err != nil {
			log.Debug("miniblocksProcessor.GetMiniblocksHashesHexEncoded cannot calculate miniblock hash",
				"error", err)
			continue
		}
		encodedMiniblocksHashes = append(encodedMiniblocksHashes, hex.EncodeToString(miniblockHash))
	}

	return encodedMiniblocksHashes
}

func (mp *miniblocksProcessor) calculateHash(object interface{}) ([]byte, error) {
	return core.CalculateHash(mp.marshalier, mp.hasher, object)
}
