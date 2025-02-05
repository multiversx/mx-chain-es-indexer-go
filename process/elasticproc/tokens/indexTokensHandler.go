package tokens

import (
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc"
)

type indexTokensHandler struct{}

// NewDisabledIndexTokensHandler creates a new disabled index tokens handler
func NewDisabledIndexTokensHandler() *indexTokensHandler {
	return &indexTokensHandler{}
}

// IndexCrossChainTokens should do nothing and return no error
func (it *indexTokensHandler) IndexCrossChainTokens(_ elasticproc.DatabaseClientHandler, _ []*data.ScResult, _ *data.BufferSlice) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (it *indexTokensHandler) IsInterfaceNil() bool {
	return it == nil
}
