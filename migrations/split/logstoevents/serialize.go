package logstoevents

import (
	"encoding/json"
	"fmt"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
)

func serializeLogEvent(logEvent *data.LogEvent, buffSlice *data.BufferSlice) error {
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, logEvent.ID, "\n"))
	serializedData, errMarshal := json.Marshal(logEvent)
	if errMarshal != nil {
		return errMarshal
	}

	return buffSlice.PutData(meta, serializedData)
}
