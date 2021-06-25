package tags

import (
	"bytes"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

// Serialize will serialize tagsCount in a way that Elastic Search expects a bulk request
func (tc *tagsCount) Serialize() ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()
	for tag, count := range tc.tags {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, tag, "\n"))
		serializedData := []byte(fmt.Sprintf(`{ "count" : %d }`, count))

		err := buffSlice.PutData(meta, serializedData)
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}
