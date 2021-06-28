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
		"tags":      []string{"Art", "Art", "Sport", "Market"},
		"something": []string{"aaa"},
	}

	tagsC.ParseTagsFromAttributes(nil)
	tagsC.ParseTagsFromAttributes(attributes)
	tagsC.ParseTagsFromAttributes(attributes)
	tagsC.ParseTagsFromAttributes(attributes)

	require.Equal(t, 3, tagsC.Len())

	tagsS, ok := tagsC.(*tagsCount)
	require.True(t, ok)

	for _, value := range tagsS.tags {
		require.Equal(t, 3, value)
	}
}

func TestTagsCount_GetTags(t *testing.T) {
	t.Parallel()

	tagsC := NewTagsCount()

	attributes := &data.Attributes{
		"tags": []string{"Art", "Art", "Sport", "Market"},
	}

	tagsC.ParseTagsFromAttributes(attributes)

	tags := tagsC.GetTags()
	require.Len(t, tags, 3)
}
