package rabbitmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Client defines what a RabbitMQ client should be able to do
type Client interface {
	// Publish sends the message to the exchange and returns after the server confirmed it
	Publish(ctx context.Context, exchange string, messageID string, body []byte) error
	Close() error
	IsInterfaceNil() bool
}

// the next interfaces are implemented by the amqp library, they are used in order to test the client without a server

type amqpConnection interface {
	Channel() (amqpChannel, error)
	NotifyClose(receiver chan *amqp.Error) chan *amqp.Error
	// Abort closes the connection without waiting more than a short time for the server
	Abort() error
}

type amqpChannel interface {
	Confirm(noWait bool) error
	ExchangeDeclare(name, kind string) error
	Publish(ctx context.Context, exchange string, msg amqp.Publishing) (amqpConfirmation, error)
}

type amqpConfirmation interface {
	WaitContext(ctx context.Context) (bool, error)
}
