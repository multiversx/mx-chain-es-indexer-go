package rabbitmq

import "errors"

// ErrNilClient signals that a nil RabbitMQ client has been provided
var ErrNilClient = errors.New("nil RabbitMQ client")

// ErrInvalidURL signals that the provided RabbitMQ url is not valid.
// The url is not part of the error because it can contain credentials
var ErrInvalidURL = errors.New("invalid RabbitMQ url")

// ErrInvalidExchangeName signals that an empty exchange name has been provided
var ErrInvalidExchangeName = errors.New("invalid RabbitMQ exchange name")

// ErrInvalidExchangeType signals that an empty exchange type has been provided
var ErrInvalidExchangeType = errors.New("invalid RabbitMQ exchange type")

// ErrInvalidPublishTimeout signals that an invalid publish timeout has been provided
var ErrInvalidPublishTimeout = errors.New("invalid RabbitMQ publish timeout")

// ErrClientClosed signals that the RabbitMQ client is closed
var ErrClientClosed = errors.New("RabbitMQ client is closed")

// ErrNotConnected signals that there is no connection to the RabbitMQ server
var ErrNotConnected = errors.New("not connected to the RabbitMQ server")

// ErrMessageNotConfirmed signals that the RabbitMQ server did not confirm the message
var ErrMessageNotConfirmed = errors.New("message was not confirmed by the RabbitMQ server")
