//go:build integrationtests

package integrationtests

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	indexerdata "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestIndexLogSourceShardAndAfterDestinationAndAgainSource(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}
	body := &dataBlock.Body{}

	address1 := "erd1ju8pkvg57cwdmjsjx58jlmnuf4l9yspstrhr9tgsrt98n9edpm2qtlgy99"
	address2 := "erd1w7jyzuj6cv4ngw8luhlkakatjpmjh3ql95lmxphd3vssc4vpymks6k5th7"

	// INDEX ON SOURCE
	pool := &outport.Pool{
		Logs: []*coreData.LogData{
			{
				LogHandler: &transaction.Log{
					Address: decodeAddress(address1),
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address1),
							Identifier: []byte(core.BuiltInFunctionESDTTransfer),
							Topics:     [][]byte{[]byte("ESDT-abcd"), big.NewInt(0).Bytes(), big.NewInt(1).Bytes()},
						},
						nil,
					},
				},
				TxHash: "cross-log",
			},
		},
	}
	err = esProc.SaveTransactions(body, header, pool, map[string]*outport.AlteredAccount{}, false, testNumOfShards)
	require.Nil(t, err)

	ids := []string{hex.EncodeToString([]byte("cross-log"))}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.LogsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t,
		readExpectedResult("./testdata/logsCrossShard/log-at-source.json"),
		string(genericResponse.Docs[0].Source),
	)

	// INDEX ON DESTINATION
	header = &dataBlock.Header{
		Round:     50,
		TimeStamp: 6040,
	}
	pool = &outport.Pool{
		Logs: []*coreData.LogData{
			{
				LogHandler: &transaction.Log{
					Address: decodeAddress(address1),
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address1),
							Identifier: []byte(core.BuiltInFunctionESDTTransfer),
							Topics:     [][]byte{[]byte("ESDT-abcd"), big.NewInt(0).Bytes(), big.NewInt(1).Bytes()},
						},
						{

							Address:    decodeAddress(address2),
							Identifier: []byte("do-something"),
							Topics:     [][]byte{[]byte("topic1"), []byte("topic2")},
						},
						nil,
					},
				},
				TxHash: "cross-log",
			},
		},
	}
	err = esProc.SaveTransactions(body, header, pool, map[string]*outport.AlteredAccount{}, false, testNumOfShards)
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexerdata.LogsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t,
		readExpectedResult("./testdata/logsCrossShard/log-at-destination.json"),
		string(genericResponse.Docs[0].Source),
	)

	// INDEX ON SOURCE AGAIN SHOULD NOT CHANGE
	header = &dataBlock.Header{
		Round:     50,
		TimeStamp: 5000,
	}
	pool = &outport.Pool{
		Logs: []*coreData.LogData{
			{
				LogHandler: &transaction.Log{
					Address: decodeAddress(address1),
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address1),
							Identifier: []byte(core.BuiltInFunctionESDTTransfer),
							Topics:     [][]byte{[]byte("ESDT-abcd"), big.NewInt(0).Bytes(), big.NewInt(1).Bytes()},
						},
						nil,
					},
				},
				TxHash: "cross-log",
			},
		},
	}
	err = esProc.SaveTransactions(body, header, pool, map[string]*outport.AlteredAccount{}, false, testNumOfShards)
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexerdata.LogsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t,
		readExpectedResult("./testdata/logsCrossShard/log-at-destination.json"),
		string(genericResponse.Docs[0].Source),
	)
}
