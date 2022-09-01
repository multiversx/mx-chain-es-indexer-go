//go:build integrationtests

package integrationtests

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	indexerdata "github.com/ElrondNetwork/elastic-indexer-go/process/dataindexer"
	"github.com/ElrondNetwork/elrond-go-core/core"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestIndexLogSourceShardAndAfterDestinationAndAgainSource(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, shardCoordinator)
	require.Nil(t, err)

	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}
	body := &dataBlock.Body{}

	// INDEX ON SOURCE
	pool := &outport.Pool{
		Logs: []*coreData.LogData{
			{
				LogHandler: &transaction.Log{
					Address: []byte("addr-1"),
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
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
	err = esProc.SaveTransactions(body, header, pool, map[string]*outport.AlteredAccount{})
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
					Address: []byte("addr-1"),
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionESDTTransfer),
							Topics:     [][]byte{[]byte("ESDT-abcd"), big.NewInt(0).Bytes(), big.NewInt(1).Bytes()},
						},
						{

							Address:    []byte("addr-3"),
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
	err = esProc.SaveTransactions(body, header, pool, map[string]*outport.AlteredAccount{})
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
					Address: []byte("addr-1"),
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
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
	err = esProc.SaveTransactions(body, header, pool, map[string]*outport.AlteredAccount{})
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexerdata.LogsIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t,
		readExpectedResult("./testdata/logsCrossShard/log-at-destination.json"),
		string(genericResponse.Docs[0].Source),
	)
}
