//go:build integrationtests

package integrationtests

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	indexerdata "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestNFTUpdateMetadata(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	bigUri := bytes.Repeat([]byte("a"), 50000)
	esdtCreateData := &esdt.ESDigitalToken{
		TokenMetaData: &esdt.MetaData{
			URIs: [][]byte{[]byte("uri"), []byte("uri"), bigUri, bigUri, bigUri},
		},
	}
	marshalizedCreate, _ := json.Marshal(esdtCreateData)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
		ShardID:   1,
	}
	body := &dataBlock.Body{}

	// CREATE NFT data
	address := "erd1w7jyzuj6cv4ngw8luhlkakatjpmjh3ql95lmxphd3vssc4vpymks6k5th7"
	pool := &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(1).Bytes(), marshalizedCreate},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids := []string{"NFT-abcd-0e"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token.json"), string(genericResponse.Docs[0].Source))

	// Add URIS 1
	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionESDTNFTAddURI),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("uri"), bigUri},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	// Add URIS 2 --- results should be the same
	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionESDTNFTAddURI),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("uri"), bigUri},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	// Update attributes 1
	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-add-uris.json"), string(genericResponse.Docs[0].Source))

	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionESDTNFTUpdateAttributes),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("tags:test,free,fun;description:This is a test description for an awesome nft;metadata:metadata-test")},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-update-attributes.json"), string(genericResponse.Docs[0].Source))

	// Update attributes 2

	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionESDTNFTUpdateAttributes),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("something")},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-update-attributes-second.json"), string(genericResponse.Docs[0].Source))

	// Freeze nft
	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionESDTFreeze),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("something")},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)
	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-freeze.json"), string(genericResponse.Docs[0].Source))

	// UnFreeze nft
	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionESDTUnFreeze),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("something")},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)
	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-un-freeze.json"), string(genericResponse.Docs[0].Source))

	// Set new uris
	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.ESDTSetNewURIs),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("uri"), []byte("uri"), []byte("uri"), []byte("uri"), []byte("uri")},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)
	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-set-new-uris.json"), string(genericResponse.Docs[0].Source))

	// new creator
	newCreator := "erd12m3x8jp6dl027pj5f2nw6ght2cyhhjfrs86cdwsa8xn83r375qfqrwpdx0"
	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(newCreator),
							Identifier: []byte(core.ESDTModifyCreator),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes()},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)
	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-new-creator.json"), string(genericResponse.Docs[0].Source))

	// new royalties
	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.ESDTModifyRoyalties),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), big.NewInt(100).Bytes()},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)
	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-new-royalties.json"), string(genericResponse.Docs[0].Source))
}

func TestCreateNFTAndMetaDataRecreate(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esdtCreateData := &esdt.ESDigitalToken{
		TokenMetaData: &esdt.MetaData{
			Name: []byte("token-token-token"),
			URIs: [][]byte{[]byte("uri"), []byte("uri")},
		},
	}
	marshalizedCreate, _ := json.Marshal(esdtCreateData)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
		ShardID:   1,
	}
	body := &dataBlock.Body{}

	// CREATE NFT data
	address := "erd1w7jyzuj6cv4ngw8luhlkakatjpmjh3ql95lmxphd3vssc4vpymks6k5th7"
	pool := &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("NEW-abcd"), big.NewInt(100).Bytes(), big.NewInt(1).Bytes(), marshalizedCreate},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids := []string{"NEW-abcd-64"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-before-recreate.json"), string(genericResponse.Docs[0].Source))

	// RECREATE
	reCreate := &esdt.ESDigitalToken{
		TokenMetaData: &esdt.MetaData{
			Name: []byte("token"),
			URIs: [][]byte{[]byte("uri")},
			Hash: []byte("hash"),
		},
	}
	marshalizedReCreate, _ := json.Marshal(reCreate)

	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.ESDTMetaDataRecreate),
							Topics:     [][]byte{[]byte("NEW-abcd"), big.NewInt(100).Bytes(), big.NewInt(0).Bytes(), marshalizedReCreate},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-recreate.json"), string(genericResponse.Docs[0].Source))

	// UPDATE
	update := &esdt.ESDigitalToken{
		TokenMetaData: &esdt.MetaData{
			Name: []byte("token-second"),
			URIs: [][]byte{[]byte("uri")},
			Hash: []byte("hash"),
		},
	}
	marshalizedUpdate, _ := json.Marshal(update)

	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.ESDTMetaDataUpdate),
							Topics:     [][]byte{[]byte("NEW-abcd"), big.NewInt(100).Bytes(), big.NewInt(0).Bytes(), marshalizedUpdate},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-update.json"), string(genericResponse.Docs[0].Source))
}

func TestMultipleESDTMetadataRecreate(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esdtCreateData := &esdt.ESDigitalToken{
		TokenMetaData: &esdt.MetaData{
			Name: []byte("YELLOW"),
			URIs: [][]byte{[]byte("uri"), []byte("uri")},
		},
	}
	marshalizedCreate, _ := json.Marshal(esdtCreateData)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
		ShardID:   1,
	}
	body := &dataBlock.Body{}

	// CREATE NFT data
	address := "erd1w7jyzuj6cv4ngw8luhlkakatjpmjh3ql95lmxphd3vssc4vpymks6k5th7"
	pool := &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("COLORS-df0e82"), big.NewInt(1).Bytes(), big.NewInt(1).Bytes(), marshalizedCreate},
						},
						nil,
					},
				},
			},
			{
				TxHash: hex.EncodeToString([]byte("h2")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("COLORS-df0e82"), big.NewInt(2).Bytes(), big.NewInt(1).Bytes(), marshalizedCreate},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	// RECREATE
	reCreate := &esdt.ESDigitalToken{
		TokenMetaData: &esdt.MetaData{
			Name: []byte("GREEN"),
			URIs: [][]byte{[]byte("uri")},
			Hash: []byte("hash"),
		},
	}
	marshalizedReCreate, _ := json.Marshal(reCreate)

	pool = &outport.TransactionPool{
		Logs: []*outport.LogData{
			{
				TxHash: hex.EncodeToString([]byte("h1")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.ESDTMetaDataRecreate),
							Topics:     [][]byte{[]byte("COLORS-df0e82"), big.NewInt(1).Bytes(), big.NewInt(0).Bytes(), marshalizedReCreate},
						},
						nil,
					},
				},
			},
			{
				TxHash: hex.EncodeToString([]byte("h2")),
				Log: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.ESDTMetaDataRecreate),
							Topics:     [][]byte{[]byte("COLORS-df0e82"), big.NewInt(2).Bytes(), big.NewInt(0).Bytes(), marshalizedReCreate},
						},
						nil,
					},
				},
			},
		},
	}
	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, nil, testNumOfShards))
	require.Nil(t, err)

	ids := []string{"COLORS-df0e82-01", "COLORS-df0e82-02"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-color-1.json"), string(genericResponse.Docs[0].Source))
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-color-2.json"), string(genericResponse.Docs[1].Source))
}
