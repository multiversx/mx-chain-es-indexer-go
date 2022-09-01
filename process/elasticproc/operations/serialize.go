package operations

import (
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/process/elasticproc/converters"
)

// SerializeSCRs will serialize smart contract results
func (op *operationsProcessor) SerializeSCRs(scrs []*data.ScResult, buffSlice *data.BufferSlice, index string) error {
	for _, scr := range scrs {
		meta, serializedData, err := op.prepareSerializedDataForAScResult(scr, index)
		if err != nil {
			return err
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			return err
		}
	}

	return nil
}

func (op *operationsProcessor) prepareSerializedDataForAScResult(
	scr *data.ScResult,
	index string,
) ([]byte, []byte, error) {
	metaData := []byte(fmt.Sprintf(`{"update":{"_index":"%s","_id":"%s"}}%s`, index, converters.JsonEscape(scr.Hash), "\n"))
	marshaledSCR, err := json.Marshal(scr)
	if err != nil {
		return nil, nil, err
	}

	selfShardID := op.shardCoordinator.SelfId()
	isCrossShardOnSourceShard := scr.SenderShard != scr.ReceiverShard && scr.SenderShard == selfShardID
	if isCrossShardOnSourceShard {
		serializedData :=
			[]byte(fmt.Sprintf(`{"script":{"source":"return"},"upsert":%s}`,
				string(marshaledSCR)))

		return metaData, serializedData, nil
	}

	meta := []byte(fmt.Sprintf(`{ "index" : { "_index":"%s","_id" : "%s" } }%s`, index, converters.JsonEscape(scr.Hash), "\n"))

	return meta, marshaledSCR, nil
}
