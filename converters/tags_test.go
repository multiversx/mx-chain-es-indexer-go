package converters

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrepareTagsShouldWork(t *testing.T) {
	t.Parallel()

	attributes := []byte("tags:test,free,fun;description:This is a test description for an awesome nft")
	prepared := ExtractTagsFromAttributes(attributes)
	require.Equal(t, []string{"test", "free", "fun"}, prepared)

	attributes = []byte("tags:test,free,fun;description: ")
	prepared = ExtractTagsFromAttributes(attributes)
	require.Equal(t, []string{"test", "free", "fun"}, prepared)

	attributes = []byte("description:;")
	prepared = ExtractTagsFromAttributes(attributes)
	require.Nil(t, prepared)

	attributes = []byte("")
	prepared = ExtractTagsFromAttributes(attributes)
	require.Nil(t, prepared)

	attributes = []byte(";")
	prepared = ExtractTagsFromAttributes(attributes)
	require.Nil(t, prepared)

	attributes = []byte("  ")
	prepared = ExtractTagsFromAttributes(attributes)
	require.Nil(t, prepared)

	attributes = []byte("attribute")
	prepared = ExtractTagsFromAttributes(attributes)
	require.Nil(t, prepared)

	attributes = []byte("shard:1,3")
	prepared = ExtractTagsFromAttributes(attributes)
	require.Nil(t, prepared)

	attributes = []byte("tags:;metadata:")
	prepared = ExtractTagsFromAttributes(attributes)
	require.Nil(t, prepared)

	attributes = []byte("tags:,,,,,,;metadata:")
	prepared = ExtractTagsFromAttributes(attributes)
	require.Nil(t, prepared)

	attributes = []byte("tags")
	prepared = ExtractTagsFromAttributes(attributes)
	require.Nil(t, prepared)
}
