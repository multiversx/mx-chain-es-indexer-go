package converters

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJsonEscape(t *testing.T) {
	t.Parallel()

	require.Equal(t, "hello", JsonEscape("hello"))
	require.Equal(t, "\\\\", JsonEscape("\\"))
	require.Equal(t, `\"'\\/.,\u003c\u003e'\"`, JsonEscape(`"'\/.,<>'"`))
	require.Equal(t, `tag\u003e`, JsonEscape(`tag>`))
	require.Equal(t, ",.\\u003c.\\u003e\\u003c\\u003c\\u003c\\u003c\\u003e\\u003e\\u003e\\u003e\\u003e", JsonEscape(",.<.><<<<>>>>>"))
}
