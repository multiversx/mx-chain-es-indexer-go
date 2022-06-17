package checkers

import "math"

func (cc *clusterChecker) CompareCounts() error {
	for _, index := range cc.indicesNoTimestamp {
		err := cc.compareCount(index)
		if err != nil {
			return err
		}
	}

	for _, index := range cc.indicesWithTimestamp {
		err := cc.compareCount(index)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cc *clusterChecker) compareCount(index string) error {
	countSourceCluster, err := cc.clientSource.DoCountRequest(index, nil)
	if err != nil {
		return err
	}

	countDestinationCluster, err := cc.clientDestination.DoCountRequest(index, nil)
	if err != nil {
		return err
	}

	difference := int64(countSourceCluster) - int64(countDestinationCluster)

	if difference == 0 {
		log.Info(cc.logPrefix+": number of documents is the same", "index", index,
			"source cluster", countSourceCluster,
			"destination cluster", countDestinationCluster,
		)
	} else if difference < 0 {
		log.Info(cc.logPrefix+": number of documents", "index", index,
			"source cluster", countSourceCluster,
			"destination cluster", countDestinationCluster,
			"in destination cluster are more elements, difference", uint64(math.Abs(float64(difference))),
		)

	} else {
		log.Info(cc.logPrefix+": number of documents", "index", index,
			"source cluster", countSourceCluster,
			"destination cluster", countDestinationCluster,
			"in source cluster are more elements, difference", uint64(math.Abs(float64(difference))),
		)
	}

	return nil
}
