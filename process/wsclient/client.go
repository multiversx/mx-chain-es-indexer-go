package wsclient

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go-core/websocketOutportDriver"
	"github.com/ElrondNetwork/elrond-go-core/websocketOutportDriver/data"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/gorilla/websocket"
)

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
	urlReceive string
	actions    map[data.OperationType]func(marshalledData []byte) error
}

func NewWebSocketClient(urlReceive string, actions map[data.OperationType]func(marshalledData []byte) error) (*client, error) {
	urlReceiveData := url.URL{Scheme: "ws", Host: fmt.Sprintf(urlReceive), Path: "/operations"}

	return &client{
		actions:    actions,
		urlReceive: urlReceiveData.String(),
	}, nil
}

func (c *client) Start() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	log.Info("connecting to", "url", c.urlReceive)

	var wsConnection *websocket.Conn
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			var err error
			wsConnection, err = c.openConnection()
			if err != nil {
				log.Error(fmt.Sprintf("websocket error, retrying in %v...", retryDuration), "error", err.Error())
				time.Sleep(retryDuration)
				continue
			}

			c.listeningOnWebSocket(wsConnection)
			time.Sleep(retryDuration)
		}
	}()

	timer := time.NewTimer(time.Second)
	defer timer.Stop()

	for {
		select {
		case <-done:
			return
		case <-timer.C:
		case <-interrupt:
			log.Info("interrupt")
			if wsConnection == nil {
				return
			}

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := wsConnection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Error("write close", "error", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func (c *client) openConnection() (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(c.urlReceive, nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (c *client) listeningOnWebSocket(wsConnection *websocket.Conn) {
	for {
		_, message, err := wsConnection.ReadMessage()
		if err == nil {
			c.verifyPayloadAndSendAckIfNeeded(message, wsConnection)
			continue
		}

		_, isConnectionClosed := err.(*websocket.CloseError)
		if !isConnectionClosed {
			log.Warn("websocket error, retrying in %v...", "error", err.Error())
		} else {
			log.Warn(fmt.Sprintf("websocket terminated by the server side, retrying in %v...", retryDuration), "error", err.Error())
		}
		return
	}
}

func (c *client) verifyPayloadAndSendAckIfNeeded(payload []byte, ackHandler wsConn) {
	uint64ByteSliceConverter := uint64ByteSlice.NewBigEndianConverter()
	if len(payload) == 0 {
		log.Error("empty payload")
		return
	}

	payloadParser, _ := websocketOutportDriver.NewWebSocketPayloadParser(uint64ByteSliceConverter)
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
		counterBytes := uint64ByteSliceConverter.ToByteSlice(payloadData.Counter)
		err = ackHandler.WriteMessage(websocket.BinaryMessage, counterBytes)
		if err != nil {
			log.Error("write acknowledge message", "error", err.Error())
		}
	}
}

func (c *client) Close() {
	//TODO implement me
	panic("implement me")
}
