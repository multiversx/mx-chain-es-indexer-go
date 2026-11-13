package converters

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/stretchr/testify/require"
)

func TestTruncateFieldIfExceedsMaxLength(t *testing.T) {
	t.Parallel()

	field := "my-field"
	truncated := TruncateFieldIfExceedsMaxLength(field)
	require.Equal(t, field, truncated)

	maxLengthField := strings.Repeat("a", data.MaxFieldLength) + strings.Repeat("b", data.MaxFieldLength)
	truncated = TruncateFieldIfExceedsMaxLength(maxLengthField)
	require.Equal(t, strings.Repeat("a", data.MaxFieldLength), truncated)
}

func TestTruncateSliceElementsIfExceedsMaxLength(t *testing.T) {
	t.Parallel()

	fields := []string{"my-field", "my-field"}
	truncated := TruncateSliceElementsIfExceedsMaxLength(fields)
	require.Equal(t, fields, truncated)

	bigField := strings.Repeat("a", data.MaxFieldLength) + strings.Repeat("b", data.MaxFieldLength)
	fields = []string{bigField, bigField}
	truncated = TruncateSliceElementsIfExceedsMaxLength(fields)
	require.Equal(t, []string{strings.Repeat("a", data.MaxFieldLength), strings.Repeat("a", data.MaxFieldLength)}, truncated)
}

func TestTruncateFieldIfExceedsMaxLength_InvalidUTF8(t *testing.T) {
	t.Parallel()

	// every invalid byte is indexed as the replacement character, which has 3 bytes:
	// the length should be checked on the value that will be indexed
	invalidField := strings.Repeat("\xff", 11000)
	truncated := TruncateFieldIfExceedsMaxLength(invalidField)
	require.True(t, utf8.ValidString(truncated))
	require.Equal(t, strings.Repeat("\uFFFD", data.MaxFieldLength/3), truncated)
	require.Equal(t, data.MaxFieldLength, lenAfterJSONEncoding(t, truncated))

	// without the truncation the indexed value would have 33000 bytes
	require.Equal(t, 33000, lenAfterJSONEncoding(t, invalidField))
}

func TestTruncateFieldIfExceedsMaxLength_ShortInvalidUTF8IsIndexedTheSame(t *testing.T) {
	t.Parallel()

	field := "abc\xff\xfedef"
	truncated := TruncateFieldIfExceedsMaxLength(field)
	require.Equal(t, "abc\uFFFD\uFFFDdef", truncated)

	// the JSON encoder was already doing the same replacement, so the indexed value does not change
	var fromField, fromTruncated string
	fieldBytes, _ := json.Marshal(field)
	truncatedBytes, _ := json.Marshal(truncated)
	require.Nil(t, json.Unmarshal(fieldBytes, &fromField))
	require.Nil(t, json.Unmarshal(truncatedBytes, &fromTruncated))
	require.Equal(t, fromField, fromTruncated)
}

func TestTruncateFieldIfExceedsMaxLength_CutsBetweenCharacters(t *testing.T) {
	t.Parallel()

	// 2 bytes character
	truncated := TruncateFieldIfExceedsMaxLength(strings.Repeat("é", data.MaxFieldLength))
	require.Equal(t, strings.Repeat("é", data.MaxFieldLength/2), truncated)

	// 4 bytes character after one byte: the last character does not fit and is not cut in the middle
	truncated = TruncateFieldIfExceedsMaxLength("a" + strings.Repeat("😀", data.MaxFieldLength))
	require.True(t, utf8.ValidString(truncated))
	require.Equal(t, "a"+strings.Repeat("😀", (data.MaxFieldLength-1)/4), truncated)
	require.LessOrEqual(t, len(truncated), data.MaxFieldLength)

	// a valid field that has exactly the max length is not changed
	maxLengthField := strings.Repeat("é", data.MaxFieldLength/2)
	require.Equal(t, maxLengthField, TruncateFieldIfExceedsMaxLength(maxLengthField))
}

func lenAfterJSONEncoding(t *testing.T, field string) int {
	fieldBytes, err := json.Marshal(field)
	require.Nil(t, err)

	var decoded string
	err = json.Unmarshal(fieldBytes, &decoded)
	require.Nil(t, err)

	return len(decoded)
}
