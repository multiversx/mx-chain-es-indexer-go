package elasticproc

import (
	"github.com/multiversx/mx-chain-es-indexer-go/data"
)

// IndexTokenHandlerMock -
type IndexTokenHandlerMock struct {
	IndexCrossChainTokensCalled func(elasticClient DatabaseClientHandler, scrs []*data.ScResult, buffSlice *data.BufferSlice) error
}

// IndexCrossChainTokens -
func (ithh *IndexTokenHandlerMock) IndexCrossChainTokens(elasticClient DatabaseClientHandler, scrs []*data.ScResult, buffSlice *data.BufferSlice) error {
	if ithh.IndexCrossChainTokensCalled != nil {
		return ithh.IndexCrossChainTokensCalled(elasticClient, scrs, buffSlice)
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ithh *IndexTokenHandlerMock) IsInterfaceNil() bool {
	return ithh == nil
}
