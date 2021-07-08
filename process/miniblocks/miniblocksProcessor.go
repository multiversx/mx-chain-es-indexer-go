package miniblocks

import (
	"encoding/hex"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

var log = logger.GetOrCreate("indexer/process/miniblocks")

type miniblocksProcessor struct {
	hasher      hashing.Hasher
	marshalier  marshal.Marshalizer
	selfShardID uint32
}

// NewMiniblocksProcessor will create a new instance of miniblocksProcessor
func NewMiniblocksProcessor(
	selfShardID uint32,
	hasher hashing.Hasher,
	marshalier marshal.Marshalizer,
) (*miniblocksProcessor, error) {
	if check.IfNil(marshalier) {
		return nil, indexer.ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, indexer.ErrNilHasher
	}

	return &miniblocksProcessor{
		hasher:      hasher,
		marshalier:  marshalier,
		selfShardID: selfShardID,
	}, nil
}

// PrepareDBMiniblocks will prepare miniblocks from body
func (mp *miniblocksProcessor) PrepareDBMiniblocks(header nodeData.HeaderHandler, body *block.Body) []*data.Miniblock {
	headerHash, err := mp.calculateHash(header)
	if err != nil {
		log.Warn("indexer: could not calculate header hash", "error", err)
		return nil
	}

	dbMiniblocks := make([]*data.Miniblock, 0)
	for _, miniblock := range body.MiniBlocks {
		dbMiniblock, errPrepareMiniblock := mp.prepareMiniblockForDB(miniblock, header, headerHash)
		if errPrepareMiniblock != nil {
			log.Warn("miniblocksProcessor.PrepareDBMiniblocks cannot prepare miniblock", "error", errPrepareMiniblock)
			continue
		}

		dbMiniblocks = append(dbMiniblocks, dbMiniblock)
	}

	return dbMiniblocks
}

func (mp *miniblocksProcessor) prepareMiniblockForDB(
	miniblock *block.MiniBlock,
	header nodeData.HeaderHandler,
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

	encodedHeaderHash := hex.EncodeToString(headerHash)
	if dbMiniblock.SenderShardID == header.GetShardID() {
		dbMiniblock.SenderBlockHash = encodedHeaderHash
	} else {
		dbMiniblock.ReceiverBlockHash = encodedHeaderHash
	}

	if dbMiniblock.SenderShardID == dbMiniblock.ReceiverShardID {
		dbMiniblock.ReceiverBlockHash = encodedHeaderHash
	}

	return dbMiniblock, nil
}

// GetMiniblocksHashesHexEncoded will compute miniblocks hashes in a hexadecimal encoding
func (mp *miniblocksProcessor) GetMiniblocksHashesHexEncoded(header nodeData.HeaderHandler, body *block.Body) []string {
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
