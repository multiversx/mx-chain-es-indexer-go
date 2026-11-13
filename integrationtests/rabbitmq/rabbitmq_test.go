//go:build integrationtests

package rabbitmq

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-es-indexer-go/client"
	"github.com/multiversx/mx-chain-es-indexer-go/client/logging"
	"github.com/multiversx/mx-chain-es-indexer-go/config"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataprocessor/factory"
	"github.com/multiversx/mx-chain-es-indexer-go/process/outputs"
	"github.com/multiversx/mx-chain-es-indexer-go/process/outputs/elastic"
	rabbitOutput "github.com/multiversx/mx-chain-es-indexer-go/process/outputs/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
)

const (
	esURL           = "http://localhost:9200"
	rabbitURL       = "amqp://guest:guest@localhost:5672"
	testNumOfShards = 3
	receiveTimeout  = 10 * time.Second
	sender          = "erd12m3x8jp6dl027pj5f2nw6ght2cyhhjfrs86cdwsa8xn83r375qfqrwpdx0"
	receiver        = "erd1v7e552pz9py4hv6raan0c4jflez3e6csdmzcgrncg0qrnk4tywvsqx0h5j"
)

var addressConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, "erd")

type blockMessage struct {
	ShardID    uint32                   `json:"shardID"`
	Hash       string                   `json:"hash"`
	Round      uint64                   `json:"round"`
	Operations []map[string]interface{} `json:"operations"`
	Events     []map[string]interface{} `json:"events"`
}

type revertMessage struct {
	ShardID    uint32   `json:"shardID"`
	Hash       string   `json:"hash"`
	Round      uint64   `json:"round"`
	Operations []string `json:"operations"`
}

type esResponse struct {
	Docs []struct {
		Found  bool                   `json:"found"`
		ID     string                 `json:"_id"`
		Source map[string]interface{} `json:"_source"`
	} `json:"docs"`
}

func decodeAddress(t *testing.T, address string) []byte {
	decoded, err := addressConverter.Decode(address)
	require.Nil(t, err)

	return decoded
}

func randomSuffix(t *testing.T) string {
	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	require.Nil(t, err)

	return hex.EncodeToString(randomBytes)
}

func createESClient(t *testing.T) elastic.DatabaseClientHandler {
	esClient, err := client.NewElasticClient(elasticsearch.Config{
		Addresses: []string{esURL},
		Logger:    &logging.CustomLogger{},
	})
	require.Nil(t, err)

	return esClient
}

// createDataProcessor creates the data processor with the Elasticsearch output and the RabbitMQ output
func createDataProcessor(
	t *testing.T,
	esClient elastic.DatabaseClientHandler,
	url string,
	blocksExchange string,
	revertsExchange string,
	continueOnError bool,
) dataindexer.DataProcessor {
	rabbitClient, err := rabbitOutput.NewClient(rabbitOutput.ArgsClient{
		URL: url,
		Exchanges: []rabbitOutput.ExchangeConfig{
			{Name: blocksExchange, Type: "fanout"},
			{Name: revertsExchange, Type: "fanout"},
		},
		ConnectionName:           "integration-tests",
		FailFast:                 continueOnError,
		RequireInitialConnection: !continueOnError,
	})
	require.Nil(t, err)

	output, err := rabbitOutput.NewOutput(rabbitOutput.ArgsOutput{
		Client:          rabbitClient,
		BlocksExchange:  blocksExchange,
		RevertsExchange: revertsExchange,
		PublishTimeout:  receiveTimeout,
	})
	require.Nil(t, err)

	dataProcessor, err := factory.CreateDataProcessor(factory.ArgDataProcessorFactory{
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		AddressPubkeyConverter:   addressConverter,
		ValidatorPubkeyConverter: mock.NewPubkeyConverterMock(32),
		DBClient:                 esClient,
		EnabledIndexes: []string{dataindexer.TransactionsIndex, dataindexer.LogsIndex, dataindexer.ScResultsIndex, dataindexer.BlockIndex,
			dataindexer.EventsIndex, dataindexer.OperationsIndex, dataindexer.MiniblocksIndex, dataindexer.ValuesIndex},
		Denomination: 18,
		EnableEpochsConfig: config.EnableEpochsConfig{
			RelayedTransactionsV1V2DisableEpoch: 1,
		},
		NumWritesInParallel: 1,
		ExtraOutputs: []outputs.OutputWithPolicy{
			{Name: "rabbitmq", Output: output, ContinueOnError: continueOnError},
		},
	})
	require.Nil(t, err)
	t.Cleanup(func() {
		_ = dataProcessor.Close()
	})

	return dataProcessor
}

// consume binds a new queue to the exchange and returns the channel with the messages published from now on
func consume(t *testing.T, exchange string) <-chan amqp.Delivery {
	conn, err := amqp.Dial(rabbitURL)
	require.Nil(t, err)
	t.Cleanup(func() {
		_ = conn.Close()
	})

	ch, err := conn.Channel()
	require.Nil(t, err)

	// the same exchange as the one declared by the indexer
	err = ch.ExchangeDeclare(exchange, "fanout", true, false, false, false, nil)
	require.Nil(t, err)

	queue, err := ch.QueueDeclare("", false, true, true, false, nil)
	require.Nil(t, err)

	err = ch.QueueBind(queue.Name, "", exchange, false, nil)
	require.Nil(t, err)

	deliveries, err := ch.Consume(queue.Name, "", true, true, false, false, nil)
	require.Nil(t, err)

	return deliveries
}

// deleteExchangesAtTheEnd removes the exchanges from the RabbitMQ server when the test is finished.
// It has to be called before the data processor and the consumers are created: the cleanup functions are called
// in the reverse order, so the exchanges are removed after the data processor and the consumers are closed
func deleteExchangesAtTheEnd(t *testing.T, exchanges ...string) {
	t.Cleanup(func() {
		conn, err := amqp.Dial(rabbitURL)
		require.Nil(t, err)
		defer func() {
			_ = conn.Close()
		}()

		ch, err := conn.Channel()
		require.Nil(t, err)

		for _, exchange := range exchanges {
			err = ch.ExchangeDelete(exchange, false, false)
			require.Nil(t, err)
		}
	})
}

func receive(t *testing.T, deliveries <-chan amqp.Delivery) amqp.Delivery {
	select {
	case delivery := <-deliveries:
		return delivery
	case <-time.After(receiveTimeout):
		require.Fail(t, "no message was received")
		return amqp.Delivery{}
	}
}

func requireNoMessage(t *testing.T, deliveries <-chan amqp.Delivery) {
	select {
	case delivery := <-deliveries:
		require.Fail(t, "a message was received", "message id: %s", delivery.MessageId)
	case <-time.After(time.Second):
	}
}

type testBlock struct {
	blockHash []byte
	txHash    []byte
	scrHash   []byte
	header    *dataBlock.Header
	body      *dataBlock.Body
	pool      *outport.TransactionPool
}

func (tb *testBlock) outportBlock() *outport.OutportBlockWithHeader {
	return &outport.OutportBlockWithHeader{
		Header: tb.header,
		OutportBlock: &outport.OutportBlock{
			BlockData: &outport.BlockData{
				HeaderHash:  tb.blockHash,
				Body:        tb.body,
				TimestampMs: tb.header.TimeStamp * 1000,
			},
			TransactionPool: tb.pool,
			NumberOfShards:  testNumOfShards,
			ShardID:         tb.header.ShardID,
		},
	}
}

// createTestBlock creates a block with a transaction, a smart contract result and a log with an event
func createTestBlock(t *testing.T, suffix string, round uint64) *testBlock {
	blockHash := []byte("block-" + suffix)
	txHash := []byte("tx-" + suffix)
	scrHash := []byte("scr-" + suffix)

	shardID := uint32(2)
	miniblockHash := []byte("mb-" + suffix)
	header := &dataBlock.Header{
		Round:     round,
		TimeStamp: round * 6,
		ShardID:   shardID,
		MiniBlockHeaders: []dataBlock.MiniBlockHeader{
			{Hash: miniblockHash, SenderShardID: shardID, ReceiverShardID: shardID, Type: dataBlock.TxBlock, TxCount: 1},
		},
	}
	body := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.TxBlock,
				SenderShardID:   shardID,
				ReceiverShardID: shardID,
				TxHashes:        [][]byte{txHash},
			},
		},
	}

	pool := &outport.TransactionPool{
		Transactions: map[string]*outport.TxInfo{
			hex.EncodeToString(txHash): {
				Transaction: &transaction.Transaction{
					Nonce:    1,
					SndAddr:  decodeAddress(t, sender),
					RcvAddr:  decodeAddress(t, receiver),
					GasLimit: 70000,
					GasPrice: 1000000000,
					Value:    big.NewInt(1000),
				},
				FeeInfo: &outport.FeeInfo{
					GasUsed:        70000,
					Fee:            big.NewInt(70000000000000),
					InitialPaidFee: big.NewInt(70000000000000),
				},
			},
		},
		SmartContractResults: map[string]*outport.SCRInfo{
			hex.EncodeToString(scrHash): {
				SmartContractResult: &smartContractResult.SmartContractResult{
					Nonce:          2,
					SndAddr:        decodeAddress(t, receiver),
					RcvAddr:        decodeAddress(t, sender),
					PrevTxHash:     txHash,
					OriginalTxHash: txHash,
					Value:          big.NewInt(10),
				},
				FeeInfo: &outport.FeeInfo{},
			},
		},
		Logs: []*transaction.LogData{
			{
				TxHash: hex.EncodeToString(txHash),
				Log: &transaction.Log{
					Address: decodeAddress(t, sender),
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(t, sender),
							Identifier: []byte("writeLog"),
							Topics:     [][]byte{[]byte("topic")},
						},
					},
				},
			},
		},
	}

	return &testBlock{
		blockHash: blockHash,
		txHash:    txHash,
		scrHash:   scrHash,
		header:    header,
		body:      body,
		pool:      pool,
	}
}

func getDocuments(t *testing.T, esClient elastic.DatabaseClientHandler, index string, ids []string) map[string]map[string]interface{} {
	response := &esResponse{}
	err := esClient.DoMultiGet(context.Background(), ids, index, true, response)
	require.Nil(t, err)

	documents := make(map[string]map[string]interface{})
	for _, doc := range response.Docs {
		require.True(t, doc.Found, "index %s, id %s", index, doc.ID)
		documents[doc.ID] = doc.Source
	}

	return documents
}

func TestBlockAndRevertMessages(t *testing.T) {
	suffix := randomSuffix(t)
	blocksExchange := "it_blocks_" + suffix
	revertsExchange := "it_reverts_" + suffix
	deleteExchangesAtTheEnd(t, blocksExchange, revertsExchange)

	esClient := createESClient(t)
	dataProcessor := createDataProcessor(t, esClient, rabbitURL, blocksExchange, revertsExchange, false)
	blocks := consume(t, blocksExchange)
	reverts := consume(t, revertsExchange)

	block := createTestBlock(t, suffix, 100)
	blockHashHex := hex.EncodeToString(block.blockHash)
	txHashHex := hex.EncodeToString(block.txHash)
	scrHashHex := hex.EncodeToString(block.scrHash)

	err := dataProcessor.SaveTransactions(block.outportBlock())
	require.Nil(t, err)

	// ################ ONE MESSAGE FOR THE BLOCK ##########################
	delivery := receive(t, blocks)
	require.Equal(t, blockHashHex, delivery.MessageId)
	require.Equal(t, "application/json", delivery.ContentType)
	require.Equal(t, amqp.Persistent, delivery.DeliveryMode)

	message := &blockMessage{}
	require.Nil(t, json.Unmarshal(delivery.Body, message))
	require.Equal(t, uint32(2), message.ShardID)
	require.Equal(t, blockHashHex, message.Hash)
	require.Equal(t, uint64(100), message.Round)
	require.Len(t, message.Operations, 2)
	require.Len(t, message.Events, 1)

	// the operations from the message are the documents from the operations index, with the hash
	operations := getDocuments(t, esClient, dataindexer.OperationsIndex, []string{txHashHex, scrHashHex})
	require.Equal(t, txHashHex, message.Operations[0]["hash"])
	require.Equal(t, "normal", message.Operations[0]["type"])
	require.Equal(t, scrHashHex, message.Operations[1]["hash"])
	require.Equal(t, "unsigned", message.Operations[1]["type"])
	for _, operation := range message.Operations {
		hash := operation["hash"].(string)
		expected := operations[hash]
		expected["hash"] = hash
		require.Equal(t, expected, operation)
	}

	// the events from the message are the documents from the events index, with the id
	eventID := fmt.Sprintf("%s-2-0", txHashHex)
	events := getDocuments(t, esClient, dataindexer.EventsIndex, []string{eventID})
	expectedEvent := events[eventID]
	expectedEvent["id"] = eventID
	require.Equal(t, expectedEvent, message.Events[0])
	require.Equal(t, "writeLog", message.Events[0]["identifier"])

	// ################ THE SAME BLOCK IS RECEIVED AGAIN ##########################
	// the block was saved by all the outputs, nothing is published
	err = dataProcessor.SaveTransactions(block.outportBlock())
	require.Nil(t, err)
	requireNoMessage(t, blocks)

	// ################ REVERT ##########################
	err = dataProcessor.RevertBlock(block.header, &outport.BlockData{
		HeaderHash:  block.blockHash,
		Body:        block.body,
		TimestampMs: block.header.TimeStamp * 1000,
	})
	require.Nil(t, err)

	delivery = receive(t, reverts)
	require.Equal(t, blockHashHex+":revert", delivery.MessageId)
	require.Equal(t, amqp.Persistent, delivery.DeliveryMode)

	revert := &revertMessage{}
	require.Nil(t, json.Unmarshal(delivery.Body, revert))
	require.Equal(t, uint32(2), revert.ShardID)
	require.Equal(t, blockHashHex, revert.Hash)
	require.Equal(t, uint64(100), revert.Round)
	require.Equal(t, []string{txHashHex}, revert.Operations)
	requireNoMessage(t, blocks)

	// ################ THE BLOCK IS RECEIVED AGAIN AFTER THE REVERT ##########################
	err = dataProcessor.SaveTransactions(block.outportBlock())
	require.Nil(t, err)

	delivery = receive(t, blocks)
	require.Equal(t, blockHashHex, delivery.MessageId)
}

func TestBlockWithoutOperationsAndEventsIsPublished(t *testing.T) {
	suffix := randomSuffix(t)
	blocksExchange := "it_blocks_" + suffix
	revertsExchange := "it_reverts_" + suffix
	deleteExchangesAtTheEnd(t, blocksExchange, revertsExchange)

	esClient := createESClient(t)
	dataProcessor := createDataProcessor(t, esClient, rabbitURL, blocksExchange, revertsExchange, false)
	blocks := consume(t, blocksExchange)

	emptyBlock := &outport.OutportBlockWithHeader{
		Header: &dataBlock.Header{Round: 101, TimeStamp: 606, ShardID: 2},
		OutportBlock: &outport.OutportBlock{
			BlockData:       &outport.BlockData{HeaderHash: []byte("empty-" + suffix), Body: &dataBlock.Body{}, TimestampMs: 606000},
			TransactionPool: &outport.TransactionPool{},
			NumberOfShards:  testNumOfShards,
			ShardID:         2,
		},
	}
	err := dataProcessor.SaveTransactions(emptyBlock)
	require.Nil(t, err)

	// the message has the information about the block and the lists are empty
	emptyBlockHashHex := hex.EncodeToString([]byte("empty-" + suffix))
	delivery := receive(t, blocks)
	require.Equal(t, emptyBlockHashHex, delivery.MessageId)
	require.JSONEq(t, `{"shardID":2,"hash":"`+emptyBlockHashHex+`","round":101,"operations":[],"events":[]}`, string(delivery.Body))

	// the next block that has data is published
	block := createTestBlock(t, suffix, 102)
	err = dataProcessor.SaveTransactions(block.outportBlock())
	require.Nil(t, err)

	delivery = receive(t, blocks)
	require.Equal(t, hex.EncodeToString(block.blockHash), delivery.MessageId)
	requireNoMessage(t, blocks)
}

func TestHeaderV3OneMessageForEveryExecutedBlock(t *testing.T) {
	suffix := randomSuffix(t)
	blocksExchange := "it_blocks_" + suffix
	revertsExchange := "it_reverts_" + suffix
	deleteExchangesAtTheEnd(t, blocksExchange, revertsExchange)

	esClient := createESClient(t)
	dataProcessor := createDataProcessor(t, esClient, rabbitURL, blocksExchange, revertsExchange, false)
	blocks := consume(t, blocksExchange)

	firstExecuted := createTestBlock(t, "first-"+suffix, 200)
	secondExecuted := createTestBlock(t, "second-"+suffix, 201)
	// the third executed block has no operations and no events
	emptyExecutedHash := []byte("emptyExecuted-" + suffix)

	// the header notarizes the execution results of three blocks, the block of the header itself is not executed yet
	header := &dataBlock.HeaderV3{
		ShardID:     2,
		Round:       203,
		TimestampMs: 1218000,
		ExecutionResults: []*dataBlock.ExecutionResult{
			{
				BaseExecutionResult: &dataBlock.BaseExecutionResult{HeaderHash: firstExecuted.blockHash, HeaderNonce: 1, HeaderRound: 200},
				AccumulatedFees:     big.NewInt(0),
				DeveloperFees:       big.NewInt(0),
			},
			{
				BaseExecutionResult: &dataBlock.BaseExecutionResult{HeaderHash: secondExecuted.blockHash, HeaderNonce: 2, HeaderRound: 201},
				AccumulatedFees:     big.NewInt(0),
				DeveloperFees:       big.NewInt(0),
			},
			{
				BaseExecutionResult: &dataBlock.BaseExecutionResult{HeaderHash: emptyExecutedHash, HeaderNonce: 3, HeaderRound: 202},
				AccumulatedFees:     big.NewInt(0),
				DeveloperFees:       big.NewInt(0),
			},
		},
	}
	obh := &outport.OutportBlockWithHeader{
		Header: header,
		OutportBlock: &outport.OutportBlock{
			BlockData: &outport.BlockData{
				HeaderHash:  []byte("header-" + suffix),
				Body:        &dataBlock.Body{},
				TimestampMs: 1218000,
				Results: map[string]*outport.ExecutionResultData{
					hex.EncodeToString(firstExecuted.blockHash): {
						HeaderNonce:     1,
						TimestampMs:     1200000,
						Body:            firstExecuted.body,
						TransactionPool: firstExecuted.pool,
					},
					hex.EncodeToString(secondExecuted.blockHash): {
						HeaderNonce:     2,
						TimestampMs:     1206000,
						Body:            secondExecuted.body,
						TransactionPool: secondExecuted.pool,
					},
					hex.EncodeToString(emptyExecutedHash): {
						HeaderNonce:     3,
						TimestampMs:     1212000,
						Body:            &dataBlock.Body{},
						TransactionPool: &outport.TransactionPool{},
					},
				},
			},
			TransactionPool: &outport.TransactionPool{},
			NumberOfShards:  testNumOfShards,
			ShardID:         2,
		},
	}

	err := dataProcessor.SaveTransactions(obh)
	require.Nil(t, err)

	// one message for every executed block, with the hash and the round of the executed block
	for _, expected := range []struct {
		block *testBlock
		round uint64
	}{{firstExecuted, 200}, {secondExecuted, 201}} {
		delivery := receive(t, blocks)
		require.Equal(t, hex.EncodeToString(expected.block.blockHash), delivery.MessageId)

		message := &blockMessage{}
		require.Nil(t, json.Unmarshal(delivery.Body, message))
		require.Equal(t, hex.EncodeToString(expected.block.blockHash), message.Hash)
		require.Equal(t, expected.round, message.Round)
		require.Equal(t, uint32(2), message.ShardID)
		require.NotEmpty(t, message.Operations)
		require.Equal(t, hex.EncodeToString(expected.block.txHash), message.Operations[0]["hash"])
	}

	// the executed block without operations and without events is published with empty lists
	emptyExecutedHashHex := hex.EncodeToString(emptyExecutedHash)
	delivery := receive(t, blocks)
	require.Equal(t, emptyExecutedHashHex, delivery.MessageId)
	require.JSONEq(t, `{"shardID":2,"hash":"`+emptyExecutedHashHex+`","round":202,"operations":[],"events":[]}`, string(delivery.Body))

	// there is no message for the block of the header: it will be published when its execution result is received
	requireNoMessage(t, blocks)
}

func TestServerCannotBeReached(t *testing.T) {
	suffix := randomSuffix(t)
	// nothing listens on this port
	unreachableURL := "amqp://guest:guest@127.0.0.1:1"

	esClient := createESClient(t)
	// the errors are only logged: the block is saved in Elasticsearch and the indexing continues
	dataProcessor := createDataProcessor(t, esClient, unreachableURL, "it_blocks_"+suffix, "it_reverts_"+suffix, true)

	block := createTestBlock(t, suffix, 300)
	startTime := time.Now()
	err := dataProcessor.SaveTransactions(block.outportBlock())
	require.Nil(t, err)
	require.Less(t, time.Since(startTime), receiveTimeout)

	operations := getDocuments(t, esClient, dataindexer.OperationsIndex, []string{hex.EncodeToString(block.txHash)})
	require.Len(t, operations, 1)
}
