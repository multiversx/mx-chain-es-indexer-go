package disabled

import (
	"bytes"
	"context"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/require"
)

func TestDisabledElasticClient_MethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	ec := NewDisabledElasticClient()
	require.False(t, check.IfNil(ec))

	require.NotPanics(t, func() {
		_ = ec.DoBulkRequest(context.Background(), new(bytes.Buffer), "")
		_ = ec.DoQueryRemove(context.Background(), "", new(bytes.Buffer))
		_ = ec.DoMultiGet(context.Background(), make([]string, 0), "", true, nil)
		_ = ec.DoScrollRequest(context.Background(), "", []byte(""), true, nil)
		_, _ = ec.DoCountRequest(context.Background(), "", []byte(""))
		_ = ec.UpdateByQuery(context.Background(), "", new(bytes.Buffer))
		_ = ec.PutMappings("", new(bytes.Buffer))
		_ = ec.CheckAndCreateIndex("")
		_ = ec.CheckAndCreateAlias("", "")
		_ = ec.CheckAndCreateTemplate("", new(bytes.Buffer))
		_ = ec.CheckAndCreatePolicy("", new(bytes.Buffer))
	})
}
