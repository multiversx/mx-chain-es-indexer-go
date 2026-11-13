package rabbitmq

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"
	"time"

	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/metrics"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/stretchr/testify/require"
)

type publishedMessage struct {
	exchange  string
	messageID string
	body      []byte
}

func createMockArgsOutput() ArgsOutput {
	return ArgsOutput{
		Client:          &mock.RabbitClientStub{},
		BlocksExchange:  "blocks",
		RevertsExchange: "reverts",
		PublishTimeout:  time.Second,
	}
}

func createPreparedBlock(units ...*data.PreparedBlockUnit) *data.PreparedBlock {
	return &data.PreparedBlock{
		OutportBlock: &outport.OutportBlockWithHeader{
			Header:       &dataBlock.Header{ShardID: 1, Round: 50},
			OutportBlock: &outport.OutportBlock{BlockData: &outport.BlockData{}},
		},
		Units: units,
	}
}

func TestNewOutput(t *testing.T) {
	t.Parallel()

	t.Run("nil client", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsOutput()
		args.Client = nil
		output, err := NewOutput(args)
		require.Nil(t, output)
		require.Equal(t, ErrNilClient, err)
	})
	t.Run("empty blocks exchange", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsOutput()
		args.BlocksExchange = ""
		output, err := NewOutput(args)
		require.Nil(t, output)
		require.Equal(t, ErrInvalidExchangeName, err)
	})
	t.Run("empty reverts exchange", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsOutput()
		args.RevertsExchange = ""
		output, err := NewOutput(args)
		require.Nil(t, output)
		require.Equal(t, ErrInvalidExchangeName, err)
	})
	t.Run("invalid publish timeout", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsOutput()
		args.PublishTimeout = 0
		output, err := NewOutput(args)
		require.Nil(t, output)
		require.Equal(t, ErrInvalidPublishTimeout, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		output, err := NewOutput(createMockArgsOutput())
		require.Nil(t, err)
		require.False(t, output.IsInterfaceNil())
	})
}

func TestRabbitOutput_SaveTransactions(t *testing.T) {
	t.Parallel()

	published := make([]publishedMessage, 0)
	args := createMockArgsOutput()
	args.Client = &mock.RabbitClientStub{
		PublishCalled: func(ctx context.Context, exchange string, messageID string, body []byte) error {
			// all the messages of a block have to be published before the timeout
			_, hasDeadline := ctx.Deadline()
			require.True(t, hasDeadline)

			published = append(published, publishedMessage{exchange: exchange, messageID: messageID, body: body})
			return nil
		},
	}
	statusMetrics := metrics.NewStatusMetrics()
	args.StatusMetrics = statusMetrics
	output, _ := NewOutput(args)

	// the block of the header v3 is not executed yet, the data comes from three executed blocks, the last one has no data
	headerUnit := &data.PreparedBlockUnit{
		BlockHash:  []byte("header"),
		Executed:   false,
		HeaderData: &data.HeaderData{ShardID: 1, Round: 53},
		Logs:       &data.PreparedLogsResults{},
	}
	firstUnit := createBlockUnit()
	firstUnit.BlockHash = []byte("first")
	secondUnit := createBlockUnit()
	secondUnit.BlockHash = []byte("second")
	secondUnit.HeaderData = &data.HeaderData{ShardID: 1, Round: 51}
	thirdUnit := &data.PreparedBlockUnit{
		BlockHash:  []byte("third"),
		Executed:   true,
		HeaderData: &data.HeaderData{ShardID: 1, Round: 52},
		Logs:       &data.PreparedLogsResults{},
	}

	err := output.SaveTransactions(createPreparedBlock(headerUnit, firstUnit, secondUnit, thirdUnit))
	require.Nil(t, err)

	// one message for every executed block, even if it has no operations and no events, the message id is the hash of the block
	require.Len(t, published, 3)
	require.Equal(t, "blocks", published[0].exchange)
	require.Equal(t, hex.EncodeToString([]byte("first")), published[0].messageID)
	require.Equal(t, "blocks", published[1].exchange)
	require.Equal(t, hex.EncodeToString([]byte("second")), published[1].messageID)
	require.Equal(t, "blocks", published[2].exchange)
	require.Equal(t, hex.EncodeToString([]byte("third")), published[2].messageID)
	require.Equal(t,
		`{"shardID":1,"hash":"`+hex.EncodeToString([]byte("third"))+`","round":52,"operations":[],"events":[]}`,
		string(published[2].body),
	)

	message := &struct {
		ShardID    uint32                   `json:"shardID"`
		Hash       string                   `json:"hash"`
		Round      uint64                   `json:"round"`
		Operations []map[string]interface{} `json:"operations"`
		Events     []map[string]interface{} `json:"events"`
	}{}
	require.Nil(t, json.Unmarshal(published[1].body, message))
	require.Equal(t, uint32(1), message.ShardID)
	require.Equal(t, hex.EncodeToString([]byte("second")), message.Hash)
	require.Equal(t, uint64(51), message.Round)
	require.Len(t, message.Operations, 2)
	require.Equal(t, "txHash", message.Operations[0]["hash"])
	require.Equal(t, "scrHash", message.Operations[1]["hash"])
	require.Len(t, message.Events, 1)
	require.Equal(t, "txHash-1-0", message.Events[0]["id"])

	blockMetrics := statusMetrics.GetMetrics()["rabbit_block_1"]
	require.NotNil(t, blockMetrics)
	require.Equal(t, uint64(3), blockMetrics.OperationsCount)
	require.Equal(t, uint64(0), blockMetrics.TotalErrorsCount)
	require.Equal(t, uint64(len(published[0].body)+len(published[1].body)+len(published[2].body)), blockMetrics.TotalData)
}

func TestRabbitOutput_SaveTransactionsBlockWithoutOperationsAndEventsIsPublished(t *testing.T) {
	t.Parallel()

	published := make([]publishedMessage, 0)
	args := createMockArgsOutput()
	args.Client = &mock.RabbitClientStub{
		PublishCalled: func(_ context.Context, exchange string, messageID string, body []byte) error {
			published = append(published, publishedMessage{exchange: exchange, messageID: messageID, body: body})
			return nil
		},
	}
	output, _ := NewOutput(args)

	emptyUnit := &data.PreparedBlockUnit{
		BlockHash:  []byte("empty"),
		Executed:   true,
		HeaderData: &data.HeaderData{ShardID: 2, Round: 60},
	}
	err := output.SaveTransactions(createPreparedBlock(emptyUnit))
	require.Nil(t, err)

	require.Len(t, published, 1)
	require.Equal(t, "blocks", published[0].exchange)
	require.Equal(t, hex.EncodeToString([]byte("empty")), published[0].messageID)
	require.Equal(t,
		`{"shardID":2,"hash":"`+hex.EncodeToString([]byte("empty"))+`","round":60,"operations":[],"events":[]}`,
		string(published[0].body),
	)
}

func TestRabbitOutput_SaveTransactionsBlockNotExecutedIsNotPublished(t *testing.T) {
	t.Parallel()

	args := createMockArgsOutput()
	args.Client = &mock.RabbitClientStub{
		PublishCalled: func(_ context.Context, _ string, _ string, _ []byte) error {
			require.Fail(t, "should have not been called")
			return nil
		},
	}
	output, _ := NewOutput(args)

	// the block of a header v3: it will be published when its execution result is received
	headerUnit := &data.PreparedBlockUnit{
		BlockHash:  []byte("header"),
		Executed:   false,
		HeaderData: &data.HeaderData{ShardID: 1, Round: 52},
		Logs:       &data.PreparedLogsResults{},
	}
	err := output.SaveTransactions(createPreparedBlock(headerUnit))
	require.Nil(t, err)
}

func TestRabbitOutput_SaveTransactionsPublishError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("publish error")
	publishCalls := 0
	args := createMockArgsOutput()
	args.Client = &mock.RabbitClientStub{
		PublishCalled: func(_ context.Context, _ string, _ string, _ []byte) error {
			publishCalls++
			return expectedErr
		},
	}
	statusMetrics := metrics.NewStatusMetrics()
	args.StatusMetrics = statusMetrics
	output, _ := NewOutput(args)

	err := output.SaveTransactions(createPreparedBlock(createBlockUnit(), createBlockUnit()))
	require.True(t, errors.Is(err, expectedErr))
	require.Contains(t, err.Error(), hex.EncodeToString([]byte("blockHash")))
	// the next blocks are not published after an error
	require.Equal(t, 1, publishCalls)
	require.Equal(t, uint64(1), statusMetrics.GetMetrics()["rabbit_block_1"].TotalErrorsCount)
}

func TestRabbitOutput_RevertBlock(t *testing.T) {
	t.Parallel()

	published := make([]publishedMessage, 0)
	args := createMockArgsOutput()
	args.Client = &mock.RabbitClientStub{
		PublishCalled: func(_ context.Context, exchange string, messageID string, body []byte) error {
			published = append(published, publishedMessage{exchange: exchange, messageID: messageID, body: body})
			return nil
		},
	}
	output, _ := NewOutput(args)

	err := output.RevertBlock(&data.RevertedBlock{
		Header:    &dataBlock.Header{ShardID: 2, Round: 77},
		BlockData: &outport.BlockData{HeaderHash: []byte("blockHash")},
		TxHashes:  []string{"tx1"},
		ScrHashes: []string{"scr1"},
	})
	require.Nil(t, err)

	blockHash := hex.EncodeToString([]byte("blockHash"))
	require.Len(t, published, 1)
	require.Equal(t, "reverts", published[0].exchange)
	require.Equal(t, blockHash+":revert", published[0].messageID)
	require.JSONEq(t, `{"shardID":2,"hash":"`+blockHash+`","round":77,"operations":["tx1","scr1"]}`, string(published[0].body))
}

func TestRabbitOutput_RevertBlockPublishError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("publish error")
	args := createMockArgsOutput()
	args.Client = &mock.RabbitClientStub{
		PublishCalled: func(_ context.Context, _ string, _ string, _ []byte) error {
			return expectedErr
		},
	}
	output, _ := NewOutput(args)

	err := output.RevertBlock(&data.RevertedBlock{
		Header:    &dataBlock.Header{ShardID: 2, Round: 77},
		BlockData: &outport.BlockData{HeaderHash: []byte("blockHash")},
	})
	require.True(t, errors.Is(err, expectedErr))
}

func TestRabbitOutput_CloseShouldInterruptThePublishAndCloseTheClient(t *testing.T) {
	t.Parallel()

	publishStarted := make(chan struct{})
	closeCalls := 0
	args := createMockArgsOutput()
	args.PublishTimeout = time.Minute
	args.Client = &mock.RabbitClientStub{
		PublishCalled: func(ctx context.Context, _ string, _ string, _ []byte) error {
			close(publishStarted)
			<-ctx.Done()
			return ctx.Err()
		},
		CloseCalled: func() error {
			closeCalls++
			return nil
		},
	}
	output, _ := NewOutput(args)

	errCh := make(chan error, 1)
	go func() {
		errCh <- output.SaveTransactions(createPreparedBlock(createBlockUnit()))
	}()

	<-publishStarted
	require.Nil(t, output.Close())

	select {
	case err := <-errCh:
		require.True(t, errors.Is(err, context.Canceled))
	case <-time.After(5 * time.Second):
		require.Fail(t, "the publish was not interrupted")
	}
	require.Equal(t, 1, closeCalls)
}

func TestRabbitOutput_TheOtherCallsDoNothing(t *testing.T) {
	t.Parallel()

	args := createMockArgsOutput()
	args.Client = &mock.RabbitClientStub{
		PublishCalled: func(_ context.Context, _ string, _ string, _ []byte) error {
			require.Fail(t, "should have not been called")
			return nil
		},
	}
	output, _ := NewOutput(args)

	require.Nil(t, output.SaveHeader(&outport.OutportBlockWithHeader{}))
	require.Nil(t, output.SaveMiniblocks(&outport.OutportBlockWithHeader{}))
	require.Nil(t, output.SaveValidatorsRating(&outport.ValidatorsRating{}))
	require.Nil(t, output.SaveRoundsInfo(&outport.RoundsInfo{}))
	require.Nil(t, output.SaveShardValidatorsPubKeys(&outport.ValidatorsPubKeys{}))
	require.Nil(t, output.SaveAccounts(&outport.Accounts{}))
}
