package miniblocks

import (
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
)

// SerializeBulkMiniBlocks will serialize the provided miniblocks slice in a way that Elastic Search expects a bulk request
func (mp *miniblocksProcessor) SerializeBulkMiniBlocks(
	bulkMbs []*data.Miniblock,
	existsInDb map[string]bool,
	buffSlice *data.BufferSlice,
	index string,
) {
	for _, mb := range bulkMbs {
		meta, serializedData, err := mp.prepareMiniblockData(mb, existsInDb[mb.Hash], index)
		if err != nil {
			log.Warn("miniblocksProcessor.prepareMiniblockData cannot prepare miniblock data", "error", err)
			continue
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			log.Warn("miniblocksProcessor.putInBufferMiniblockData cannot prepare miniblock data", "error", err)
			continue
		}
	}
}

func (mp *miniblocksProcessor) prepareMiniblockData(miniblockDB *data.Miniblock, isInDB bool, index string) ([]byte, []byte, error) {
	if !isInDB {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_index":"%s", "_id" : "%s"} }%s`, index, miniblockDB.Hash, "\n"))
		serializedData, err := json.Marshal(miniblockDB)

		return meta, serializedData, err
	}

	// prepare data for update operation
	meta := []byte(fmt.Sprintf(`{ "update" : {"_index":"%s", "_id" : "%s" } }%s`, index, miniblockDB.Hash, "\n"))
	if mp.selfShardID == miniblockDB.SenderShardID && miniblockDB.ProcessingTypeOnDestination != block.Processed.String() {
		// prepare for update sender block hash
		serializedData := []byte(fmt.Sprintf(`{ "doc" : { "senderBlockHash" : "%s", "procTypeS": "%s" } }`, miniblockDB.SenderBlockHash, miniblockDB.ProcessingTypeOnSource))

		return meta, serializedData, nil
	}

	// prepare for update receiver block hash
	serializedData := []byte(fmt.Sprintf(`{ "doc" : { "receiverBlockHash" : "%s", "procTypeD": "%s" } }`, miniblockDB.ReceiverBlockHash, miniblockDB.ProcessingTypeOnDestination))

	return meta, serializedData, nil
}
