package logsevents

import (
	"bytes"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

func (logsAndEventsProcessor) SerializeLogs(logs *data.Logs) ([]*bytes.Buffer, error) {
	return nil, nil
}
