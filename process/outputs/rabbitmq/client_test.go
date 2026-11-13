package rabbitmq

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
)

const testURL = "amqp://user:secretPass@localhost:5672/"

type fakeConfirmation struct {
	confirmed bool
	// if wait is set, the confirmation arrives only when the channel is closed
	wait chan struct{}
}

func (fc *fakeConfirmation) WaitContext(ctx context.Context) (bool, error) {
	if fc.wait == nil {
		return fc.confirmed, nil
	}

	select {
	case <-fc.wait:
		return fc.confirmed, nil
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

type fakeChannel struct {
	conn           *fakeConnection
	mut            sync.Mutex
	confirmMode    bool
	declared       []ExchangeConfig
	published      []amqp.Publishing
	exchanges      []string
	confirmErr     error
	declareErr     error
	publishHandler func(ctx context.Context, exchange string, msg amqp.Publishing) (amqpConfirmation, error)
}

func (fch *fakeChannel) Confirm(_ bool) error {
	fch.mut.Lock()
	defer fch.mut.Unlock()

	fch.confirmMode = true

	return fch.confirmErr
}

func (fch *fakeChannel) ExchangeDeclare(name, kind string) error {
	fch.mut.Lock()
	defer fch.mut.Unlock()

	fch.declared = append(fch.declared, ExchangeConfig{Name: name, Type: kind})

	return fch.declareErr
}

func (fch *fakeChannel) Publish(ctx context.Context, exchange string, msg amqp.Publishing) (amqpConfirmation, error) {
	// nothing can be published on a closed connection
	if fch.conn != nil && fch.conn.isClosed() {
		return nil, amqp.ErrClosed
	}

	fch.mut.Lock()
	fch.published = append(fch.published, msg)
	fch.exchanges = append(fch.exchanges, exchange)
	handler := fch.publishHandler
	fch.mut.Unlock()

	if handler != nil {
		return handler(ctx, exchange, msg)
	}

	return &fakeConfirmation{confirmed: true}, nil
}

func (fch *fakeChannel) numPublished() int {
	fch.mut.Lock()
	defer fch.mut.Unlock()

	return len(fch.published)
}

type fakeConnection struct {
	channel    *fakeChannel
	channelErr error
	abortCalls int32

	mut       sync.Mutex
	notifiers []chan *amqp.Error
	closed    bool
}

func newFakeConnection() *fakeConnection {
	conn := &fakeConnection{channel: &fakeChannel{}}
	conn.channel.conn = conn

	return conn
}

func (fc *fakeConnection) isClosed() bool {
	fc.mut.Lock()
	defer fc.mut.Unlock()

	return fc.closed
}

func (fc *fakeConnection) Channel() (amqpChannel, error) {
	if fc.channelErr != nil {
		return nil, fc.channelErr
	}

	return fc.channel, nil
}

func (fc *fakeConnection) NotifyClose(receiver chan *amqp.Error) chan *amqp.Error {
	fc.mut.Lock()
	defer fc.mut.Unlock()

	if fc.closed {
		close(receiver)
		return receiver
	}
	fc.notifiers = append(fc.notifiers, receiver)

	return receiver
}

func (fc *fakeConnection) Abort() error {
	atomic.AddInt32(&fc.abortCalls, 1)
	fc.closeWithError(nil)

	return nil
}

// closeWithError simulates a connection closed by the server or by the network
func (fc *fakeConnection) closeWithError(err *amqp.Error) {
	fc.mut.Lock()
	defer fc.mut.Unlock()

	if fc.closed {
		return
	}
	fc.closed = true
	for _, notifier := range fc.notifiers {
		if err != nil {
			notifier <- err
		}
		close(notifier)
	}
}

func (fc *fakeConnection) numAborts() int {
	return int(atomic.LoadInt32(&fc.abortCalls))
}

// fakeDialer returns the provided connections one by one, after that it returns an error
type fakeDialer struct {
	mut         sync.Mutex
	connections []*fakeConnection
	errors      []error
	numDials    int
}

func (fd *fakeDialer) dial(_ string, _ time.Duration, _ string) (amqpConnection, error) {
	fd.mut.Lock()
	defer fd.mut.Unlock()

	idx := fd.numDials
	fd.numDials++

	if idx < len(fd.errors) && fd.errors[idx] != nil {
		return nil, fd.errors[idx]
	}
	if idx >= len(fd.connections) || fd.connections[idx] == nil {
		return nil, errors.New("no more connections")
	}

	return fd.connections[idx], nil
}

func (fd *fakeDialer) dials() int {
	fd.mut.Lock()
	defer fd.mut.Unlock()

	return fd.numDials
}

func createMockArgsClient() ArgsClient {
	return ArgsClient{
		URL: testURL,
		Exchanges: []ExchangeConfig{
			{Name: "blocks", Type: "fanout"},
			{Name: "reverts", Type: "fanout"},
		},
		ReconnectInterval:        10 * time.Millisecond,
		ConnectionName:           "test",
		RequireInitialConnection: true,
	}
}

func contextWithTimeout(t *testing.T, timeout time.Duration) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)

	return ctx
}

func TestNewClient_InvalidArguments(t *testing.T) {
	t.Parallel()

	dialer := &fakeDialer{connections: []*fakeConnection{newFakeConnection()}}

	t.Run("invalid url", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsClient()
		args.URL = "http://user:secretPass@localhost"
		c, err := newClient(args, dialer.dial)
		require.Nil(t, c)
		require.Equal(t, ErrInvalidURL, err)
	})
	t.Run("empty exchange name", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsClient()
		args.Exchanges = []ExchangeConfig{{Name: "", Type: "fanout"}}
		c, err := newClient(args, dialer.dial)
		require.Nil(t, c)
		require.Equal(t, ErrInvalidExchangeName, err)
	})
	t.Run("empty exchange type", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsClient()
		args.Exchanges = []ExchangeConfig{{Name: "blocks", Type: ""}}
		c, err := newClient(args, dialer.dial)
		require.Nil(t, c)
		require.True(t, errors.Is(err, ErrInvalidExchangeType))
	})
}

func TestNewClient_ShouldConnectAndDeclareTheExchanges(t *testing.T) {
	t.Parallel()

	conn := newFakeConnection()
	dialer := &fakeDialer{connections: []*fakeConnection{conn}}

	c, err := newClient(createMockArgsClient(), dialer.dial)
	require.Nil(t, err)
	require.False(t, c.IsInterfaceNil())
	defer func() {
		_ = c.Close()
	}()

	require.Equal(t, 1, dialer.dials())
	require.True(t, conn.channel.confirmMode)
	require.Equal(t, []ExchangeConfig{{Name: "blocks", Type: "fanout"}, {Name: "reverts", Type: "fanout"}}, conn.channel.declared)
}

func TestNewClient_CannotConnect(t *testing.T) {
	t.Parallel()

	dialErr := errors.New("dial amqp://user:secretPass@localhost:5672/: connection refused")

	t.Run("the initial connection is required", func(t *testing.T) {
		t.Parallel()

		dialer := &fakeDialer{errors: []error{dialErr}}
		c, err := newClient(createMockArgsClient(), dialer.dial)
		require.Nil(t, c)
		require.True(t, errors.Is(err, ErrNotConnected))
		// the error is logged, it should not contain the password
		require.NotContains(t, err.Error(), "secretPass")
		require.Contains(t, err.Error(), "connection refused")
	})
	t.Run("the initial connection is not required: the client connects in the background", func(t *testing.T) {
		t.Parallel()

		conn := newFakeConnection()
		dialer := &fakeDialer{errors: []error{dialErr, dialErr}, connections: []*fakeConnection{nil, nil, conn}}
		args := createMockArgsClient()
		args.RequireInitialConnection = false

		c, err := newClient(args, dialer.dial)
		require.Nil(t, err)
		defer func() {
			_ = c.Close()
		}()

		err = c.Publish(contextWithTimeout(t, 5*time.Second), "blocks", "id", []byte("body"))
		require.Nil(t, err)
		require.Equal(t, 3, dialer.dials())
		require.Equal(t, 1, conn.channel.numPublished())
	})
	t.Run("the exchange cannot be declared", func(t *testing.T) {
		t.Parallel()

		conn := newFakeConnection()
		conn.channel.declareErr = errors.New("declare error")
		dialer := &fakeDialer{connections: []*fakeConnection{conn}}

		c, err := newClient(createMockArgsClient(), dialer.dial)
		require.Nil(t, c)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "blocks")
		// the connection that cannot be used is closed
		require.Equal(t, 1, conn.numAborts())
	})
}

func TestClient_PublishShouldSendPersistentMessageAndWaitTheConfirmation(t *testing.T) {
	t.Parallel()

	conn := newFakeConnection()
	dialer := &fakeDialer{connections: []*fakeConnection{conn}}
	c, _ := newClient(createMockArgsClient(), dialer.dial)
	defer func() {
		_ = c.Close()
	}()

	err := c.Publish(contextWithTimeout(t, 5*time.Second), "blocks", "blockHash", []byte(`{"hash":"blockHash"}`))
	require.Nil(t, err)

	require.Equal(t, []string{"blocks"}, conn.channel.exchanges)
	require.Equal(t, []amqp.Publishing{
		{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    "blockHash",
			Body:         []byte(`{"hash":"blockHash"}`),
		},
	}, conn.channel.published)
	require.Equal(t, 0, conn.numAborts())
}

func TestClient_PublishErrorShouldReconnectAndRetry(t *testing.T) {
	t.Parallel()

	t.Run("the publish returns an error", func(t *testing.T) {
		t.Parallel()

		firstConn := newFakeConnection()
		firstConn.channel.publishHandler = func(_ context.Context, _ string, _ amqp.Publishing) (amqpConfirmation, error) {
			return nil, errors.New("channel closed")
		}
		secondConn := newFakeConnection()
		dialer := &fakeDialer{connections: []*fakeConnection{firstConn, secondConn}}
		c, _ := newClient(createMockArgsClient(), dialer.dial)
		defer func() {
			_ = c.Close()
		}()

		err := c.Publish(contextWithTimeout(t, 5*time.Second), "blocks", "id", []byte("body"))
		require.Nil(t, err)

		// the first connection was closed and the message was published on a new connection
		require.Equal(t, 1, firstConn.numAborts())
		require.Equal(t, 2, dialer.dials())
		require.Equal(t, 1, secondConn.channel.numPublished())
		// the exchanges are declared for every connection
		require.Len(t, secondConn.channel.declared, 2)
	})
	t.Run("the server does not confirm the message", func(t *testing.T) {
		t.Parallel()

		firstConn := newFakeConnection()
		firstConn.channel.publishHandler = func(_ context.Context, _ string, _ amqp.Publishing) (amqpConfirmation, error) {
			return &fakeConfirmation{confirmed: false}, nil
		}
		secondConn := newFakeConnection()
		dialer := &fakeDialer{connections: []*fakeConnection{firstConn, secondConn}}
		c, _ := newClient(createMockArgsClient(), dialer.dial)
		defer func() {
			_ = c.Close()
		}()

		err := c.Publish(contextWithTimeout(t, 5*time.Second), "blocks", "id", []byte("body"))
		require.Nil(t, err)
		require.Equal(t, 1, firstConn.channel.numPublished())
		require.Equal(t, 1, secondConn.channel.numPublished())
	})
}

func TestClient_ConnectionClosedByTheServerShouldReconnectInTheBackground(t *testing.T) {
	t.Parallel()

	firstConn := newFakeConnection()
	secondConn := newFakeConnection()
	dialer := &fakeDialer{connections: []*fakeConnection{firstConn, secondConn}}
	c, _ := newClient(createMockArgsClient(), dialer.dial)
	defer func() {
		_ = c.Close()
	}()

	firstConn.closeWithError(&amqp.Error{Code: 320, Reason: "CONNECTION_FORCED"})

	// no publish is needed in order to connect again
	require.Eventually(t, func() bool {
		return dialer.dials() == 2
	}, 5*time.Second, 5*time.Millisecond)

	err := c.Publish(contextWithTimeout(t, 5*time.Second), "blocks", "id", []byte("body"))
	require.Nil(t, err)
	require.Equal(t, 0, firstConn.channel.numPublished())
	require.Equal(t, 1, secondConn.channel.numPublished())
}

func TestClient_PublishWithoutConnection(t *testing.T) {
	t.Parallel()

	t.Run("the publish waits for the connection until the context is done", func(t *testing.T) {
		t.Parallel()

		conn := newFakeConnection()
		dialer := &fakeDialer{connections: []*fakeConnection{conn}}
		c, _ := newClient(createMockArgsClient(), dialer.dial)
		defer func() {
			_ = c.Close()
		}()

		// the connection is lost and the server cannot be reached anymore
		conn.closeWithError(&amqp.Error{Code: 320, Reason: "CONNECTION_FORCED"})

		startTime := time.Now()
		err := c.Publish(contextWithTimeout(t, 100*time.Millisecond), "blocks", "id", []byte("body"))
		require.True(t, errors.Is(err, context.DeadlineExceeded))
		require.GreaterOrEqual(t, time.Since(startTime), 100*time.Millisecond)
		require.Equal(t, 0, conn.channel.numPublished())
	})
	t.Run("fail fast: the publish returns an error immediately", func(t *testing.T) {
		t.Parallel()

		conn := newFakeConnection()
		dialer := &fakeDialer{connections: []*fakeConnection{conn}}
		args := createMockArgsClient()
		args.FailFast = true
		c, _ := newClient(args, dialer.dial)
		defer func() {
			_ = c.Close()
		}()

		conn.closeWithError(&amqp.Error{Code: 320, Reason: "CONNECTION_FORCED"})
		require.Eventually(t, func() bool {
			state, _ := c.getState()
			return state == nil
		}, 5*time.Second, time.Millisecond)

		startTime := time.Now()
		err := c.Publish(contextWithTimeout(t, time.Minute), "blocks", "id", []byte("body"))
		require.True(t, errors.Is(err, ErrNotConnected))
		require.Less(t, time.Since(startTime), 5*time.Second)
	})
	t.Run("fail fast: a publish error is returned without retry", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("channel closed")
		firstConn := newFakeConnection()
		firstConn.channel.publishHandler = func(_ context.Context, _ string, _ amqp.Publishing) (amqpConfirmation, error) {
			return nil, expectedErr
		}
		secondConn := newFakeConnection()
		dialer := &fakeDialer{connections: []*fakeConnection{firstConn, secondConn}}
		args := createMockArgsClient()
		args.FailFast = true
		c, _ := newClient(args, dialer.dial)
		defer func() {
			_ = c.Close()
		}()

		err := c.Publish(contextWithTimeout(t, time.Minute), "blocks", "id", []byte("body"))
		require.Equal(t, expectedErr, err)
		require.Equal(t, 0, secondConn.channel.numPublished())

		// the client connects again in the background and the next message is published
		require.Eventually(t, func() bool {
			return c.Publish(contextWithTimeout(t, time.Minute), "blocks", "id", []byte("body")) == nil
		}, 5*time.Second, 5*time.Millisecond)
		require.Equal(t, 1, secondConn.channel.numPublished())
	})
}

func TestClient_PublishTimeoutShouldCloseTheConnection(t *testing.T) {
	t.Parallel()

	// the server never confirms the message
	conn := newFakeConnection()
	conn.channel.publishHandler = func(_ context.Context, _ string, _ amqp.Publishing) (amqpConfirmation, error) {
		return &fakeConfirmation{wait: make(chan struct{})}, nil
	}
	dialer := &fakeDialer{connections: []*fakeConnection{conn}}
	c, _ := newClient(createMockArgsClient(), dialer.dial)
	defer func() {
		_ = c.Close()
	}()

	err := c.Publish(contextWithTimeout(t, 50*time.Millisecond), "blocks", "id", []byte("body"))
	require.True(t, errors.Is(err, context.DeadlineExceeded))

	// the library cannot interrupt a publish that is in progress, so the connection is closed when the time is up
	require.Eventually(t, func() bool {
		return conn.numAborts() > 0
	}, 5*time.Second, time.Millisecond)
}

func TestClient_Close(t *testing.T) {
	t.Parallel()

	t.Run("close interrupts the publish that waits for a connection", func(t *testing.T) {
		t.Parallel()

		conn := newFakeConnection()
		dialer := &fakeDialer{connections: []*fakeConnection{conn}}
		c, _ := newClient(createMockArgsClient(), dialer.dial)

		conn.closeWithError(&amqp.Error{Code: 320, Reason: "CONNECTION_FORCED"})
		require.Eventually(t, func() bool {
			state, _ := c.getState()
			return state == nil
		}, 5*time.Second, time.Millisecond)

		errCh := make(chan error, 1)
		go func() {
			errCh <- c.Publish(contextWithTimeout(t, time.Minute), "blocks", "id", []byte("body"))
		}()

		time.Sleep(20 * time.Millisecond)
		require.Nil(t, c.Close())

		select {
		case err := <-errCh:
			require.Equal(t, ErrClientClosed, err)
		case <-time.After(5 * time.Second):
			require.Fail(t, "the publish was not interrupted")
		}
	})
	t.Run("close can be called more than once and the connection is closed", func(t *testing.T) {
		t.Parallel()

		conn := newFakeConnection()
		dialer := &fakeDialer{connections: []*fakeConnection{conn}}
		c, _ := newClient(createMockArgsClient(), dialer.dial)

		require.Nil(t, c.Close())
		require.Nil(t, c.Close())
		require.Equal(t, 1, conn.numAborts())

		err := c.Publish(contextWithTimeout(t, time.Second), "blocks", "id", []byte("body"))
		require.Equal(t, ErrClientClosed, err)

		// no connection is created after close
		time.Sleep(50 * time.Millisecond)
		require.Equal(t, 1, dialer.dials())
	})
}

func TestClient_ConcurrentPublishAndClose(t *testing.T) {
	t.Parallel()

	conn := newFakeConnection()
	dialer := &fakeDialer{connections: []*fakeConnection{conn}}
	c, _ := newClient(createMockArgsClient(), dialer.dial)

	wg := sync.WaitGroup{}
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = c.Publish(contextWithTimeout(t, time.Second), "blocks", "id", []byte("body"))
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(time.Millisecond)
		_ = c.Close()
	}()

	wg.Wait()
}

func TestRedactURL(t *testing.T) {
	t.Parallel()

	loggableURL, password, err := redactURL("amqp://user:secretPass@rabbit.example.com:5673/vhost")
	require.Nil(t, err)
	require.Equal(t, "amqp://rabbit.example.com:5673/vhost", loggableURL)
	require.Equal(t, "secretPass", password)
	require.NotContains(t, loggableURL, "user")

	_, _, err = redactURL("not a url with secretPass")
	require.Equal(t, ErrInvalidURL, err)
}
