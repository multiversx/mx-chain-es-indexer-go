package statistics

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("process/statistics")

type statisticsProcessor struct {
}

// NewStatisticsProcessor will create a new instance of a statisticsProcessor
func NewStatisticsProcessor() *statisticsProcessor {
	return &statisticsProcessor{}
}

// SerializeRoundsInfo will serialize information about rounds
func (sp *statisticsProcessor) SerializeRoundsInfo(roundsInfo []*data.RoundInfo) *bytes.Buffer {
	buff := &bytes.Buffer{}
	for _, info := range roundsInfo {
		serializedRoundInfo, meta := serializeRoundInfo(info)

		buff.Grow(len(meta) + len(serializedRoundInfo))
		_, err := buff.Write(meta)
		if err != nil {
			log.Warn("generalInfoProcessor.SaveRoundsInfo cannot write meta", "error", err)
		}

		_, err = buff.Write(serializedRoundInfo)
		if err != nil {
			log.Warn("generalInfoProcessor.SaveRoundsInfo cannot write serialized round info", "error", err)
		}
	}

	return buff
}

func serializeRoundInfo(info *data.RoundInfo) ([]byte, []byte) {
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%d_%d", "_type" : "%s" } }%s`,
		info.ShardId, info.Index, "_doc", "\n"))

	serializedInfo, err := json.Marshal(info)
	if err != nil {
		log.Warn("serializeRoundInfo could not serialize round info, will skip indexing this round info", "error", err)
		return nil, nil
	}

	serializedInfo = append(serializedInfo, "\n"...)

	return serializedInfo, meta
}
