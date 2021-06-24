package tags

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/stretchr/testify/require"
)

func TestTagsCount_ExtractTagsFromAttributes(t *testing.T) {
	t.Parallel()

	tagsC := NewTagsCount()

	attributes := &data.Attributes{
		"tags": []string{"Art", "Art", "Sport", "Market"},
	}

	tagsC.ExtractTagsFromAttributes(attributes)
	tagsC.ExtractTagsFromAttributes(attributes)
	tagsC.ExtractTagsFromAttributes(attributes)

	require.Equal(t, 3, tagsC.Len())

	tagsS, ok := tagsC.(*tagsCount)
	require.True(t, ok)

	for _, value := range tagsS.tags {
		require.Equal(t, 3, value)
	}
}
