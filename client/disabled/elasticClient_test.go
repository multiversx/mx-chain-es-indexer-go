package disabled

import (
	"bytes"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/require"
)

func TestDisabledElasticClient_MethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	ec := NewDisabledElasticClient()
	require.False(t, check.IfNil(ec))

	require.NotPanics(t, func() {
		_ = ec.DoBulkRequest(nil, new(bytes.Buffer), "")
		_ = ec.DoQueryRemove(nil, "", new(bytes.Buffer))
		_ = ec.DoMultiGet(nil, make([]string, 0), "", true, nil)
		_ = ec.DoScrollRequest(nil, "", []byte(""), true, nil)
		_, _ = ec.DoCountRequest(nil, "", []byte(""))
		_ = ec.UpdateByQuery(nil, "", new(bytes.Buffer))
		_ = ec.PutMappings("", new(bytes.Buffer))
		_ = ec.CheckAndCreateIndex("")
		_ = ec.CheckAndCreateAlias("", "")
		_ = ec.CheckAndCreateTemplate("", new(bytes.Buffer))
		_ = ec.CheckAndCreatePolicy("", new(bytes.Buffer))
	})
}
