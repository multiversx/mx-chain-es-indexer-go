package tags

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTagsCount_Serialize(t *testing.T) {
	t.Parallel()

	tagsC := NewTagsCount()

	tagsC.ParseTags([]string{"Art"})
	tagsC.ParseTags([]string{"Art"})

	buff, err := tagsC.Serialize()
	require.Nil(t, err)

	expected := `{ "update" : { "_id" : "Art", "_type" : "_doc" } }
{"script": {"source": "ctx._source.count += params.count","lang": "painless","params": {"count": 2}},"upsert": {"count": 2}}
`
	require.Equal(t, expected, buff[0].String())
}
