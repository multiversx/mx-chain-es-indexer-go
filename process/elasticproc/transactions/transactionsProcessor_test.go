package transactions

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/receipt"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgsTxsDBProc() *ArgsTransactionProcessor {
	args := &ArgsTransactionProcessor{
		AddressPubkeyConverter: mock.NewPubkeyConverterMock(10),
		Hasher:                 &mock.HasherMock{},
		Marshalizer:            &mock.MarshalizerMock{},
	}
	return args
}

func TestIsSCRForSenderWithGasUsed(t *testing.T) {
	t.Parallel()

	txHash := "txHash"
	nonce := uint64(10)
	sender := "sender"

	tx := &data.Transaction{
		Hash:   txHash,
		Nonce:  nonce,
		Sender: sender,
	}
	sc := &data.ScResult{
		Data:       []byte("@6f6b@something"),
		Nonce:      nonce + 1,
		Receiver:   sender,
		PrevTxHash: txHash,
	}

	require.True(t, isSCRForSenderWithRefund(sc, tx))
}

func TestPrepareTransactionsForDatabase(t *testing.T) {
	t.Parallel()

	txHash1 := []byte("txHash1")
	tx1 := outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
		GasLimit: 100,
		GasPrice: 100,
	}, 0, big.NewInt(0))
	txHash2 := []byte("txHash2")
	tx2 := outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
		GasLimit: 100,
		GasPrice: 100,
	}, 0, big.NewInt(0))
	txHash3 := []byte("txHash3")
	tx3 := outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{}, 0, big.NewInt(0))
	txHash4 := []byte("txHash4")
	tx4 := outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{}, 0, big.NewInt(0))
	txHash5 := []byte("txHash5")
	tx5 := outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{}, 0, big.NewInt(0))

	rTx1Hash := []byte("rTxHash1")
	rTx1 := outport.NewTransactionHandlerWithGasAndFee(&rewardTx.RewardTx{}, 0, big.NewInt(0))
	rTx2Hash := []byte("rTxHash2")
	rTx2 := outport.NewTransactionHandlerWithGasAndFee(&rewardTx.RewardTx{}, 0, big.NewInt(0))

	recHash1 := []byte("recHash1")
	rec1 := outport.NewTransactionHandlerWithGasAndFee(&receipt.Receipt{
		Value:  big.NewInt(100),
		TxHash: txHash1,
	}, 0, big.NewInt(0))
	recHash2 := []byte("recHash2")
	rec2 := outport.NewTransactionHandlerWithGasAndFee(&receipt.Receipt{
		Value:  big.NewInt(200),
		TxHash: txHash2,
	}, 0, big.NewInt(0))

	scHash1 := []byte("scHash1")
	scResult1 := outport.NewTransactionHandlerWithGasAndFee(&smartContractResult.SmartContractResult{
		OriginalTxHash: txHash1,
		PrevTxHash:     txHash1,
		GasLimit:       1,
	}, 0, big.NewInt(0))
	scHash2 := []byte("scHash2")
	scResult2 := outport.NewTransactionHandlerWithGasAndFee(&smartContractResult.SmartContractResult{
		OriginalTxHash: txHash1,
		PrevTxHash:     txHash1,
		GasLimit:       1,
	}, 0, big.NewInt(0))
	scHash3 := []byte("scHash3")
	scResult3 := outport.NewTransactionHandlerWithGasAndFee(&smartContractResult.SmartContractResult{
		OriginalTxHash: txHash3,
		Data:           []byte("@" + "6F6B"),
	}, 0, big.NewInt(0))

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes: [][]byte{txHash1, txHash2, txHash3},
				Type:     block.TxBlock,
			},
			{
				TxHashes: [][]byte{txHash4},
				Type:     block.TxBlock,
			},
			{
				TxHashes: [][]byte{scHash1, scHash2},
				Type:     block.SmartContractResultBlock,
			},
			{
				TxHashes: [][]byte{scHash3},
				Type:     block.SmartContractResultBlock,
			},
			{
				TxHashes: [][]byte{recHash1, recHash2},
				Type:     block.ReceiptBlock,
			},
			{
				TxHashes: [][]byte{rTx1Hash, rTx2Hash},
				Type:     block.RewardsBlock,
			},
			{
				TxHashes: [][]byte{txHash5},
				Type:     block.InvalidBlock,
			},
		},
	}
	header := &block.Header{}

	pool := &outport.Pool{
		Txs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(txHash1): tx1,
			string(txHash2): tx2,
			string(txHash3): tx3,
			string(txHash4): tx4,
		},
		Scrs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(scHash1): scResult1,
			string(scHash2): scResult2,
			string(scHash3): scResult3,
		},
		Rewards: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(rTx1Hash): rTx1,
			string(rTx2Hash): rTx2,
		},
		Invalid: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(txHash5): tx5,
		},
		Receipts: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(recHash1): rec1,
			string(recHash2): rec2,
		},
	}

	txDbProc, _ := NewTransactionsProcessor(createMockArgsTxsDBProc())

	results := txDbProc.PrepareTransactionsForDatabase(body, header, pool, false, 3)
	assert.Equal(t, 7, len(results.Transactions))

}

func TestRelayedTransactions(t *testing.T) {
	t.Parallel()

	txHash1 := []byte("txHash1")
	tx1 := outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
		GasLimit: 100,
		GasPrice: 100,
		Data:     []byte("relayedTx@blablabllablalba"),
	}, 0, big.NewInt(0))

	scHash1 := []byte("scHash1")
	scResult1 := outport.NewTransactionHandlerWithGasAndFee(&smartContractResult.SmartContractResult{
		OriginalTxHash: txHash1,
		PrevTxHash:     txHash1,
		GasLimit:       1,
	}, 0, big.NewInt(0))
	scHash2 := []byte("scHash2")
	scResult2 := outport.NewTransactionHandlerWithGasAndFee(&smartContractResult.SmartContractResult{
		OriginalTxHash: txHash1,
		PrevTxHash:     txHash1,
		GasLimit:       1,
	}, 0, big.NewInt(0))

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes: [][]byte{txHash1},
				Type:     block.TxBlock,
			},
			{
				TxHashes: [][]byte{scHash1, scHash2},
				Type:     block.SmartContractResultBlock,
			},
		},
	}

	header := &block.Header{}

	pool := &outport.Pool{
		Txs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(txHash1): tx1,
		},
		Scrs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(scHash1): scResult1,
			string(scHash2): scResult2,
		},
	}

	txDbProc, _ := NewTransactionsProcessor(createMockArgsTxsDBProc())

	results := txDbProc.PrepareTransactionsForDatabase(body, header, pool, false, 3)
	assert.Equal(t, 1, len(results.Transactions))
	assert.Equal(t, 2, len(results.Transactions[0].SmartContractResults))
	assert.Equal(t, transaction.TxStatusSuccess.String(), results.Transactions[0].Status)
}

func TestSetTransactionSearchOrder(t *testing.T) {
	t.Parallel()
	txHash1 := []byte("txHash1")
	tx1 := &data.Transaction{}

	txHash2 := []byte("txHash2")
	tx2 := &data.Transaction{}

	txPool := map[string]*data.Transaction{
		string(txHash1): tx1,
		string(txHash2): tx2,
	}

	txDbProc, _ := NewTransactionsProcessor(createMockArgsTxsDBProc())

	transactions := txDbProc.setTransactionSearchOrder(txPool)
	assert.True(t, txPoolHasSearchOrder(transactions, 0))
	assert.True(t, txPoolHasSearchOrder(transactions, 1))

	transactions = txDbProc.setTransactionSearchOrder(txPool)
	assert.True(t, txPoolHasSearchOrder(transactions, 0))
	assert.True(t, txPoolHasSearchOrder(transactions, 1))

	transactions = txDbProc.setTransactionSearchOrder(txPool)
	assert.True(t, txPoolHasSearchOrder(transactions, 0))
	assert.True(t, txPoolHasSearchOrder(transactions, 1))
}

func txPoolHasSearchOrder(txPool map[string]*data.Transaction, searchOrder uint32) bool {
	for _, tx := range txPool {
		if tx.SearchOrder == searchOrder {
			return true
		}
	}

	return false
}

func TestCheckGasUsedTooMuchGasProvidedCase(t *testing.T) {
	t.Parallel()

	txHash := "txHash"
	nonce := uint64(10)
	sender := "sender"

	tx := &data.Transaction{
		Hash:   txHash,
		Nonce:  nonce,
		Sender: sender,
	}
	sc := &data.ScResult{
		Data:       []byte("@6f6b@something"),
		Nonce:      nonce + 1,
		Receiver:   sender,
		PrevTxHash: txHash,
	}

	require.True(t, isSCRForSenderWithRefund(sc, tx))
}

func TestCheckGasUsedInvalidTransaction(t *testing.T) {
	t.Parallel()

	txDbProc, _ := NewTransactionsProcessor(createMockArgsTxsDBProc())

	txHash1 := []byte("txHash1")
	tx1 := outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
		GasLimit: 100,
		GasPrice: 100,
	}, 100, big.NewInt(0))
	recHash1 := []byte("recHash1")
	rec1 := outport.NewTransactionHandlerWithGasAndFee(&receipt.Receipt{
		Value:  big.NewInt(100),
		TxHash: txHash1,
	}, 0, big.NewInt(0))

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes: [][]byte{txHash1},
				Type:     block.InvalidBlock,
			},
			{
				TxHashes: [][]byte{recHash1},
				Type:     block.ReceiptBlock,
			},
		},
	}

	header := &block.Header{}

	pool := &outport.Pool{
		Invalid: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(txHash1): tx1,
		},
		Receipts: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(recHash1): rec1,
		},
	}

	results := txDbProc.PrepareTransactionsForDatabase(body, header, pool, false, 3)
	require.Len(t, results.Transactions, 1)
	require.Equal(t, tx1.GetGasLimit(), results.Transactions[0].GasUsed)
}

func TestGetRewardsTxsHashesHexEncoded(t *testing.T) {
	t.Parallel()

	txDBProc, _ := NewTransactionsProcessor(createMockArgsTxsDBProc())

	res, _ := txDBProc.GetHexEncodedHashesForRemove(nil, nil)
	require.Nil(t, res)

	header := &block.Header{
		ShardID: core.MetachainShardId,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{},
		},
	}
	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes: [][]byte{
					[]byte("h1"),
				},
				Type: block.RewardsBlock,
			},
			{
				TxHashes: [][]byte{
					[]byte("h2"),
				},
				Type: block.RewardsBlock,
			},
			{
				TxHashes: [][]byte{
					[]byte("h3"),
				},
				Type: block.TxBlock,
			},
			{
				TxHashes: [][]byte{
					[]byte("h4"),
				},
				Type:            block.TxBlock,
				SenderShardID:   core.MetachainShardId,
				ReceiverShardID: 0,
			},
			{
				TxHashes: [][]byte{
					[]byte("h5"),
				},
				Type:            block.TxBlock,
				SenderShardID:   2,
				ReceiverShardID: core.MetachainShardId,
			},
			{
				TxHashes: [][]byte{
					[]byte("h6"),
				},
				Type:            block.SmartContractResultBlock,
				SenderShardID:   2,
				ReceiverShardID: core.MetachainShardId,
			},
			{
				TxHashes: [][]byte{
					[]byte("h7"),
				},
				Type:            block.SmartContractResultBlock,
				SenderShardID:   core.MetachainShardId,
				ReceiverShardID: 0,
			},
		},
	}

	expectedHashes := []string{
		"6831", "6832", "6833", "6835",
	}
	expectedScrHashes := []string{
		"6836", "6837",
	}

	txsHashes, scrHashes := txDBProc.GetHexEncodedHashesForRemove(header, body)
	require.Equal(t, expectedHashes, txsHashes)
	require.Equal(t, expectedScrHashes, scrHashes)
}

func TestTxsDatabaseProcessor_PrepareTransactionsForDatabaseInvalidTxWithSCR(t *testing.T) {
	t.Parallel()

	txDbProc, _ := NewTransactionsProcessor(createMockArgsTxsDBProc())

	txHash1 := []byte("txHash1")
	tx1 := outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
		GasLimit: 100,
		GasPrice: 123456,
		Data:     []byte("ESDTTransfer@54474e2d383862383366@0a"),
	}, 100, big.NewInt(0))
	scResHash1 := []byte("scResHash1")
	scRes1 := outport.NewTransactionHandlerWithGasAndFee(&smartContractResult.SmartContractResult{
		OriginalTxHash: txHash1,
	}, 0, big.NewInt(0))

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes: [][]byte{txHash1},
				Type:     block.InvalidBlock,
			},
			{
				TxHashes: [][]byte{scResHash1},
				Type:     block.SmartContractResultBlock,
			},
		},
	}

	header := &block.Header{}

	pool := &outport.Pool{
		Invalid: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(txHash1): tx1,
		},
		Scrs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(scResHash1): scRes1,
		},
	}

	results := txDbProc.PrepareTransactionsForDatabase(body, header, pool, false, 3)
	require.NotNil(t, results)
	require.Len(t, results.Transactions, 1)
	require.Len(t, results.ScResults, 1)

	resultedTx := results.Transactions[0]
	require.Equal(t, transaction.TxStatusInvalid.String(), resultedTx.Status)
	require.Len(t, resultedTx.SmartContractResults, 1)
	require.Equal(t, resultedTx.GasLimit, resultedTx.GasUsed)
}

func TestTxsDatabaseProcessor_PrepareTransactionsForDatabaseESDTNFTTransfer(t *testing.T) {
	t.Parallel()

	txDbProc, _ := NewTransactionsProcessor(createMockArgsTxsDBProc())

	txHash1 := []byte("txHash1")
	tx1 := outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
		GasLimit: 100,
		GasPrice: 123456,
		Data:     []byte("ESDTNFTTransfer@595959453643392D303837363661@01@01@000000000000000005005C83E0C42EDCE394F40B24D29D298B0249C41F028974@66756E64@890479AFC610F4BEBC087D3ADA3F7C2775C736BBA91F41FD3D65092AA482D8B0@1c20"),
	}, 100, big.NewInt(0))
	scResHash1 := []byte("scResHash1")
	scRes1 := outport.NewTransactionHandlerWithGasAndFee(&smartContractResult.SmartContractResult{
		Nonce:          1,
		Data:           []byte("@" + okHexEncoded),
		OriginalTxHash: txHash1,
		PrevTxHash:     txHash1,
	}, 0, big.NewInt(0))

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes: [][]byte{txHash1},
				Type:     block.TxBlock,
			},
			{
				TxHashes: [][]byte{scResHash1},
				Type:     block.SmartContractResultBlock,
			},
		},
	}

	header := &block.Header{}

	pool := &outport.Pool{
		Txs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(txHash1): tx1,
		},
		Scrs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			string(scResHash1): scRes1,
		},
	}

	results := txDbProc.PrepareTransactionsForDatabase(body, header, pool, false, 3)
	require.NotNil(t, results)
	require.Len(t, results.Transactions, 1)
	require.Len(t, results.ScResults, 1)

	resultedTx := results.Transactions[0]
	require.Equal(t, transaction.TxStatusSuccess.String(), resultedTx.Status)
	require.Len(t, resultedTx.SmartContractResults, 1)
	require.Equal(t, resultedTx.GasLimit, resultedTx.GasUsed)
}

func TestTxsDatabaseProcessor_IssueESDTTx(t *testing.T) {
	t.Parallel()

	args := createMockArgsTxsDBProc()
	pubKeyConv, _ := pubkeyConverter.NewBech32PubkeyConverter(32, log)
	args.AddressPubkeyConverter = pubKeyConv
	txDbProc, _ := NewTransactionsProcessor(args)

	decodeBech32 := func(key string) []byte {
		decoded, _ := pubKeyConv.Decode(key)
		return decoded
	}

	// transaction success
	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes:        [][]byte{[]byte("t1")},
				Type:            block.TxBlock,
				SenderShardID:   0,
				ReceiverShardID: core.MetachainShardId,
			},
			{
				TxHashes: [][]byte{[]byte("scr1"), []byte("scr2")},
				Type:     block.SmartContractResultBlock,
			},
		},
	}
	header := &block.Header{
		ShardID: core.MetachainShardId,
	}
	pool := &outport.Pool{
		Txs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			"t1": outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
				SndAddr: decodeBech32("erd1dglncxk6sl9a3xumj78n6z2xux4ghp5c92cstv5zsn56tjgtdwpsk46qrs"),
				RcvAddr: decodeBech32("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"),
				Data:    []byte("issue@4141414141@41414141414141@0186a0@01@63616e467265657a65@74727565@63616e57697065@74727565@63616e5061757365@74727565@63616e4d696e74@74727565@63616e4275726e@74727565@63616e4368616e67654f776e6572@74727565@63616e55706772616465@74727565"),
			}, 0, big.NewInt(0)),
		},
		Scrs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			"scr1": outport.NewTransactionHandlerWithGasAndFee(&smartContractResult.SmartContractResult{
				OriginalTxHash: []byte("t1"),
				Data:           []byte("ESDTTransfer@414141414141412d323436626461@0186a0"),
				SndAddr:        decodeBech32("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"),
				RcvAddr:        decodeBech32("erd1dglncxk6sl9a3xumj78n6z2xux4ghp5c92cstv5zsn56tjgtdwpsk46qrs"),
			}, 0, big.NewInt(0)),
			"scr2": outport.NewTransactionHandlerWithGasAndFee(&smartContractResult.SmartContractResult{
				OriginalTxHash: []byte("t1"),
				Data:           []byte("@6f6b"),
				SndAddr:        decodeBech32("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"),
				RcvAddr:        decodeBech32("erd1dglncxk6sl9a3xumj78n6z2xux4ghp5c92cstv5zsn56tjgtdwpsk46qrs"),
			}, 0, big.NewInt(0)),
		},
	}

	res := txDbProc.PrepareTransactionsForDatabase(body, header, pool, false, 3)
	require.Equal(t, "success", res.Transactions[0].Status)
	require.Equal(t, 2, len(res.ScResults))

	// transaction fail
	pool = &outport.Pool{
		Txs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			"t1": outport.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
				SndAddr: decodeBech32("erd1dglncxk6sl9a3xumj78n6z2xux4ghp5c92cstv5zsn56tjgtdwpsk46qrs"),
				RcvAddr: decodeBech32("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"),
				Data:    []byte("issue@4141414141@41414141414141@0186a0@01@63616e467265657a65@74727565@63616e57697065@74727565@63616e5061757365@74727565@63616e4d696e74@74727565@63616e4275726e@74727565@63616e4368616e67654f776e6572@74727565@63616e55706772616465@74727565"),
			}, 0, big.NewInt(0)),
		},
		Scrs: map[string]coreData.TransactionHandlerWithGasUsedAndFee{
			"scr1": outport.NewTransactionHandlerWithGasAndFee(&smartContractResult.SmartContractResult{
				OriginalTxHash: []byte("t1"),
				Data:           []byte("75736572206572726f72"),
				SndAddr:        decodeBech32("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"),
				RcvAddr:        decodeBech32("erd1dglncxk6sl9a3xumj78n6z2xux4ghp5c92cstv5zsn56tjgtdwpsk46qrs"),
			}, 0, big.NewInt(0)),
		},
	}

	res = txDbProc.PrepareTransactionsForDatabase(body, header, pool, false, 3)
	require.Equal(t, "fail", res.Transactions[0].Status)
	require.Equal(t, 1, len(res.ScResults))
}
