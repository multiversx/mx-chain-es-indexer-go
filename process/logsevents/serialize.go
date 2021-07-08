package logsevents

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

// SerializeLogs will serialize the provided logs in a way that Elastic Search expects a bulk request
func (logsAndEventsProcessor) SerializeLogs(logs []*data.Logs) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()
	for _, log := range logs {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, log.ID, "\n"))
		serializedData, errMarshal := json.Marshal(log)
		if errMarshal != nil {
			return nil, errMarshal
		}

		err := buffSlice.PutData(meta, serializedData)
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}
