package block

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/errors"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go-logger/check"
	"github.com/ElrondNetwork/elrond-go/core"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

var log = logger.GetOrCreate("indexer/process/block")

type blockProcessor struct {
	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
}

// NewBlockProcessor will create a new instance of block processor
func NewBlockProcessor(hasher hashing.Hasher, marshalizer marshal.Marshalizer) (*blockProcessor, error) {
	if check.IfNil(hasher) {
		return nil, errors.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, errors.ErrNilMarshalizer
	}

	return &blockProcessor{
		hasher:      hasher,
		marshalizer: marshalizer,
	}, nil
}

// PrepareBlockForDB will prepare a database block and serialize it for database
func (bp *blockProcessor) PrepareBlockForDB(
	header nodeData.HeaderHandler,
	signersIndexes []uint64,
	body *block.Body,
	notarizedHeadersHashes []string,
	sizeTxs int,
) (*data.Block, error) {
	if check.IfNil(header) {
		return nil, errors.ErrNilHeaderHandler
	}
	if body == nil {
		return nil, errors.ErrNilBlockBody
	}

	blockSizeInBytes, headerHash, err := bp.computeBlockSizeAndHeaderHash(header, body)
	if err != nil {
		return nil, err
	}

	miniblocksHashes := bp.getEncodedMBSHashes(body)
	leaderIndex := bp.getLeaderIndex(signersIndexes)

	elasticBlock := &data.Block{
		Nonce:                 header.GetNonce(),
		Round:                 header.GetRound(),
		Epoch:                 header.GetEpoch(),
		ShardID:               header.GetShardID(),
		Hash:                  hex.EncodeToString(headerHash),
		MiniBlocksHashes:      miniblocksHashes,
		NotarizedBlocksHashes: notarizedHeadersHashes,
		Proposer:              leaderIndex,
		Validators:            signersIndexes,
		PubKeyBitmap:          hex.EncodeToString(header.GetPubKeysBitmap()),
		Size:                  int64(blockSizeInBytes),
		SizeTxs:               int64(sizeTxs),
		Timestamp:             time.Duration(header.GetTimeStamp()),
		TxCount:               header.GetTxCount(),
		StateRootHash:         hex.EncodeToString(header.GetRootHash()),
		PrevHash:              hex.EncodeToString(header.GetPrevHash()),
		SearchOrder:           computeBlockSearchOrder(header),
		AccumulatedFees:       header.GetAccumulatedFees().String(),
		DeveloperFees:         header.GetDeveloperFees().String(),
		EpochStartBlock:       header.IsStartOfEpochBlock(),
	}

	return elasticBlock, nil
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

func (bp *blockProcessor) computeBlockSizeAndHeaderHash(header nodeData.HeaderHandler, body *block.Body) (int, []byte, error) {
	headerBytes, err := bp.marshalizer.Marshal(header)
	if err != nil {
		return 0, nil, err
	}
	bodyBytes, err := bp.marshalizer.Marshal(body)
	if err != nil {
		return 0, nil, err
	}

	blockSize := len(headerBytes) + len(bodyBytes)

	headerHash := bp.hasher.Compute(string(headerBytes))

	return blockSize, headerHash, nil
}

func (bp *blockProcessor) getLeaderIndex(signersIndexes []uint64) uint64 {
	if len(signersIndexes) > 0 {
		return signersIndexes[0]
	}

	return 0
}

func computeBlockSearchOrder(header nodeData.HeaderHandler) uint64 {
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
func (bp *blockProcessor) ComputeHeaderHash(header nodeData.HeaderHandler) ([]byte, error) {
	return core.CalculateHash(bp.marshalizer, bp.hasher, header)
}
