package gin

import (
	"context"
	"errors"
	"net/http"
	"time"

	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("api/gin")

// ErrNilHttpServer signals that a nil http server has been provided
var ErrNilHttpServer = errors.New("nil http server")

type httpServer struct {
	server server
}

// NewHttpServer returns a new instance of httpServer
func NewHttpServer(server server) (*httpServer, error) {
	if server == nil {
		return nil, ErrNilHttpServer
	}

	return &httpServer{
		server: server,
	}, nil
}

// Start will handle the starting of the gin web server. This call is blocking, and it should be
// called on a go routine (different from the main one)
func (h *httpServer) Start() {
	err := h.server.ListenAndServe()
	if err == nil {
		return
	}

	if err == http.ErrServerClosed {
		log.Debug("ListenAndServe - webserver closed")
		return
	}

	log.Error("could not start webserver", "error", err.Error())
}

// Close will handle the stopping of the gin web server
func (h *httpServer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return h.server.Shutdown(ctx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (h *httpServer) IsInterfaceNil() bool {
	return h == nil
}
