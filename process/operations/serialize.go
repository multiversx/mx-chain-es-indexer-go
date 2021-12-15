package operations

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

// SerializeSCRS will serialize smart contract results
func (op *operationsProcessor) SerializeSCRs(scrs []*data.ScResult) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()

	for _, scr := range scrs {
		meta, serializedData, err := op.prepareSerializedDataForAScResult(scr)
		if err != nil {
			return nil, err
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}

func (op *operationsProcessor) prepareSerializedDataForAScResult(
	scr *data.ScResult,
) ([]byte, []byte, error) {
	metaData := []byte(fmt.Sprintf(`{"update":{"_id":"%s", "_type": "_doc"}}%s`, scr.Hash, "\n"))
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

	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s", "_type" : "%s" } }%s`, scr.Hash, "_doc", "\n"))

	return meta, marshaledSCR, nil
}
