package converters

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrepareTagsShouldWork(t *testing.T) {
	t.Parallel()

	attributes := []byte("tags:test,free,fun;description:This is a test description for an awesome nft")
	prepared := ExtractTagsFromAttributes(attributes)
	require.Equal(t, []string{"test", "free", "fun"}, prepared)

	attributes = []byte("tags:TEST,free,fun;description: ")
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

	hexEncodedAttributes := "746167733a5c5c5c5c5c5c5c5c5c5c5c5c5c5c5c5c5c5c5c2c3c3c3c3e3e3e2626262626262626262626262626262c272727273b6d657461646174613a516d533757525566464464516458654c513637516942394a33663746654d69343554526d6f79415741563568345a"
	attributes, _ = hex.DecodeString(hexEncodedAttributes)
	prepared = ExtractTagsFromAttributes(attributes)
	require.Equal(t, []string{"\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\", "<<<>>>&&&&&&&&&&&&&&&", "''''"}, prepared)
}

func TestExtractMetadataFromAttributesShouldWork(t *testing.T) {
	t.Parallel()

	attributes := []byte("tags:,,,,,,;metadata:something")
	prepared := ExtractMetaDataFromAttributes(attributes)
	require.Equal(t, "something", prepared)

	attributes = []byte("tags:,,,,,,;metadata:SOMETHING")
	prepared = ExtractMetaDataFromAttributes(attributes)
	require.Equal(t, "SOMETHING", prepared)

	attributes = []byte("tags:,,,,,,;metadate:SOMETHING")
	prepared = ExtractMetaDataFromAttributes(attributes)
	require.Equal(t, "", prepared)
}
