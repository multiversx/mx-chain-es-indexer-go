package tags

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/stretchr/testify/require"
)

func TestTagsCount_Serialize(t *testing.T) {
	t.Parallel()

	tagsC := NewTagsCount()

	attributes := &data.Attributes{
		"tags": []string{"Art"},
	}

	tagsC.ParseTagsFromAttributes(attributes)
	tagsC.ParseTagsFromAttributes(attributes)

	buff, err := tagsC.Serialize()
	require.Nil(t, err)

	expected := `{ "index" : { "_id" : "Art" } }
{ "count" : 2 }
`
	require.Equal(t, expected, buff[0].String())
}
