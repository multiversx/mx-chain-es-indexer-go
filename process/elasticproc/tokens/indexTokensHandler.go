package tokens

import (
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc"
)

type indexTokensHandler struct{}

// NewIndexTokensHandler creates a new index tokens handler
func NewIndexTokensHandler() *indexTokensHandler {
	return &indexTokensHandler{}
}

// IndexCrossChainTokens returns no error
func (it *indexTokensHandler) IndexCrossChainTokens(_ elasticproc.DatabaseClientHandler, _ []*data.ScResult, _ *data.BufferSlice) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (it *indexTokensHandler) IsInterfaceNil() bool {
	return it == nil
}
