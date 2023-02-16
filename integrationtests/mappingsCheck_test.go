//go:build integrationtests

package integrationtests

import (
	"testing"

	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestMappingsOfESDTsIndex(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	_, err = CreateElasticProcessor(esClient)
	require.Nil(t, err)

	mappings, err := getIndexMappings(dataindexer.ESDTsIndex)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/mappings/esdts.json"), mappings)
}
