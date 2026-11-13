package rabbitmq

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const abortTimeout = time.Second

type connectionAdapter struct {
	conn *amqp.Connection
}

func dialAMQP(url string, dialTimeout time.Duration, connectionName string) (amqpConnection, error) {
	properties := amqp.NewConnectionProperties()
	properties.SetClientConnectionName(connectionName)

	conn, err := amqp.DialConfig(url, amqp.Config{
		Dial:       amqp.DefaultDial(dialTimeout),
		Properties: properties,
	})
	if err != nil {
		return nil, err
	}

	return &connectionAdapter{conn: conn}, nil
}

// Channel opens a new channel
func (ca *connectionAdapter) Channel() (amqpChannel, error) {
	ch, err := ca.conn.Channel()
	if err != nil {
		return nil, err
	}

	return &channelAdapter{ch: ch}, nil
}

// NotifyClose registers the receiver for the close event of the connection
func (ca *connectionAdapter) NotifyClose(receiver chan *amqp.Error) chan *amqp.Error {
	return ca.conn.NotifyClose(receiver)
}

// Abort closes the connection without waiting more than a short time for the server
func (ca *connectionAdapter) Abort() error {
	return ca.conn.CloseDeadline(time.Now().Add(abortTimeout))
}

type channelAdapter struct {
	ch *amqp.Channel
}

// Confirm puts the channel in the confirm mode: the server confirms every published message
func (cha *channelAdapter) Confirm(noWait bool) error {
	return cha.ch.Confirm(noWait)
}

// ExchangeDeclare declares a durable exchange
func (cha *channelAdapter) ExchangeDeclare(name, kind string) error {
	durable, autoDelete, internal, noWait := true, false, false, false

	return cha.ch.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, nil)
}

// Publish publishes the message and returns the confirmation that has to be waited
func (cha *channelAdapter) Publish(ctx context.Context, exchange string, msg amqp.Publishing) (amqpConfirmation, error) {
	routingKey, mandatory, immediate := "", false, false

	confirmation, err := cha.ch.PublishWithDeferredConfirmWithContext(ctx, exchange, routingKey, mandatory, immediate, msg)
	if err != nil {
		return nil, err
	}
	if confirmation == nil {
		// the channel is not in the confirm mode
		return nil, ErrMessageNotConfirmed
	}

	return confirmation, nil
}
