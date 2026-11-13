package rabbitmq

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	logger "github.com/multiversx/mx-chain-logger-go"
	amqp "github.com/rabbitmq/amqp091-go"
)

var log = logger.GetOrCreate("indexer/rabbitmq")

const (
	defaultDialTimeout       = 5 * time.Second
	defaultReconnectInterval = time.Second
	contentTypeJSON          = "application/json"
)

// ExchangeConfig holds the name and the type of an exchange
type ExchangeConfig struct {
	Name string
	Type string
}

// ArgsClient holds the arguments needed to create a RabbitMQ client
type ArgsClient struct {
	URL string
	// Exchanges are declared every time a connection is established
	Exchanges []ExchangeConfig
	// DialTimeout is the maximum time to wait for a connection to be established
	DialTimeout time.Duration
	// ReconnectInterval is the time to wait between two attempts to connect
	ReconnectInterval time.Duration
	// ConnectionName is the name of the connection, as it is seen on the server
	ConnectionName string
	// FailFast should be true if a publish should return an error immediately when there is no connection
	// or when the publish fails. If it is false, the publish is retried until the provided context is done
	FailFast bool
	// RequireInitialConnection should be true if the client cannot be created without a connection to the server
	RequireInitialConnection bool
}

type connectionState struct {
	conn amqpConnection
	ch   amqpChannel
}

type client struct {
	url               string
	loggableURL       string
	password          string
	exchanges         []ExchangeConfig
	dialTimeout       time.Duration
	reconnectInterval time.Duration
	connectionName    string
	failFast          bool
	dialFunc          func(url string, dialTimeout time.Duration, connectionName string) (amqpConnection, error)

	mutState sync.Mutex
	// state is nil when there is no connection
	state *connectionState
	// connected is closed when a connection is available and it is replaced when the connection is lost
	connected    chan struct{}
	reconnecting bool

	// only one message is published at a time
	mutPublish sync.Mutex

	closeCh   chan struct{}
	closeOnce sync.Once
}

// NewClient creates a RabbitMQ client that publishes persistent messages and waits for the confirmation of the server.
// If the connection is lost, the client connects again in the background
func NewClient(args ArgsClient) (*client, error) {
	return newClient(args, dialAMQP)
}

func newClient(
	args ArgsClient,
	dialFunc func(url string, dialTimeout time.Duration, connectionName string) (amqpConnection, error),
) (*client, error) {
	loggableURL, password, err := redactURL(args.URL)
	if err != nil {
		return nil, err
	}
	for _, exchange := range args.Exchanges {
		if exchange.Name == "" {
			return nil, ErrInvalidExchangeName
		}
		if exchange.Type == "" {
			return nil, fmt.Errorf("%w for exchange %s", ErrInvalidExchangeType, exchange.Name)
		}
	}

	dialTimeout := args.DialTimeout
	if dialTimeout <= 0 {
		dialTimeout = defaultDialTimeout
	}
	reconnectInterval := args.ReconnectInterval
	if reconnectInterval <= 0 {
		reconnectInterval = defaultReconnectInterval
	}

	c := &client{
		url:               args.URL,
		loggableURL:       loggableURL,
		password:          password,
		exchanges:         args.Exchanges,
		dialTimeout:       dialTimeout,
		reconnectInterval: reconnectInterval,
		connectionName:    args.ConnectionName,
		failFast:          args.FailFast,
		dialFunc:          dialFunc,
		connected:         make(chan struct{}),
		closeCh:           make(chan struct{}),
	}

	err = c.connect()
	if err == nil {
		return c, nil
	}
	if args.RequireInitialConnection {
		return nil, fmt.Errorf("%w, cannot connect to the RabbitMQ server %s", err, loggableURL)
	}

	log.Warn("cannot connect to the RabbitMQ server, will retry in the background", "url", loggableURL, "error", err)
	c.startReconnecting()

	return c, nil
}

// Publish sends the message to the exchange and returns after the server confirmed it.
// If the client was created with FailFast, an error is returned at the first failure,
// otherwise the publish is retried (the client connects again if needed) until the context is done
func (c *client) Publish(ctx context.Context, exchange string, messageID string, body []byte) error {
	c.mutPublish.Lock()
	defer c.mutPublish.Unlock()

	var lastErr error = ErrNotConnected
	for {
		if c.isClosed() {
			return ErrClientClosed
		}
		if ctx.Err() != nil {
			return fmt.Errorf("%w, last error: %s", ctx.Err(), lastErr.Error())
		}

		state, connected := c.getState()
		if state == nil {
			if c.failFast {
				return lastErr
			}

			select {
			case <-connected:
			case <-ctx.Done():
			case <-c.closeCh:
			}
			continue
		}

		err := c.publishAndWaitConfirmation(ctx, state, exchange, messageID, body)
		if err == nil {
			return nil
		}

		lastErr = err
		log.Warn("cannot publish the message to the RabbitMQ server", "exchange", exchange, "message id", messageID, "error", err)

		// a new connection is used for the next attempt
		c.handleConnectionLost(state)
		if c.failFast {
			return err
		}
	}
}

func (c *client) publishAndWaitConfirmation(ctx context.Context, state *connectionState, exchange string, messageID string, body []byte) error {
	// the library does not interrupt a publish that is in progress when the context is done,
	// closing the connection is the only way to interrupt it
	stopAbort := context.AfterFunc(ctx, func() {
		_ = state.conn.Abort()
	})
	defer stopAbort()

	confirmation, err := state.ch.Publish(ctx, exchange, amqp.Publishing{
		ContentType:  contentTypeJSON,
		DeliveryMode: amqp.Persistent,
		MessageId:    messageID,
		Body:         body,
	})
	if err != nil {
		return err
	}

	confirmed, err := confirmation.WaitContext(ctx)
	if err != nil {
		return err
	}
	if !confirmed {
		return ErrMessageNotConfirmed
	}

	return nil
}

// Close closes the connection, a publish that is in progress is interrupted
func (c *client) Close() error {
	var err error
	c.closeOnce.Do(func() {
		close(c.closeCh)

		c.mutState.Lock()
		state := c.state
		c.state = nil
		c.mutState.Unlock()

		if state != nil {
			err = state.conn.Abort()
		}
	})

	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *client) IsInterfaceNil() bool {
	return c == nil
}

func (c *client) isClosed() bool {
	select {
	case <-c.closeCh:
		return true
	default:
		return false
	}
}

func (c *client) getState() (*connectionState, chan struct{}) {
	c.mutState.Lock()
	defer c.mutState.Unlock()

	return c.state, c.connected
}

func (c *client) connect() error {
	conn, err := c.dialFunc(c.url, c.dialTimeout, c.connectionName)
	if err != nil {
		return c.sanitizeError(err)
	}

	state, err := c.prepareConnection(conn)
	if err != nil {
		_ = conn.Abort()
		return err
	}

	closeNotifier := conn.NotifyClose(make(chan *amqp.Error, 1))

	c.mutState.Lock()
	if c.isClosed() {
		c.mutState.Unlock()
		_ = conn.Abort()
		return ErrClientClosed
	}
	c.state = state
	close(c.connected)
	c.mutState.Unlock()

	log.Info("connected to the RabbitMQ server", "url", c.loggableURL)

	go c.waitConnectionClose(state, closeNotifier)

	return nil
}

func (c *client) prepareConnection(conn amqpConnection) (*connectionState, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// every published message has to be confirmed by the server
	err = ch.Confirm(false)
	if err != nil {
		return nil, err
	}

	// the server can be a different one after a reconnect, so the exchanges are declared for every connection
	for _, exchange := range c.exchanges {
		err = ch.ExchangeDeclare(exchange.Name, exchange.Type)
		if err != nil {
			return nil, fmt.Errorf("%w when declaring the exchange %s", err, exchange.Name)
		}
	}

	return &connectionState{conn: conn, ch: ch}, nil
}

func (c *client) waitConnectionClose(state *connectionState, closeNotifier chan *amqp.Error) {
	select {
	case amqpErr := <-closeNotifier:
		if !c.isClosed() {
			log.Warn("the connection to the RabbitMQ server was closed", "url", c.loggableURL, "reason", amqpErr)
		}
		c.handleConnectionLost(state)
	case <-c.closeCh:
	}
}

func (c *client) handleConnectionLost(state *connectionState) {
	c.mutState.Lock()
	if c.state != state {
		// the connection was already replaced
		c.mutState.Unlock()
		return
	}
	c.state = nil
	c.connected = make(chan struct{})
	c.mutState.Unlock()

	_ = state.conn.Abort()

	c.startReconnecting()
}

func (c *client) startReconnecting() {
	c.mutState.Lock()
	shouldStart := !c.reconnecting && !c.isClosed()
	if shouldStart {
		c.reconnecting = true
	}
	c.mutState.Unlock()

	if shouldStart {
		go c.reconnect()
	}
}

func (c *client) reconnect() {
	defer func() {
		c.mutState.Lock()
		c.reconnecting = false
		c.mutState.Unlock()
	}()

	for {
		select {
		case <-time.After(c.reconnectInterval):
		case <-c.closeCh:
			return
		}

		err := c.connect()
		if err == nil || c.isClosed() {
			return
		}

		log.Warn("cannot connect to the RabbitMQ server, will retry", "url", c.loggableURL, "error", err)
	}
}

// the password is removed from the error message because the error will be logged
func (c *client) sanitizeError(err error) error {
	message := err.Error()
	if c.password != "" {
		message = strings.ReplaceAll(message, c.password, "***")
		message = strings.ReplaceAll(message, url.QueryEscape(c.password), "***")
		message = strings.ReplaceAll(message, url.PathEscape(c.password), "***")
	}

	return fmt.Errorf("%w: %s", ErrNotConnected, message)
}
