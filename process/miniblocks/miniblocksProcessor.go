package miniblocks

import (
	"encoding/hex"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("indexer/process/miniblocks")

type miniblocksProcessor struct {
	hasher       hashing.Hasher
	marshalier   marshal.Marshalizer
	selfShardID  uint32
	importDBMode bool
}

// NewMiniblocksProcessor will create a new instance of miniblocksProcessor
func NewMiniblocksProcessor(
	selfShardID uint32,
	hasher hashing.Hasher,
	marshalier marshal.Marshalizer,
	isImportDBMode bool,
) (*miniblocksProcessor, error) {
	if check.IfNil(marshalier) {
		return nil, indexer.ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, indexer.ErrNilHasher
	}

	return &miniblocksProcessor{
		hasher:       hasher,
		marshalier:   marshalier,
		selfShardID:  selfShardID,
		importDBMode: isImportDBMode,
	}, nil
}

// PrepareDBMiniblocks will prepare miniblocks from body
func (mp *miniblocksProcessor) PrepareDBMiniblocks(header coreData.HeaderHandler, body *block.Body) []*data.Miniblock {
	headerHash, err := mp.calculateHash(header)
	if err != nil {
		log.Warn("indexer: could not calculate header hash", "error", err)
		return nil
	}

	dbMiniblocks := make([]*data.Miniblock, 0)
	for mbIndex, miniblock := range body.MiniBlocks {
		dbMiniblock, errPrepareMiniblock := mp.prepareMiniblockForDB(mbIndex, miniblock, header, headerHash)
		if errPrepareMiniblock != nil {
			log.Warn("miniblocksProcessor.PrepareDBMiniblocks cannot prepare miniblock", "error", errPrepareMiniblock)
			continue
		}

		dbMiniblocks = append(dbMiniblocks, dbMiniblock)
	}

	return dbMiniblocks
}

func (mp *miniblocksProcessor) prepareMiniblockForDB(
	mbIndex int,
	miniblock *block.MiniBlock,
	header coreData.HeaderHandler,
	headerHash []byte,
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
	}

	processingType := mp.computeProcessingType(mbIndex, header)

	encodedHeaderHash := hex.EncodeToString(headerHash)
	if dbMiniblock.SenderShardID == header.GetShardID() {
		dbMiniblock.SenderBlockHash = encodedHeaderHash
		dbMiniblock.ProcessingTypeOnSource = processingType
	} else {
		dbMiniblock.ReceiverBlockHash = encodedHeaderHash
		dbMiniblock.ProcessingTypeOnDestination = processingType
	}

	if dbMiniblock.SenderShardID == dbMiniblock.ReceiverShardID {
		dbMiniblock.ReceiverBlockHash = encodedHeaderHash
		dbMiniblock.ProcessingTypeOnDestination = processingType
	}

	return dbMiniblock, nil
}

func (mp *miniblocksProcessor) computeProcessingType(mbIndex int, header coreData.HeaderHandler) string {
	miniblockHeaders := header.GetMiniBlockHeaderHandlers()
	if len(miniblockHeaders) < mbIndex+1 {
		return ""
	}

	currentMbHeader := miniblockHeaders[mbIndex]
	reserved := currentMbHeader.GetReserved()
	if len(reserved) > 0 && reserved[0] == byte(block.Scheduled) {
		return block.Scheduled.String()
	}

	return block.Normal.String()
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
		shouldIgnore := selfShardID == miniblock.SenderShardID && mp.importDBMode && isCrossShard
		if shouldIgnore {
			continue
		}

		isDstMe := selfShardID == miniblock.ReceiverShardID
		if isDstMe && isCrossShard {
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
