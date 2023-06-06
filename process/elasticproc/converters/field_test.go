package converters

import (
	"strings"
	"testing"

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
