package transactions

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
)

type rewardTxData struct{}

// NewRewardTxData creates a new reward tx data
func NewRewardTxData() *rewardTxData {
	return &rewardTxData{}
}

// GetSender return the metachain shard id as string
func (rtd *rewardTxData) GetSender() string {
	return fmt.Sprintf("%d", core.MetachainShardId)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rtd *rewardTxData) IsInterfaceNil() bool {
	return rtd == nil
}
