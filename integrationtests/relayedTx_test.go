//go:build integrationtests

package integrationtests

import (
	"encoding/hex"
	"math/big"
	"testing"

	coreData "github.com/multiversx/mx-chain-core-go/data"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	indexerdata "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestRelayedTransactionGasUsedCrossShard(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	txHash := []byte("relayedTx")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}

	scrHash1 := []byte("scrHashRefund")
	body := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.TxBlock,
				SenderShardID:   0,
				ReceiverShardID: 0,
				TxHashes:        [][]byte{txHash},
			},
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   0,
				ReceiverShardID: 1,
				TxHashes:        [][]byte{scrHash1},
			},
		},
	}

	address1 := "erd1k7j6ewjsla4zsgv8v6f6fe3dvrkgv3d0d9jerczw45hzedhyed8sh2u34u"
	address2 := "erd14eyayfrvlrhzfrwg5zwleua25mkzgncggn35nvc6xhv5yxwml2es0f3dht"
	address3 := "erd1qqqqqqqqqqqqqpgq3dswlnnlkfd3gqrcv3dhzgnvh8ryf27g5rfsecnn2s"
	initialTx := &transaction.Transaction{
		Nonce:    1196667,
		SndAddr:  decodeAddress(address1),
		RcvAddr:  decodeAddress(address2),
		GasLimit: 16610000,
		GasPrice: 1000000000,
		Data:     []byte("relayedTx@7b226e6f6e6365223a322c2276616c7565223a302c227265636569766572223a22414141414141414141414146414974673738352f736c73554148686b57334569624c6e47524b76496f4e4d3d222c2273656e646572223a22726b6e534a477a343769534e794b43642f504f717075776b5477684534306d7a476a585a51686e622b724d3d222c226761735072696365223a313030303030303030302c226761734c696d6974223a31353030303030302c2264617461223a22633246325a5546306447567a644746306157397551444668597a49314d6a5935596d51335a44497759324a6959544d31596d566c4f4459314d4464684f574e6a4e7a677a5a4755774f445a694e4445334e546b345a54517a59544e6b5a6a566a593245795a5468684d6a6c414d6a51344e54677a4d574e6d4d5445304d54566d596a41354d6a63774e4451324e5755324e7a597a59574d314f4445345a5467314e4751345957526d4e54417a596a63354d6a6c6b4f54526c4e6d49794e6a49775a673d3d222c22636861696e4944223a224d513d3d222c2276657273696f6e223a312c227369676e6174757265223a225239462b34546352415a386d7771324559303163596c337662716c46657176387a76474a775a6833594d4f556b4234643451574e66376744626c484832576b71614a76614845744356617049713365356562384e41773d3d227d"),
		Value:    big.NewInt(0),
	}

	scr1 := &smartContractResult.SmartContractResult{
		Nonce:          2,
		GasPrice:       1000000000,
		GasLimit:       14732500,
		SndAddr:        decodeAddress(address2),
		RcvAddr:        decodeAddress(address3),
		Data:           []byte("aveAttestation@1ac25269bd7d20cbba35bee86507a9cc783de086b417598e43a3df5cca2e8a29@2485831cf11415fb092704465e6763ac5818e854d8adf503b7929d94e6b2620f"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
	}

	tx := outport.NewTransactionHandlerWithGasAndFee(initialTx, 16610000, big.NewInt(1760000000000000))
	tx.SetInitialPaidFee(big.NewInt(1760000000000000))

	pool := &outport.Pool{
		Txs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(txHash): tx,
		},
		Scrs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(scrHash1): outport.NewTransactionHandlerWithGasAndFee(scr1, 0, big.NewInt(0)),
		},
	}
	err = esProc.SaveTransactions(body, header, pool, nil, false, testNumOfShards)
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/relayedTx/relayed-tx-source.json"),
		string(genericResponse.Docs[0].Source),
	)

	// EXECUTE transfer on the destination shard
	bodyDstShard := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   0,
				ReceiverShardID: 1,
				TxHashes:        [][]byte{scrHash1},
			},
		},
	}
	scrWithRefund := []byte("scrWithRefund")
	refundValueBig, _ := big.NewInt(0).SetString("86271830000000", 10)
	poolDstShard := &outport.Pool{
		Scrs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(scrHash1): outport.NewTransactionHandlerWithGasAndFee(scr1, 0, big.NewInt(0)),
			string(scrWithRefund): outport.NewTransactionHandlerWithGasAndFee(&smartContractResult.SmartContractResult{
				Nonce:          3,
				SndAddr:        decodeAddress(address3),
				RcvAddr:        decodeAddress(address1),
				PrevTxHash:     []byte("f639cb7a0231191e04ec19dcb1359bd93a03fe8dc4a28a80d00835c5d1c988f8"),
				OriginalTxHash: txHash,
				Value:          refundValueBig,
				Data:           []byte(""),
				ReturnMessage:  []byte("gas refund for relayer"),
			}, 7982817, big.NewInt(1673728170000000)),
		},
	}

	err = esProc.SaveTransactions(bodyDstShard, header, poolDstShard, nil, false, testNumOfShards)
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/relayedTx/relayed-tx-after-refund.json"),
		string(genericResponse.Docs[0].Source),
	)
}

func TestRelayedTransactionIntraShard(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	txHash := []byte("relayedTxIntra")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}

	scrHash1 := []byte("scrAtt")
	body := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.TxBlock,
				SenderShardID:   0,
				ReceiverShardID: 0,
				TxHashes:        [][]byte{txHash},
			},
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   0,
				ReceiverShardID: 1,
				TxHashes:        [][]byte{scrHash1},
			},
		},
	}

	address1 := "erd1k7j6ewjsla4zsgv8v6f6fe3dvrkgv3d0d9jerczw45hzedhyed8sh2u34u"
	address2 := "erd14eyayfrvlrhzfrwg5zwleua25mkzgncggn35nvc6xhv5yxwml2es0f3dht"
	initialTx := &transaction.Transaction{
		Nonce:    1196665,
		SndAddr:  decodeAddress(address1),
		RcvAddr:  decodeAddress(address2),
		GasLimit: 15406000,
		GasPrice: 1000000000,
		Data:     []byte("relayedTx@7b226e6f6e6365223a302c2276616c7565223a302c227265636569766572223a22726b6e534a477a343769534e794b43642f504f717075776b5477684534306d7a476a585a51686e622b724d3d222c2273656e646572223a22726b6e534a477a343769534e794b43642f504f717075776b5477684534306d7a476a585a51686e622b724d3d222c226761735072696365223a313030303030303030302c226761734c696d6974223a31333233323030302c2264617461223a22553246325a55746c65565a686248566c514459794e6b55324d6b41324d6a5a464e6a497a4d544d354e6a55324d544d324e6b4d32515463784d7a6b32517a63304e6a55334e4463304e7a67324e544d334e7a517a4e7a59334e6b557a4d4463774d7a497a4f5464424e7a59334e545a444e6a45334e5459324e6a63334e544d304e7a457a4e7a59334e7a4a414e6a55334e44593451444d774e7a677a4d544d304e4451324d6a4d784e6a557a4d444d314e6a597a4e54517a4e6a457a4e544d314e6a497a4d544d7a4e4445304d7a59784d7a63324d7a4d304e6a4d7a4d444d784e6a597a4e5459304e44557a4e544d314e44497a4f54517a4e4445324e6a59784d7a637a4f5541324d6a63304e6a4e414e6a49324d7a4d784e7a45324d5459314d7a557a4f444d7a4d7a4d334d6a63794e7a597a4e444d774d7a417a4e545a444e6a593252444d314e7a593252444d324e7a5532515463324e6a517a4e44597a4e7a517a4f545a424e6a4d7a4e545a424e7a633351544d344e6a55334f446377222c22636861696e4944223a224d513d3d222c2276657273696f6e223a312c227369676e6174757265223a227166704a47767344444255514e2f5255474f5053755232484f4a614b70384536634e54773033433769345577762f4c54736d2b6a704239756c483966532b44716172714f6248417038666d72306a415531736e3541673d3d227d"),
		Value:    big.NewInt(0),
	}

	scr1 := &smartContractResult.SmartContractResult{
		Nonce:          0,
		GasPrice:       1000000000,
		GasLimit:       12750000,
		SndAddr:        decodeAddress(address2),
		RcvAddr:        decodeAddress(address2),
		Data:           []byte("SaveKeyValue@626E62@626E6231396561366C6A71396C746574747865377437676E307032397A76756C61756667753471376772@657468@307831344462316530356635436135356231334143613763346330316635644535354239434166613739@627463@62633171616535383333727276343030356C666D35766D36756A7664346374396A63356A777A38657870"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
	}

	refundValueBig, _ := big.NewInt(0).SetString("48500000000000", 10)
	scrHash2 := []byte("scrWithRefund")
	scr2 := &smartContractResult.SmartContractResult{
		Nonce:          1,
		RcvAddr:        decodeAddress(address2),
		SndAddr:        decodeAddress(address1),
		PrevTxHash:     []byte("a98ee38f22153ae9fb497504b228077fb515502946b87c7d570852476ca3329b"),
		OriginalTxHash: txHash,
		ReturnMessage:  []byte("gas refund for relayer"),
		Value:          refundValueBig,
	}

	tx := outport.NewTransactionHandlerWithGasAndFee(initialTx, 10556000, big.NewInt(2257820000000000))
	tx.SetInitialPaidFee(big.NewInt(2306320000000000))
	pool := &outport.Pool{
		Txs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(txHash): tx,
		},
		Scrs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(scrHash1): outport.NewTransactionHandlerWithGasAndFee(scr1, 0, big.NewInt(0)),
			string(scrHash2): outport.NewTransactionHandlerWithGasAndFee(scr2, 0, big.NewInt(0)),
		},
	}
	err = esProc.SaveTransactions(body, header, pool, nil, false, testNumOfShards)
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	require.JSONEq(t,
		readExpectedResult("./testdata/relayedTx/relayed-tx-intra.json"),
		string(genericResponse.Docs[0].Source),
	)
}
