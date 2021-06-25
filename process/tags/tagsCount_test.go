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

func TestTagsCount_ParseTagsFromDB(t *testing.T) {
	t.Parallel()

	tagsC := NewTagsCount()

	attributes := &data.Attributes{
		"tags": []string{"Art", "Art", "Sport", "Market"},
	}

	tagsC.ParseTagsFromAttributes(attributes)

	response := &data.ResponseTags{}
	tagsC.ParseTagsFromDB(response)

	tagsC.ParseTagsFromDB(nil)

	response = &data.ResponseTags{
		Docs: []data.ResponseTagDB{
			{
				Found: false,
			},
			{
				Found:   true,
				TagName: "Art",
				Source: data.SourceTag{
					Count: 10,
				},
			},
			{
				Found:   true,
				TagName: "Sport",
				Source: data.SourceTag{
					Count: 25,
				},
			},
		},
	}
	tagsC.ParseTagsFromDB(response)

	tagsS, ok := tagsC.(*tagsCount)
	require.True(t, ok)

	require.Equal(t, 11, tagsS.tags["Art"])
	require.Equal(t, 26, tagsS.tags["Sport"])
	require.Equal(t, 1, tagsS.tags["Market"])
}
