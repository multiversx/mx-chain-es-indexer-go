package tags

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTagsCount_ExtractTagsFromAttributes(t *testing.T) {
	t.Parallel()

	tagsC := NewTagsCount()

	tagsC.ParseTags(nil)
	tagsC.ParseTags([]string{"Art", "Art", "Sport", "Market"})
	tagsC.ParseTags([]string{"Art", "Art", "Sport", "Market"})
	tagsC.ParseTags([]string{"Art", "Art", "Sport", "Market"})

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

	tagsC.ParseTags([]string{"Art", "Art", "Sport", "Market"})

	tags := tagsC.GetTags()
	require.Len(t, tags, 3)
}
