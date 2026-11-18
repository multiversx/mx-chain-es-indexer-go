package converters

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerateBaseUUID(t *testing.T) {
	t.Parallel()

	uuid := GenerateBase64UUID()
	require.NotEmpty(t, uuid)
	require.Len(t, uuid, 24)
}
