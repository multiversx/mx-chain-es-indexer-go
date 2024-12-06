package transactions

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
)

type sovereignRewardTxData struct{}

// NewSovereignRewardTxData creates a new sovereign reward tx data
func NewSovereignRewardTxData() *sovereignRewardTxData {
	return &sovereignRewardTxData{}
}

// GetSender return the sovereign shard id as string
func (srtd *sovereignRewardTxData) GetSender() string {
	return fmt.Sprintf("%d", core.SovereignChainShardId)
}

// IsInterfaceNil returns true if there is no value under the interface
func (srtd *sovereignRewardTxData) IsInterfaceNil() bool {
	return srtd == nil
}
