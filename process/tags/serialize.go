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
		if tag == "" {
			continue
		}

		meta := []byte(fmt.Sprintf(`{ "update" : { "_id" : "%s", "_type" : "_doc" } }%s`, tag, "\n"))
		serializedDataStr := fmt.Sprintf(`{"script": {"source": "ctx._source.count += params.count","lang": "painless","params": {"count": %d}},"upsert": {"count": %d}}`, count, count)

		err := buffSlice.PutData(meta, []byte(serializedDataStr))
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}
