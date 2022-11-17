package checkers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/tools/clusters-checker/pkg/client"
	"github.com/ElrondNetwork/elastic-indexer-go/tools/clusters-checker/pkg/config"
	"github.com/elastic/go-elasticsearch/v7"
)

// CreateClusterChecker will create a new instance of clusterChecker structure
func CreateClusterChecker(cfg *config.Config, interval *Interval, logPrefix string, onlyIDs bool) (*clusterChecker, error) {
	clientSource, err := client.NewElasticClient(elasticsearch.Config{
		Addresses: []string{cfg.SourceCluster.URL},
		Username:  cfg.SourceCluster.User,
		Password:  cfg.SourceCluster.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create source client %s", err.Error())
	}

	clientDestination, err := client.NewElasticClient(elasticsearch.Config{
		Addresses: []string{cfg.DestinationCluster.URL},
		Username:  cfg.DestinationCluster.User,
		Password:  cfg.DestinationCluster.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create destination client %s", err.Error())
	}

	return &clusterChecker{
		clientSource:         clientSource,
		clientDestination:    clientDestination,
		indicesWithTimestamp: cfg.Compare.IndicesWithTimestamp,
		indicesNoTimestamp:   cfg.Compare.IndicesNoTimestamp,

		missingFromSource:      map[string]json.RawMessage{},
		missingFromDestination: map[string]json.RawMessage{},

		startTimestamp: int(interval.start),
		stopTimestamp:  int(interval.stop),
		logPrefix:      logPrefix,
		onlyIDs:        onlyIDs,
	}, nil
}

func CreateMultipleCheckers(cfg *config.Config, onlyIDs bool) ([]*clusterChecker, error) {
	currentTimestampUnix := time.Now().Unix()
	intervals, err := computeIntervals(cfg.Compare.BlockchainStartTime, currentTimestampUnix, int64(cfg.Compare.NumParallelReads))
	if err != nil {
		return nil, err
	}

	checkers := make([]*clusterChecker, 0, cfg.Compare.NumParallelReads)

	for idx := 0; idx < cfg.Compare.NumParallelReads; idx++ {
		logPrefix := "instance_" + strconv.FormatUint(uint64(idx), 10)
		cc, errC := CreateClusterChecker(cfg, intervals[idx], logPrefix, onlyIDs)
		if errC != nil {
			return nil, errC
		}

		checkers = append(checkers, cc)
	}

	return checkers, nil
}

func computeIntervals(startTime, endTime int64, numIntervals int64) ([]*Interval, error) {
	if startTime > endTime {
		return nil, errors.New("blockchain start time is greater than current timestamp")
	}
	if numIntervals < 2 {
		return []*Interval{{
			start: startTime,
			stop:  endTime,
		}}, nil
	}

	difference := endTime - startTime

	step := difference / numIntervals

	intervals := make([]*Interval, 0)
	for idx := int64(0); idx < numIntervals; idx++ {
		start := startTime + idx*step
		stop := startTime + (idx+1)*step

		if idx == numIntervals-1 {
			stop = endTime
		}

		intervals = append(intervals, &Interval{
			start: start,
			stop:  stop,
		})
	}

	return intervals, nil
}
