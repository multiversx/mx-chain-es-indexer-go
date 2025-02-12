package tokens

import (
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc"
)

type disabledTndexTokensHandler struct{}

// NewDisabledIndexTokensHandler creates a new disabled index tokens handler
func NewDisabledIndexTokensHandler() *disabledTndexTokensHandler {
	return &disabledTndexTokensHandler{}
}

// IndexCrossChainTokens should do nothing and return no error
func (dit *disabledTndexTokensHandler) IndexCrossChainTokens(_ elasticproc.DatabaseClientHandler, _ []*data.ScResult, _ *data.BufferSlice) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dit *disabledTndexTokensHandler) IsInterfaceNil() bool {
	return dit == nil
}
