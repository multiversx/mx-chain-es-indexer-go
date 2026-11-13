package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	indexerCore "github.com/multiversx/mx-chain-es-indexer-go/core"
	"github.com/multiversx/mx-chain-es-indexer-go/core/request"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/metrics"
	"github.com/multiversx/mx-chain-es-indexer-go/process/outputs"
)

const revertMessageIDSuffix = ":revert"

// ArgsOutput holds the arguments needed to create the RabbitMQ output
type ArgsOutput struct {
	Client Client
	// BlocksExchange is the exchange where a message with the operations and the events is published for every block
	BlocksExchange string
	// RevertsExchange is the exchange where a message is published for every reverted block
	RevertsExchange string
	// PublishTimeout is the maximum time to wait for all the messages of a block to be published and confirmed
	PublishTimeout time.Duration
	// StatusMetrics is optional
	StatusMetrics indexerCore.StatusMetricsHandler
}

// rabbitOutput publishes the operations and the events of every block to RabbitMQ
type rabbitOutput struct {
	outputs.BaseOutput

	client          Client
	blocksExchange  string
	revertsExchange string
	publishTimeout  time.Duration
	statusMetrics   indexerCore.StatusMetricsHandler

	ctx    context.Context
	cancel context.CancelFunc
}

// NewOutput creates the output that publishes the operations and the events of every block to RabbitMQ
func NewOutput(args ArgsOutput) (*rabbitOutput, error) {
	if check.IfNil(args.Client) {
		return nil, ErrNilClient
	}
	if args.BlocksExchange == "" || args.RevertsExchange == "" {
		return nil, ErrInvalidExchangeName
	}
	if args.PublishTimeout <= 0 {
		return nil, ErrInvalidPublishTimeout
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &rabbitOutput{
		client:          args.Client,
		blocksExchange:  args.BlocksExchange,
		revertsExchange: args.RevertsExchange,
		publishTimeout:  args.PublishTimeout,
		statusMetrics:   args.StatusMetrics,
		ctx:             ctx,
		cancel:          cancel,
	}, nil
}

// SaveTransactions publishes a message with the operations and the events of every executed block from the provided data.
// A block without operations and without events is published too, with empty lists.
// The block of a header v3 is not executed yet: it is published when its execution result is received
func (ro *rabbitOutput) SaveTransactions(block *data.PreparedBlock) error {
	// the timeout is for all the messages of the block
	ctx, cancel := context.WithTimeout(ro.ctx, ro.publishTimeout)
	defer cancel()

	for _, unit := range block.Units {
		message := newBlockMessage(unit)
		if !unit.Executed && message.isEmpty() {
			continue
		}

		body, err := json.Marshal(message)
		if err != nil {
			return err
		}

		err = ro.publish(ctx, ro.blocksExchange, message.Hash, body, request.RabbitBlockTopic, message.ShardID)
		if err != nil {
			return fmt.Errorf("%w when publishing the block with hash %s", err, message.Hash)
		}
	}

	return nil
}

// RevertBlock publishes a message with the hashes of the operations of the reverted block
func (ro *rabbitOutput) RevertBlock(revertedBlock *data.RevertedBlock) error {
	ctx, cancel := context.WithTimeout(ro.ctx, ro.publishTimeout)
	defer cancel()

	message := newRevertMessage(revertedBlock)
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = ro.publish(ctx, ro.revertsExchange, message.Hash+revertMessageIDSuffix, body, request.RabbitRevertTopic, message.ShardID)
	if err != nil {
		return fmt.Errorf("%w when publishing the revert of the block with hash %s", err, message.Hash)
	}

	return nil
}

func (ro *rabbitOutput) publish(ctx context.Context, exchange string, messageID string, body []byte, topic string, shardID uint32) error {
	startTime := time.Now()
	err := ro.client.Publish(ctx, exchange, messageID, body)

	if !check.IfNil(ro.statusMetrics) {
		ro.statusMetrics.AddIndexingData(metrics.ArgsAddIndexingData{
			GotError:   err != nil,
			MessageLen: uint64(len(body)),
			Topic:      request.ExtendTopicWithShardID(topic, shardID),
			Duration:   time.Since(startTime),
		})
	}

	return err
}

// Close interrupts the publish that is in progress and closes the client
func (ro *rabbitOutput) Close() error {
	ro.cancel()

	return ro.client.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ro *rabbitOutput) IsInterfaceNil() bool {
	return ro == nil
}
