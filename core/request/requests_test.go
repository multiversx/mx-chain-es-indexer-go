package request

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func TestExtendTopicWithShardID(t *testing.T) {
	t.Parallel()

	require.Equal(t, "req_update_2", ExtendTopicWithShardID(UpdateTopic, 2))
	require.Equal(t, "req_bulk_0", ExtendTopicWithShardID(BulkTopic, 0))
	require.Equal(t, "req_get_1", ExtendTopicWithShardID(GetTopic, 1))
	require.Equal(t, "req_scroll_4294967295", ExtendTopicWithShardID(ScrollTopic, core.MetachainShardId))
}

func TestSplitTopicAndShardID(t *testing.T) {

	topic, shardID := SplitTopicAndShardID("req_update_2")
	require.Equal(t, UpdateTopic, topic)
	require.Equal(t, "2", shardID)

	topic, shardID = SplitTopicAndShardID("req_scroll_4294967295")
	require.Equal(t, ScrollTopic, topic)
	require.Equal(t, fmt.Sprintf("%d", core.MetachainShardId), shardID)

	topic, shardID = SplitTopicAndShardID("req")
	require.Equal(t, "req", topic)
	require.Equal(t, noShardID, shardID)

	topic, shardID = SplitTopicAndShardID("req_aaaa")
	require.Equal(t, "req_aaaa", topic)
	require.Equal(t, noShardID, shardID)
}
