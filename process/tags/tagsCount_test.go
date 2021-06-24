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

	tagsC.ExtractTagsFromAttributes(nil)
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

func TestTagsCount_GetTags(t *testing.T) {
	t.Parallel()

	tagsC := NewTagsCount()

	attributes := &data.Attributes{
		"tags": []string{"Art", "Art", "Sport", "Market"},
	}

	tagsC.ExtractTagsFromAttributes(attributes)

	tags := tagsC.GetTags()
	require.Len(t, tags, 3)
}

func TestTagsCount_ParseTagsFromDB(t *testing.T) {
	t.Parallel()

	tagsC := NewTagsCount()

	attributes := &data.Attributes{
		"tags": []string{"Art", "Art", "Sport", "Market"},
	}

	tagsC.ExtractTagsFromAttributes(attributes)

	response := map[string]interface{}{}
	err := tagsC.ParseTagsFromDB(response)
	require.Nil(t, err)

	err = tagsC.ParseTagsFromDB(nil)
	require.Nil(t, err)

	r1 := map[string]interface{}{
		"found": false,
	}
	r2 := map[string]interface{}{
		"found": true,
		"_id":   "Art",
		"_source": map[string]interface{}{
			"count": 10,
		},
	}
	r3 := map[string]interface{}{
		"found": true,
		"_id":   "Sport",
		"_source": map[string]interface{}{
			"count": 25,
		},
	}
	response = map[string]interface{}{
		"docs": []interface{}{
			r1, r2, r3,
		},
	}
	err = tagsC.ParseTagsFromDB(response)
	require.Nil(t, err)

	tagsS, ok := tagsC.(*tagsCount)
	require.True(t, ok)

	require.Equal(t, 11, tagsS.tags["Art"])
	require.Equal(t, 26, tagsS.tags["Sport"])
	require.Equal(t, 1, tagsS.tags["Market"])
}
