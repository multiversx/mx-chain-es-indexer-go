package utils

import (
	"math"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

type multiShardCoordinator struct {
	maskHigh       uint32
	maskLow        uint32
	selfId         uint32
	numberOfShards uint32
}

// NewMultiShardCoordinator returns a new multiShardCoordinator and initializes the masks
func NewMultiShardCoordinator(numberOfShards, selfId uint32) (*multiShardCoordinator, error) {
	sr := &multiShardCoordinator{}
	sr.selfId = selfId
	sr.numberOfShards = numberOfShards
	sr.maskHigh, sr.maskLow = sr.calculateMasks()

	return sr, nil
}

func (msc *multiShardCoordinator) calculateMasks() (uint32, uint32) {
	n := math.Ceil(math.Log2(float64(msc.numberOfShards)))
	return (1 << uint(n)) - 1, (1 << uint(n-1)) - 1
}

// ComputeId calculates the shard for a given address container
func (msc *multiShardCoordinator) ComputeId(address []byte) uint32 {
	return msc.ComputeIdFromBytes(address)
}

// ComputeIdFromBytes calculates the shard for a given address
func (msc *multiShardCoordinator) ComputeIdFromBytes(address []byte) uint32 {

	var bytesNeed int
	if msc.numberOfShards <= 256 {
		bytesNeed = 1
	} else if msc.numberOfShards <= 65536 {
		bytesNeed = 2
	} else if msc.numberOfShards <= 16777216 {
		bytesNeed = 3
	} else {
		bytesNeed = 4
	}

	startingIndex := 0
	if len(address) > bytesNeed {
		startingIndex = len(address) - bytesNeed
	}

	buffNeeded := address[startingIndex:]
	if core.IsSmartContractOnMetachain(buffNeeded, address) {
		return core.MetachainShardId
	}

	addr := uint32(0)
	for i := 0; i < len(buffNeeded); i++ {
		addr = addr<<8 + uint32(buffNeeded[i])
	}

	shard := addr & msc.maskHigh
	if shard > msc.numberOfShards-1 {
		shard = addr & msc.maskLow
	}

	return shard
}

// SelfId gets the shard id of the current node
func (msc *multiShardCoordinator) SelfId() uint32 {
	return msc.selfId
}

// IsInterfaceNil returns true if there is no value under the interface
func (msc *multiShardCoordinator) IsInterfaceNil() bool {
	return msc == nil
}
