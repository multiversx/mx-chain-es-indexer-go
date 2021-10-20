package prometheus

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreatePrometheusHandler(t *testing.T) {
	promHandler, err := CreatePrometheusHandler(false, "")
	require.Nil(t, err)

	_, ok := promHandler.(*prometheusDisabled)
	require.True(t, ok)

	promHandler, err = CreatePrometheusHandler(true, ":2132")
	require.Nil(t, err)

	_, ok = promHandler.(*prometheusClient)
	require.True(t, ok)
}
