package tags

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
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

func TestTagsCount_TruncateID(t *testing.T) {
	t.Parallel()

	tagsC := NewTagsCount()

	randomBytes := make([]byte, 600)
	_, _ = rand.Read(randomBytes)

	tagsC.ParseTags([]string{string(randomBytes)})

	buffSlice := data.NewBufferSlice(data.DefaultMaxBulkSize)
	err := tagsC.Serialize(buffSlice, "tags")
	require.Nil(t, err)

	expected := fmt.Sprintf(`{ "update" : {"_index":"tags", "_id" : "%s" } }
{"script": {"source": "ctx._source.count += params.count; ctx._source.tag = params.tag","lang": "painless","params": {"count": 1, "tag": "%s"}},"upsert": {"count": 1, "tag":"%s"}}
`, base64.StdEncoding.EncodeToString(randomBytes)[:converters.MaxIDSize], converters.JsonEscape(string(randomBytes)), converters.JsonEscape(string(randomBytes)))
	require.Equal(t, expected, buffSlice.Buffers()[0].String())
}
