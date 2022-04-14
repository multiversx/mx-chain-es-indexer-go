package checkers

import (
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var (
	log = logger.GetOrCreate("pkg/checkers")
)

type clusterChecker struct {
	clientSource      ESClient
	clientDestination ESClient
	indices           []string
}

func (cc *clusterChecker) CompareIndices() error {
	for _, index := range cc.indices {
		err := cc.compareIndex(index)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cc *clusterChecker) compareIndex(index string) error {
	return nil
}
