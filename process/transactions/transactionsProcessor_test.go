package transactions

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/pubkeyConverter"
	nodeData "github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/indexer"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgsTxsDBProc() *ArgsTransactionProcessor {
	args := &ArgsTransactionProcessor{
		AddressPubkeyConverter: &mock.PubkeyConverterMock{},
		TxFeeCalculator:        &mock.EconomicsHandlerStub{},
		ShardCoordinator:       &mock.ShardCoordinatorMock{},
		Hasher:                 &mock.HasherMock{},
		Marshalizer:            &mock.MarshalizerMock{},
		IsInImportMode:         false,
	}
	return args
}

func TestAddToAlteredAddresses(t *testing.T) {
	t.Parallel()

	sender := "senderAddress"
	receiver := "receiverAddress"
	tokenIdentifier := "Test-token"
	tx := &data.Transaction{
		Sender:              sender,
		Receiver:            receiver,
		EsdtTokenIdentifier: tokenIdentifier,
		Data:                []byte("ESDTTransfer@31323334352d373066366534@174876e800"),
	}
	alteredAddress := data.NewAlteredAccounts()
	selfShardID := uint32(0)
	mb := &block.MiniBlock{}

	grouper := txsGrouper{
		txBuilder: &dbTransactionBuilder{
			esdtProc: newEsdtTransactionHandler(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{}),
		},
	}
	grouper.addToAlteredAddresses(tx, alteredAddress, mb, selfShardID, false)

	alteredAccounts, ok := alteredAddress.Get(receiver)
	require.True(t, ok)
	require.Equal(t, &data.AlteredAccount{
		IsESDTOperation: true,
		TokenIdentifier: tokenIdentifier,
	}, alteredAccounts[0])

	alteredAccounts, ok = alteredAddress.Get(sender)
	require.True(t, ok)
	require.Equal(t, &data.AlteredAccount{
		IsSender:        true,
		IsESDTOperation: true,
		TokenIdentifier: tokenIdentifier,
	}, alteredAccounts[0])
}

func TestTestAddToAlteredAddressesESDTOnMeta(t *testing.T) {
	sender := "senderAddress"
	receiver := "receiverAddress"
	tokenIdentifier := "Test-token"
	tx := &data.Transaction{
		Sender:              sender,
		Receiver:            receiver,
		EsdtTokenIdentifier: tokenIdentifier,
	}
	alteredAddress := data.NewAlteredAccounts()
	selfShardID := core.MetachainShardId
	mb := &block.MiniBlock{
		ReceiverShardID: core.MetachainShardId,
	}

	grouper := txsGrouper{
		txBuilder: &dbTransactionBuilder{
			esdtProc: newEsdtTransactionHandler(&mock.PubkeyConverterMock{}, &mock.ShardCoordinatorMock{}),
		},
	}
	grouper.addToAlteredAddresses(tx, alteredAddress, mb, selfShardID, false)

	alteredAccounts, ok := alteredAddress.Get(receiver)
	require.True(t, ok)
	require.Equal(t, &data.AlteredAccount{
		IsESDTOperation: false,
		TokenIdentifier: tokenIdentifier,
	}, alteredAccounts[0])

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
	tx1 := &transaction.Transaction{
		GasLimit: 100,
		GasPrice: 100,
	}
	txHash2 := []byte("txHash2")
	tx2 := &transaction.Transaction{
		GasLimit: 100,
		GasPrice: 100,
	}
	txHash3 := []byte("txHash3")
	tx3 := &transaction.Transaction{}
	txHash4 := []byte("txHash4")
	tx4 := &transaction.Transaction{}
	txHash5 := []byte("txHash5")
	tx5 := &transaction.Transaction{}

	rTx1Hash := []byte("rTxHash1")
	rTx1 := &rewardTx.RewardTx{}
	rTx2Hash := []byte("rTxHash2")
	rTx2 := &rewardTx.RewardTx{}

	recHash1 := []byte("recHash1")
	rec1 := &receipt.Receipt{
		Value:  big.NewInt(100),
		TxHash: txHash1,
	}
	recHash2 := []byte("recHash2")
	rec2 := &receipt.Receipt{
		Value:  big.NewInt(200),
		TxHash: txHash2,
	}

	scHash1 := []byte("scHash1")
	scResult1 := &smartContractResult.SmartContractResult{
		OriginalTxHash: txHash1,
		PrevTxHash:     txHash1,
		GasLimit:       1,
	}
	scHash2 := []byte("scHash2")
	scResult2 := &smartContractResult.SmartContractResult{
		OriginalTxHash: txHash1,
		PrevTxHash:     txHash1,
		GasLimit:       1,
	}
	scHash3 := []byte("scHash3")
	scResult3 := &smartContractResult.SmartContractResult{
		OriginalTxHash: txHash3,
		Data:           []byte("@" + "6F6B"),
	}

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

	pool := &indexer.Pool{
		Txs: map[string]nodeData.TransactionHandler{
			string(txHash1): tx1,
			string(txHash2): tx2,
			string(txHash3): tx3,
			string(txHash4): tx4,
		},
		Scrs: map[string]nodeData.TransactionHandler{
			string(scHash1): scResult1,
			string(scHash2): scResult2,
			string(scHash3): scResult3,
		},
		Rewards: map[string]nodeData.TransactionHandler{
			string(rTx1Hash): rTx1,
			string(rTx2Hash): rTx2,
		},
		Invalid: map[string]nodeData.TransactionHandler{
			string(txHash5): tx5,
		},
		Receipts: map[string]nodeData.TransactionHandler{
			string(recHash1): rec1,
			string(recHash2): rec2,
		},
	}

	txDbProc, _ := NewTransactionsProcessor(createMockArgsTxsDBProc())

	results := txDbProc.PrepareTransactionsForDatabase(body, header, pool)
	assert.Equal(t, 7, len(results.Transactions))

}

func TestRelayedTransactions(t *testing.T) {
	t.Parallel()

	txHash1 := []byte("txHash1")
	tx1 := &transaction.Transaction{
		GasLimit: 100,
		GasPrice: 100,
		Data:     []byte("relayedTx@blablabllablalba"),
	}

	scHash1 := []byte("scHash1")
	scResult1 := &smartContractResult.SmartContractResult{
		OriginalTxHash: txHash1,
		PrevTxHash:     txHash1,
		GasLimit:       1,
	}
	scHash2 := []byte("scHash2")
	scResult2 := &smartContractResult.SmartContractResult{
		OriginalTxHash: txHash1,
		PrevTxHash:     txHash1,
		GasLimit:       1,
	}
	scHash3 := []byte("scHash3")
	scResult3 := &smartContractResult.SmartContractResult{
		OriginalTxHash: scHash1,
		Data:           []byte("@" + "6F6B"),
	}

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes: [][]byte{txHash1},
				Type:     block.TxBlock,
			},
			{
				TxHashes: [][]byte{scHash1, scHash2, scHash3},
				Type:     block.SmartContractResultBlock,
			},
		},
	}

	header := &block.Header{}

	pool := &indexer.Pool{
		Txs: map[string]nodeData.TransactionHandler{
			string(txHash1): tx1,
		},
		Scrs: map[string]nodeData.TransactionHandler{
			string(scHash1): scResult1,
			string(scHash2): scResult2,
			string(scHash3): scResult3,
		},
	}

	txDbProc, _ := NewTransactionsProcessor(createMockArgsTxsDBProc())

	results := txDbProc.PrepareTransactionsForDatabase(body, header, pool)
	assert.Equal(t, 1, len(results.Transactions))
	assert.Equal(t, 3, len(results.Transactions[0].SmartContractResults))
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

func TestAlteredAddresses(t *testing.T) {
	expectedAlteredAccounts := make(map[string]struct{})
	// addresses marked with a comment should be added to the altered addresses map

	// normal txs
	address1 := []byte("address1") // should be added
	address2 := []byte("address2")
	expectedAlteredAccounts[hex.EncodeToString(address1)] = struct{}{}
	tx1 := &transaction.Transaction{
		SndAddr: address1,
		RcvAddr: address2,
	}
	tx1Hash := []byte("tx1Hash")

	address3 := []byte("address3")
	address4 := []byte("address4") // should be added
	expectedAlteredAccounts[hex.EncodeToString(address4)] = struct{}{}
	tx2 := &transaction.Transaction{
		SndAddr: address3,
		RcvAddr: address4,
	}
	tx2Hash := []byte("tx2hash")

	txMiniBlock1 := &block.MiniBlock{
		Type:            block.TxBlock,
		TxHashes:        [][]byte{tx1Hash},
		SenderShardID:   0,
		ReceiverShardID: 1,
	}
	txMiniBlock2 := &block.MiniBlock{
		Type:            block.TxBlock,
		TxHashes:        [][]byte{tx2Hash},
		SenderShardID:   1,
		ReceiverShardID: 0,
	}

	// reward txs
	address5 := []byte("address5") // should be added
	expectedAlteredAccounts[hex.EncodeToString(address5)] = struct{}{}
	rwdTx1 := &rewardTx.RewardTx{
		RcvAddr: address5,
	}
	rwdTx1Hash := []byte("rwdTx1")

	address6 := []byte("address6")
	rwdTx2 := &rewardTx.RewardTx{
		RcvAddr: address6,
	}
	rwdTx2Hash := []byte("rwdTx2")

	rewTxMiniBlock1 := &block.MiniBlock{
		Type:            block.RewardsBlock,
		TxHashes:        [][]byte{rwdTx1Hash},
		SenderShardID:   core.MetachainShardId,
		ReceiverShardID: 0,
	}
	rewTxMiniBlock2 := &block.MiniBlock{
		Type:            block.RewardsBlock,
		TxHashes:        [][]byte{rwdTx2Hash},
		SenderShardID:   core.MetachainShardId,
		ReceiverShardID: 1,
	}

	// smart contract results
	address7 := []byte("address7") // should be added
	address8 := []byte("address8")
	expectedAlteredAccounts[hex.EncodeToString(address7)] = struct{}{}
	scr1 := &smartContractResult.SmartContractResult{
		RcvAddr: address7,
		SndAddr: address8,
	}
	scr1Hash := []byte("scr1Hash")

	address9 := []byte("address9") // should be added
	address10 := []byte("address10")
	expectedAlteredAccounts[hex.EncodeToString(address9)] = struct{}{}
	scr2 := &smartContractResult.SmartContractResult{
		RcvAddr: address9,
		SndAddr: address10,
	}
	scr2Hash := []byte("scr2Hash")

	scrMiniBlock1 := &block.MiniBlock{
		Type:            block.SmartContractResultBlock,
		TxHashes:        [][]byte{scr1Hash, scr2Hash},
		SenderShardID:   1,
		ReceiverShardID: 0,
	}

	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{txMiniBlock1, txMiniBlock2, rewTxMiniBlock1, rewTxMiniBlock2, scrMiniBlock1},
	}

	hdr := &block.Header{}

	pool := &indexer.Pool{
		Txs: map[string]nodeData.TransactionHandler{
			string(tx1Hash): tx1,
			string(tx2Hash): tx2,
		},
		Scrs: map[string]nodeData.TransactionHandler{
			string(scr1Hash): scr1,
			string(scr2Hash): scr2,
		},
		Rewards: map[string]nodeData.TransactionHandler{
			string(rwdTx1Hash): rwdTx1,
			string(rwdTx2Hash): rwdTx2,
		},
	}

	shardCoordinator := &mock.ShardCoordinatorMock{
		ComputeIdCalled: func(address []byte) uint32 {
			switch string(address) {
			case string(address1), string(address4), string(address5), string(address7), string(address9):
				return 0
			default:
				return 1
			}
		},
	}

	args := createMockArgsTxsDBProc()
	args.ShardCoordinator = shardCoordinator
	txDbProc, _ := NewTransactionsProcessor(args)

	results := txDbProc.PrepareTransactionsForDatabase(body, hdr, pool)

	for addrActual := range results.AlteredAccts.GetAll() {
		_, found := expectedAlteredAccounts[addrActual]
		if !found {
			assert.Fail(t, fmt.Sprintf("address %s not found", addrActual))
		}
	}
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
	tx1 := &transaction.Transaction{
		GasLimit: 100,
		GasPrice: 100,
	}
	recHash1 := []byte("recHash1")
	rec1 := &receipt.Receipt{
		Value:  big.NewInt(100),
		TxHash: txHash1,
	}

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

	pool := &indexer.Pool{
		Invalid: map[string]nodeData.TransactionHandler{
			string(txHash1): tx1,
		},
		Receipts: map[string]nodeData.TransactionHandler{
			string(recHash1): rec1,
		},
	}

	results := txDbProc.PrepareTransactionsForDatabase(body, header, pool)
	require.Len(t, results.Transactions, 1)
	require.Equal(t, tx1.GasLimit, results.Transactions[0].GasUsed)
}

func TestCheckGasUsedRelayedTransaction(t *testing.T) {
	t.Parallel()

	txDbProc, _ := NewTransactionsProcessor(createMockArgsTxsDBProc())

	txHash1 := []byte("txHash1")
	tx1 := &transaction.Transaction{
		GasLimit: 100,
		GasPrice: 123456,
		Data:     []byte("relayedTx@1231231231239129312"),
	}
	scResHash1 := []byte("scResHash1")
	scRes1 := &smartContractResult.SmartContractResult{
		OriginalTxHash: txHash1,
	}

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

	pool := &indexer.Pool{
		Txs: map[string]nodeData.TransactionHandler{
			string(txHash1): tx1,
		},
		Scrs: map[string]nodeData.TransactionHandler{
			string(scResHash1): scRes1,
		},
	}

	results := txDbProc.PrepareTransactionsForDatabase(body, header, pool)
	require.Len(t, results.Transactions, 1)
	require.Equal(t, tx1.GasLimit, results.Transactions[0].GasUsed)
}

func TestGetRewardsTxsHashesHexEncoded(t *testing.T) {
	t.Parallel()

	txDBProc, _ := NewTransactionsProcessor(createMockArgsTxsDBProc())

	res := txDBProc.GetRewardsTxsHashesHexEncoded(nil, nil)
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
					[]byte("h2"),
				},
				Type: block.TxBlock,
			},
		},
	}

	expectedHashes := []string{
		"6831", "6832",
	}
	txsHashes := txDBProc.GetRewardsTxsHashesHexEncoded(header, body)
	require.Equal(t, expectedHashes, txsHashes)
}

func TestTxsDatabaseProcessor_PrepareTransactionsForDatabaseInvalidTxWithSCR(t *testing.T) {
	t.Parallel()

	txDbProc, _ := NewTransactionsProcessor(createMockArgsTxsDBProc())

	txHash1 := []byte("txHash1")
	tx1 := &transaction.Transaction{
		GasLimit: 100,
		GasPrice: 123456,
		Data:     []byte("ESDTTransfer@54474e2d383862383366@0a"),
	}
	scResHash1 := []byte("scResHash1")
	scRes1 := &smartContractResult.SmartContractResult{
		OriginalTxHash: txHash1,
	}

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

	pool := &indexer.Pool{
		Invalid: map[string]nodeData.TransactionHandler{
			string(txHash1): tx1,
		},
		Scrs: map[string]nodeData.TransactionHandler{
			string(scResHash1): scRes1,
		},
	}

	results := txDbProc.PrepareTransactionsForDatabase(body, header, pool)
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
	tx1 := &transaction.Transaction{
		GasLimit: 100,
		GasPrice: 123456,
		Data:     []byte("ESDTNFTTransfer@595959453643392D303837363661@01@01@000000000000000005005C83E0C42EDCE394F40B24D29D298B0249C41F028974@66756E64@890479AFC610F4BEBC087D3ADA3F7C2775C736BBA91F41FD3D65092AA482D8B0@1c20"),
	}
	scResHash1 := []byte("scResHash1")
	scRes1 := &smartContractResult.SmartContractResult{
		OriginalTxHash: txHash1,
	}

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

	pool := &indexer.Pool{
		Txs: map[string]nodeData.TransactionHandler{
			string(txHash1): tx1,
		},
		Scrs: map[string]nodeData.TransactionHandler{
			string(scResHash1): scRes1,
		},
	}

	results := txDbProc.PrepareTransactionsForDatabase(body, header, pool)
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
	pubKeyConv, _ := pubkeyConverter.NewBech32PubkeyConverter(32)
	args.AddressPubkeyConverter = pubKeyConv
	args.ShardCoordinator = &mock.ShardCoordinatorMock{SelfID: core.MetachainShardId}
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
	header := &block.Header{}
	pool := &indexer.Pool{
		Txs: map[string]nodeData.TransactionHandler{
			"t1": &transaction.Transaction{
				SndAddr: decodeBech32("erd1dglncxk6sl9a3xumj78n6z2xux4ghp5c92cstv5zsn56tjgtdwpsk46qrs"),
				RcvAddr: decodeBech32("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"),
				Data:    []byte("issue@4141414141@41414141414141@0186a0@01@63616e467265657a65@74727565@63616e57697065@74727565@63616e5061757365@74727565@63616e4d696e74@74727565@63616e4275726e@74727565@63616e4368616e67654f776e6572@74727565@63616e55706772616465@74727565"),
			},
		},
		Scrs: map[string]nodeData.TransactionHandler{
			"scr1": &smartContractResult.SmartContractResult{
				OriginalTxHash: []byte("t1"),
				Data:           []byte("ESDTTransfer@414141414141412d323436626461@0186a0"),
				SndAddr:        decodeBech32("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"),
				RcvAddr:        decodeBech32("erd1dglncxk6sl9a3xumj78n6z2xux4ghp5c92cstv5zsn56tjgtdwpsk46qrs"),
			},
			"scr2": &smartContractResult.SmartContractResult{
				OriginalTxHash: []byte("t1"),
				Data:           []byte("@6f6b"),
				SndAddr:        decodeBech32("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"),
				RcvAddr:        decodeBech32("erd1dglncxk6sl9a3xumj78n6z2xux4ghp5c92cstv5zsn56tjgtdwpsk46qrs"),
			},
		},
	}

	res := txDbProc.PrepareTransactionsForDatabase(body, header, pool)
	require.Equal(t, "success", res.Transactions[0].Status)
	require.Equal(t, 2, len(res.ScResults))

	_, ok := res.AlteredAccts.Get("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u")
	require.True(t, ok)

	// transaction fail
	pool = &indexer.Pool{
		Txs: map[string]nodeData.TransactionHandler{
			"t1": &transaction.Transaction{
				SndAddr: decodeBech32("erd1dglncxk6sl9a3xumj78n6z2xux4ghp5c92cstv5zsn56tjgtdwpsk46qrs"),
				RcvAddr: decodeBech32("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"),
				Data:    []byte("issue@4141414141@41414141414141@0186a0@01@63616e467265657a65@74727565@63616e57697065@74727565@63616e5061757365@74727565@63616e4d696e74@74727565@63616e4275726e@74727565@63616e4368616e67654f776e6572@74727565@63616e55706772616465@74727565"),
			},
		},
		Scrs: map[string]nodeData.TransactionHandler{
			"scr1": &smartContractResult.SmartContractResult{
				OriginalTxHash: []byte("t1"),
				Data:           []byte("75736572206572726f72"),
				SndAddr:        decodeBech32("erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"),
				RcvAddr:        decodeBech32("erd1dglncxk6sl9a3xumj78n6z2xux4ghp5c92cstv5zsn56tjgtdwpsk46qrs"),
			},
		},
	}

	res = txDbProc.PrepareTransactionsForDatabase(body, header, pool)
	require.Equal(t, "fail", res.Transactions[0].Status)
	require.Equal(t, 1, len(res.ScResults))
}
