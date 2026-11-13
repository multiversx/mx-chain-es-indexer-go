package rabbitmq

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/stretchr/testify/require"
)

func createBlockUnit() *data.PreparedBlockUnit {
	return &data.PreparedBlockUnit{
		BlockHash:  []byte("blockHash"),
		Executed:   true,
		HeaderData: &data.HeaderData{ShardID: 1, Round: 50},
		OperationsTxs: []*data.Transaction{
			{
				Hash:        "txHash",
				Nonce:       1,
				Round:       50,
				Value:       "100",
				Sender:      "erd1sender",
				Receiver:    "erd1receiver",
				Status:      "success",
				Type:        "normal",
				Timestamp:   5000,
				TimestampMs: 5000000,
			},
		},
		OperationsSCRs: []*data.ScResult{
			{
				Hash:           "scrHash",
				Nonce:          2,
				Value:          "0",
				Sender:         "erd1sender",
				Receiver:       "erd1receiver",
				PrevTxHash:     "txHash",
				OriginalTxHash: "txHash",
				CallType:       "0",
				Type:           "unsigned",
				Status:         "success",
				Timestamp:      5000,
				TimestampMs:    5000000,
			},
		},
		Logs: &data.PreparedLogsResults{
			DBEvents: []*data.LogEvent{
				{
					ID:          "txHash-1-0",
					TxHash:      "txHash",
					LogAddress:  "erd1log",
					Address:     "erd1addr",
					Identifier:  "transfer",
					Topics:      []string{"01"},
					ShardID:     1,
					Timestamp:   5000,
					TimestampMs: 5000000,
				},
			},
		},
	}
}

func TestNewBlockMessage(t *testing.T) {
	t.Parallel()

	message := newBlockMessage(createBlockUnit())
	require.NotNil(t, message)

	messageBytes, err := json.Marshal(message)
	require.Nil(t, err)

	// the block info, the operations and the events are in the same message.
	// An operation is the document from the operations index and its hash, an event is the document from the events index and its id
	expected := `{
	"shardID": 1,
	"hash": "` + hex.EncodeToString([]byte("blockHash")) + `",
	"round": 50,
	"operations": [
		{"hash":"txHash","uuid":"","miniBlockHash":"","nonce":1,"round":50,"value":"100","valueNum":0,
		 "receiver":"erd1receiver","sender":"erd1sender","receiverShard":0,"senderShard":0,"gasPrice":0,"gasLimit":0,"gasUsed":0,
		 "fee":"","feeNum":0,"data":null,"signature":"","timestamp":5000,"timestampMs":5000000,"status":"success","searchOrder":0,
		 "type":"normal","epoch":0},
		{"hash":"scrHash","uuid":"","nonce":2,"gasLimit":0,"gasPrice":0,"value":"0","valueNum":0,
		 "sender":"erd1sender","receiver":"erd1receiver","senderShard":0,"receiverShard":0,"prevTxHash":"txHash","originalTxHash":"txHash",
		 "callType":"0","timestamp":5000,"timestampMs":5000000,"type":"unsigned","status":"success","epoch":0}
	],
	"events": [
		{"id":"txHash-1-0","uuid":"","txHash":"txHash","logAddress":"erd1log","address":"erd1addr","identifier":"transfer",
		 "topics":["01"],"order":0,"txOrder":0,"shardID":1,"timestamp":5000,"timestampMs":5000000}
	]
}`
	require.JSONEq(t, expected, string(messageBytes))
}

func TestNewBlockMessage_EveryFieldOfTheDocumentsIsInTheMessage(t *testing.T) {
	t.Parallel()

	toMap := func(value interface{}) map[string]interface{} {
		valueBytes, err := json.Marshal(value)
		require.Nil(t, err)

		result := make(map[string]interface{})
		require.Nil(t, json.Unmarshal(valueBytes, &result))

		return result
	}

	unit := createBlockUnit()
	// set also some of the fields that are omitted when they are empty
	unit.OperationsTxs[0].Tokens = []string{"TKN-abcdef"}
	unit.OperationsTxs[0].IsRelayed = true
	unit.OperationsTxs[0].Data = []byte("data")
	unit.OperationsSCRs[0].ReturnMessage = "message"
	unit.OperationsSCRs[0].Data = []byte("data")
	unit.Logs.DBEvents[0].Data = "0a"
	unit.Logs.DBEvents[0].OriginalTxHash = "originalTxHash"

	message := toMap(newBlockMessage(unit))
	operations := message["operations"].([]interface{})
	events := message["events"].([]interface{})

	// the message has the same fields as the document that is indexed, plus the hash or the id
	expectedTx := toMap(unit.OperationsTxs[0])
	expectedTx["hash"] = "txHash"
	require.Equal(t, expectedTx, operations[0])

	expectedSCR := toMap(unit.OperationsSCRs[0])
	expectedSCR["hash"] = "scrHash"
	require.Equal(t, expectedSCR, operations[1])

	expectedEvent := toMap(unit.Logs.DBEvents[0])
	expectedEvent["id"] = "txHash-1-0"
	require.Equal(t, expectedEvent, events[0])
}

func TestNewBlockMessage_NoOperationsAndNoEvents(t *testing.T) {
	t.Parallel()

	// the message has the information about the block and the lists are empty, they are not null
	expectedMessage := `{"shardID":1,"hash":"` + hex.EncodeToString([]byte("blockHash")) + `","round":50,"operations":[],"events":[]}`

	unit := &data.PreparedBlockUnit{
		BlockHash:  []byte("blockHash"),
		Executed:   true,
		HeaderData: &data.HeaderData{ShardID: 1, Round: 50},
	}
	message := newBlockMessage(unit)
	require.True(t, message.isEmpty())
	messageBytes, err := json.Marshal(message)
	require.Nil(t, err)
	require.Equal(t, expectedMessage, string(messageBytes))

	unit.Logs = &data.PreparedLogsResults{}
	message = newBlockMessage(unit)
	require.True(t, message.isEmpty())
	messageBytes, err = json.Marshal(message)
	require.Nil(t, err)
	require.Equal(t, expectedMessage, string(messageBytes))

	require.False(t, newBlockMessage(createBlockUnit()).isEmpty())
}

func TestNewBlockMessage_OnlyEvents(t *testing.T) {
	t.Parallel()

	unit := createBlockUnit()
	unit.OperationsTxs = nil
	unit.OperationsSCRs = nil

	message := newBlockMessage(unit)
	require.NotNil(t, message)
	require.Len(t, message.Events, 1)

	messageBytes, err := json.Marshal(message)
	require.Nil(t, err)
	// the list of operations is empty, it is not null
	require.Contains(t, string(messageBytes), `"operations":[]`)
}

func TestNewBlockMessage_OperationsAreSorted(t *testing.T) {
	t.Parallel()

	unit := createBlockUnit()
	unit.OperationsTxs = []*data.Transaction{
		{Hash: "c", ExecutionOrder: 1},
		{Hash: "b", ExecutionOrder: 0},
		{Hash: "a", ExecutionOrder: 1},
	}
	unit.OperationsSCRs = []*data.ScResult{
		{Hash: "z", ExecutionOrder: 3},
		{Hash: "y", ExecutionOrder: 2},
	}

	message := newBlockMessage(unit)
	hashes := make([]string, 0)
	for _, operation := range message.Operations {
		switch op := operation.(type) {
		case *transactionWithHash:
			hashes = append(hashes, op.Hash)
		case *scResultWithHash:
			hashes = append(hashes, op.Hash)
		}
	}

	// the transactions are first, then the smart contract results, both sorted by the execution order and by the hash
	require.Equal(t, []string{"b", "a", "c", "y", "z"}, hashes)
	// the prepared data is not changed
	require.Equal(t, "c", unit.OperationsTxs[0].Hash)
}

func TestNewBlockMessage_StatusFromTheLogsOfTheSameBlockIsApplied(t *testing.T) {
	t.Parallel()

	unit := createBlockUnit()
	unit.Logs.TxHashStatusInfo = map[string]*outport.StatusInfo{
		"txHash": {Status: "fail", ErrorEvent: true, CompletedEvent: true},
		// the status of a transaction from another block is not published
		"txFromAnotherBlock": {Status: "fail"},
	}

	message := newBlockMessage(unit)
	require.Len(t, message.Operations, 2)

	tx := message.Operations[0].(*transactionWithHash)
	require.Equal(t, "fail", tx.Status)
	require.True(t, tx.ErrorEvent)
	require.True(t, tx.CompletedEvent)

	// the prepared transaction is not changed, the other outputs use it
	require.Equal(t, "success", unit.OperationsTxs[0].Status)
	require.False(t, unit.OperationsTxs[0].ErrorEvent)
}

func TestNewRevertMessage(t *testing.T) {
	t.Parallel()

	message := newRevertMessage(&data.RevertedBlock{
		Header:    &dataBlock.Header{ShardID: 2, Round: 77},
		BlockData: &outport.BlockData{HeaderHash: []byte("blockHash")},
		TxHashes:  []string{"tx1", "tx2"},
		ScrHashes: []string{"scr1"},
	})

	messageBytes, err := json.Marshal(message)
	require.Nil(t, err)
	require.JSONEq(t, `{"shardID":2,"hash":"`+hex.EncodeToString([]byte("blockHash"))+`","round":77,"operations":["tx1","tx2","scr1"]}`, string(messageBytes))

	// a block without operations has an empty list
	message = newRevertMessage(&data.RevertedBlock{
		Header:    &dataBlock.Header{ShardID: 2, Round: 77},
		BlockData: &outport.BlockData{HeaderHash: []byte("blockHash")},
	})
	messageBytes, err = json.Marshal(message)
	require.Nil(t, err)
	require.Contains(t, string(messageBytes), `"operations":[]`)
}
