//go:build integrationtests

package integrationtests

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	indexerdata "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestPerformance_IndexingTransactions(t *testing.T) {
	setLogLevelDebug()

	numTransactions := 5000

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	txHashes := make([][]byte, 0, numTransactions)
	txPool := make(map[string]*outport.TxInfo)
	logs := make([]*transaction.LogData, 0, numTransactions)
	alteredAccounts := make(map[string]*alteredAccount.AlteredAccount)

	for i := 0; i < numTransactions; i++ {
		txHashStr := fmt.Sprintf("bench-tx-hash-%d", i)
		txHash := []byte(txHashStr)
		txHashes = append(txHashes, txHash)

		tx := &transaction.Transaction{
			Nonce:    uint64(i),
			SndAddr:  decodeAddress("erd1w7jyzuj6cv4ngw8luhlkakatjpmjh3ql95lmxphd3vssc4vpymks6k5th7"),
			RcvAddr:  decodeAddress("erd1ahmy0yjhjg87n755yv99nzla22zzwfud55sa69gk3anyxyyucq9q2hgxww"),
			GasLimit: 70000,
			GasPrice: 1000000000,
			Data:     []byte("transfer"),
			Value:    big.NewInt(100),
		}

		txPool[hex.EncodeToString(txHash)] = &outport.TxInfo{
			Transaction: tx,
			FeeInfo: &outport.FeeInfo{
				GasUsed: 50000,
				Fee:     big.NewInt(50000000000),
			},
			ExecutionOrder: uint32(i),
		}

		numEvents := 10
		events := make([]*transaction.Event, 0, numEvents)
		for j := 0; j < numEvents; j++ {
			events = append(events, &transaction.Event{
				Address:    decodeAddress("erd1w7jyzuj6cv4ngw8luhlkakatjpmjh3ql95lmxphd3vssc4vpymks6k5th7"),
				Identifier: []byte(core.BuiltInFunctionESDTNFTTransfer),
				Topics:     [][]byte{[]byte("NFT-abcdef"), big.NewInt(int64(i)).Bytes(), big.NewInt(1).Bytes(), big.NewInt(int64(j)).Bytes()},
			})
		}

		logEntry := &transaction.Log{
			Address: decodeAddress("erd1w7jyzuj6cv4ngw8luhlkakatjpmjh3ql95lmxphd3vssc4vpymks6k5th7"),
			Events:  events,
		}
		logs = append(logs, &transaction.LogData{
			TxHash: hex.EncodeToString(txHash),
			Log:    logEntry,
		})

		mockAddrBytes := make([]byte, 32)
		big.NewInt(int64(i)).FillBytes(mockAddrBytes)
		mockAddrBytes[0] = 1
		mockAddr, _ := pubKeyConverter.Encode(mockAddrBytes)

		alteredAccounts[mockAddr] = &alteredAccount.AlteredAccount{
			Address: mockAddr,
			Balance: "1000000",
			AdditionalData: &alteredAccount.AdditionalAccountData{
				BalanceChanged: true,
			},
			Tokens: []*alteredAccount.AccountTokenData{
				{
					Identifier: "TEST-123456",
					Nonce:      uint64(i),
					Balance:    "100",
				},
			},
		}
	}

	header := &dataBlock.Header{
		Round:     100,
		TimeStamp: 10000,
		ShardID:   0,
	}

	body := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.TxBlock,
				SenderShardID:   0,
				ReceiverShardID: 0,
				TxHashes:        txHashes,
			},
		},
	}

	pool := &outport.TransactionPool{
		Transactions: txPool,
		Logs:         logs,
	}

	t.Logf("Starting indexing of %d transactions...", numTransactions)
	start := time.Now()

	err = esProc.SaveTransactions(createOutportBlockWithHeader(body, header, pool, alteredAccounts, testNumOfShards))
	require.Nil(t, err)

	duration := time.Since(start)
	t.Logf("Indexed %d transactions in %v", numTransactions, duration)
	t.Logf("Average time per transaction: %v", duration/time.Duration(numTransactions))

	// Verify one transaction to ensure indexing happened
	ids := []string{hex.EncodeToString(txHashes[0])}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)
	require.True(t, genericResponse.Docs[0].Found)

	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.LogsIndex, true, genericResponse)
	require.Nil(t, err)
	require.True(t, genericResponse.Docs[0].Found)

	mockAddrBytes := make([]byte, 32)
	big.NewInt(0).FillBytes(mockAddrBytes)
	mockAddrBytes[0] = 1
	mockAddr, _ := pubKeyConverter.Encode(mockAddrBytes)

	ids = []string{mockAddr}
	err = esClient.DoMultiGet(context.Background(), ids, indexerdata.AccountsIndex, true, genericResponse)
	require.Nil(t, err)
	require.True(t, genericResponse.Docs[0].Found)
}
