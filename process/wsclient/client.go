package wsclient

import (
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters/uint64ByteSlice"
	"github.com/multiversx/mx-chain-core-go/websocketOutportDriver"
	"github.com/multiversx/mx-chain-core-go/websocketOutportDriver/data"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const closedConnection = "use of closed network connection"

type operationsHandler interface {
	GetOperationsMap() map[data.OperationType]func(marshalledData []byte) error
	Close() error
}

type wsConn interface {
	io.Closer
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
}

var (
	log           = logger.GetOrCreate("process/wsclient")
	retryDuration = time.Second * 5
)

type client struct {
	urlReceive               string
	closeActions             func() error
	actions                  map[data.OperationType]func(marshalledData []byte) error
	uint64ByteSliceConverter websocketOutportDriver.Uint64ByteSliceConverter
	wsConnection             wsConn
}

// New will create a new instance of websocket client
func New(
	urlReceive string,
	operationsHandler operationsHandler,
) (*client, error) {
	urlReceiveData := url.URL{Scheme: "ws", Host: urlReceive, Path: data.WSRoute}

	return &client{
		actions:                  operationsHandler.GetOperationsMap(),
		closeActions:             operationsHandler.Close,
		urlReceive:               urlReceiveData.String(),
		uint64ByteSliceConverter: uint64ByteSlice.NewBigEndianConverter(),
	}, nil
}

// Start will initialize the connection to the server and start to listen for messages
func (c *client) Start() {
	log.Info("connecting to", "url", c.urlReceive)

	for {
		err := c.openConnection()
		if err != nil {
			log.Warn(fmt.Sprintf("c.openConnection(), retrying in %v...", retryDuration), "error", err.Error())
			time.Sleep(retryDuration)
			continue
		}

		closed := c.listeningOnWebSocket()
		if closed {
			return
		}
	}
}

func (c *client) openConnection() error {
	var err error
	c.wsConnection, _, err = websocket.DefaultDialer.Dial(c.urlReceive, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) listeningOnWebSocket() (closed bool) {
	for {
		_, message, err := c.wsConnection.ReadMessage()
		if err == nil {
			c.verifyPayloadAndSendAckIfNeeded(message)
			continue
		}

		_, isConnectionClosed := err.(*websocket.CloseError)
		if !isConnectionClosed {
			if strings.Contains(err.Error(), closedConnection) {
				return true
			}
			log.Warn("c.listeningOnWebSocket()-> connection problem, retrying", "error", err.Error())
		} else {
			log.Warn(fmt.Sprintf("websocket terminated by the server side, retrying in %v...", retryDuration), "error", err.Error())
		}
		return
	}

}

func (c *client) verifyPayloadAndSendAckIfNeeded(payload []byte) {
	if len(payload) == 0 {
		log.Error("empty payload")
		return
	}

	payloadParser, _ := websocketOutportDriver.NewWebSocketPayloadParser(c.uint64ByteSliceConverter)
	payloadData, err := payloadParser.ExtractPayloadData(payload)
	if err != nil {
		log.Error("error while extracting payload data: " + err.Error())
		return
	}

	log.Info("processing payload",
		"counter", payloadData.Counter,
		"operation type", payloadData.OperationType,
		"message length", len(payloadData.Payload),
	)

	function, ok := c.actions[payloadData.OperationType]
	if !ok {
		log.Warn("invalid operation", "operation type", payloadData.OperationType.String())
	}

	err = function(payloadData.Payload)
	if err != nil {
		log.Error("something went wrong", "error", err.Error())
	}

	if payloadData.WithAcknowledge {
		counterBytes := c.uint64ByteSliceConverter.ToByteSlice(payloadData.Counter)
		err = c.wsConnection.WriteMessage(websocket.BinaryMessage, counterBytes)
		if err != nil {
			log.Error("write acknowledge message", "error", err.Error())
		}
	}
}

func (c *client) closeWsConnection() {
	log.Debug("closing ws connection...")
	if check.IfNilReflect(c.wsConnection) {
		return
	}

	//Cleanly close the connection by sending a close message and then
	//waiting (with timeout) for the server to close the connection.
	err := c.wsConnection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Error("cannot send close message", "error", err)
	}
	err = c.wsConnection.Close()
	if err != nil {
		log.Error("cannot close ws connection", "error", err)
	}
}

func (c *client) Close() {
	log.Info("closing all components...")
	c.closeWsConnection()

	err := c.closeActions()
	if err != nil {
		log.Error("cannot close the operations handler", "error", err)
	}
}
