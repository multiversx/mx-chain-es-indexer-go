package miniblocks

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
)

// SerializeBulkMiniBlocks will serialize the provided miniblocks slice in a way that Elastic Search expects a bulk request
func (mp *miniblocksProcessor) SerializeBulkMiniBlocks(
	bulkMbs []*data.Miniblock,
	existsInDb map[string]bool,
) *bytes.Buffer {
	buff := &bytes.Buffer{}
	for _, mb := range bulkMbs {
		meta, serializedData, err := mp.prepareMiniblockData(mb, existsInDb[mb.Hash])
		if err != nil {
			log.Warn("miniblocksProcessor.prepareMiniblockData cannot prepare miniblock data", "error", err)
			continue
		}

		err = putInBufferMiniblockData(buff, meta, serializedData)
		if err != nil {
			log.Warn("miniblocksProcessor.putInBufferMiniblockData cannot prepare miniblock data", "error", err)
			continue
		}
	}

	return buff
}

func (mp *miniblocksProcessor) prepareMiniblockData(miniblockDB *data.Miniblock, isInDB bool) ([]byte, []byte, error) {
	if !isInDB {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s", "_type" : "%s" } }%s`, miniblockDB.Hash, "_doc", "\n"))
		serializedData, err := json.Marshal(miniblockDB)

		return meta, serializedData, err
	}

	// prepare data for update operation
	meta := []byte(fmt.Sprintf(`{ "update" : { "_id" : "%s" } }%s`, miniblockDB.Hash, "\n"))
	if mp.selfShardID == miniblockDB.SenderShardID && miniblockDB.ProcessingTypeOnDestination != block.Processed.String() {
		// prepare for update sender block hash
		serializedData := []byte(fmt.Sprintf(`{ "doc" : { "senderBlockHash" : "%s", "procTypeS": "%s" } }`, miniblockDB.SenderBlockHash, miniblockDB.ProcessingTypeOnSource))

		return meta, serializedData, nil
	}

	// prepare for update receiver block hash
	serializedData := []byte(fmt.Sprintf(`{ "doc" : { "receiverBlockHash" : "%s", "procTypeD": "%s" } }`, miniblockDB.ReceiverBlockHash, miniblockDB.ProcessingTypeOnDestination))

	return meta, serializedData, nil
}

func putInBufferMiniblockData(buff *bytes.Buffer, meta, serializedData []byte) error {
	serializedData = append(serializedData, "\n"...)
	buff.Grow(len(meta) + len(serializedData))
	_, err := buff.Write(meta)
	if err != nil {
		return err
	}

	_, err = buff.Write(serializedData)
	if err != nil {
		return err
	}

	return nil
}
