package converters

import (
	coreData "github.com/multiversx/mx-chain-core-go/data"
)

type executionResultHandler interface {
	GetMiniBlockHeadersHandlers() []coreData.MiniBlockHeaderHandler
}

// GetMiniBlocksHeaderHandlersFromExecResult returns miniblock handlers based on execution result
func GetMiniBlocksHeaderHandlersFromExecResult(baseExecResult coreData.BaseExecutionResultHandler) []coreData.MiniBlockHeaderHandler {
	execResult, ok := baseExecResult.(executionResultHandler)
	if !ok {
		return []coreData.MiniBlockHeaderHandler{}
	}

	return execResult.GetMiniBlockHeadersHandlers()
}
