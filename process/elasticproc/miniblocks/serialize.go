package miniblocks

import (
	"encoding/json"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
)

// SerializeBulkMiniBlocks will serialize the provided miniblocks slice in a way that Elasticsearch expects a bulk request
func (mp *miniblocksProcessor) SerializeBulkMiniBlocks(
	bulkMbs []*data.Miniblock,
	buffSlice *data.BufferSlice,
	index string,
	shardID uint32,
) {
	for _, mb := range bulkMbs {
		meta, serializedData, err := mp.prepareMiniblockData(mb, index, shardID)
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

func (mp *miniblocksProcessor) prepareMiniblockData(miniblockDB *data.Miniblock, index string, shardID uint32) ([]byte, []byte, error) {
	mbHash := miniblockDB.Hash
	miniblockDB.Hash = ""

	mbBytes, errMarshal := json.Marshal(miniblockDB)
	if errMarshal != nil {
		return nil, nil, errMarshal
	}

	// prepare data for update operation
	meta := []byte(fmt.Sprintf(`{ "update" : {"_index":"%s", "_id" : "%s" } }%s`, index, converters.JsonEscape(mbHash), "\n"))

	onSourceNotProcessed := shardID == miniblockDB.SenderShardID && miniblockDB.ProcessingTypeOnDestination != block.Processed.String()
	codeToExecute := `
	if ('create' == ctx.op) {
			ctx._source = params.mb
	} else {
		if (params.osnp) {
			ctx._source.senderBlockHash = params.mb.senderBlockHash;
			ctx._source.procTypeS = params.mb.procTypeS;
		} else {
			ctx._source.receiverBlockHash = params.mb.receiverBlockHash;
			ctx._source.procTypeD = params.mb.procTypeD;
		}
	}
`

	serializedDataStr := fmt.Sprintf(`{"scripted_upsert": true, "script": {`+
		`"source": "%s",`+
		`"lang": "painless",`+
		`"params": { "mb": %s, "osnp": %t }},`+
		`"upsert": {}}`,
		converters.FormatPainlessSource(codeToExecute), mbBytes, onSourceNotProcessed,
	)

	return meta, []byte(serializedDataStr), nil
}
