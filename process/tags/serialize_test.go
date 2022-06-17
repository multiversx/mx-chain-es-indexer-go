package tags

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTagsCount_Serialize(t *testing.T) {
	t.Parallel()

	tagsC := NewTagsCount()

	tagsC.ParseTags([]string{"Art"})
	tagsC.ParseTags([]string{"Art"})

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	err := tagsC.Serialize(buffSlice, "tags")
	require.Nil(t, err)

	expected := `{ "update" : {"_index":"tags", "_id" : "QXJ0" } }
{"script": {"source": "ctx._source.count += params.count; ctx._source.tag = params.tag","lang": "painless","params": {"count": 2, "tag": "Art"}},"upsert": {"count": 2, "tag":"Art"}}
`
	require.Equal(t, expected, buffSlice.Buffers()[0].String())
}
