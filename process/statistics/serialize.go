package statistics

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

// SerializeGeneralInfo will serialize statistics information
func (sp *statisticsProcessor) SerializeStatistics(genInfo *data.TPS, shardsInfo []*data.TPS, index string) (*bytes.Buffer, error) {
	buff, err := serializeStatisticInfo(genInfo, index)
	if err != nil {
		return nil, err
	}

	for _, shardInfo := range shardsInfo {
		errSerialize := serializeShardInfo(buff, shardInfo, index)
		if errSerialize != nil {
			log.Warn("serializeShardInfo", "shardID", shardInfo.ShardID, "error", errSerialize)

			continue
		}
	}

	return buff, nil
}

func serializeStatisticInfo(generalInfo *data.TPS, index string) (*bytes.Buffer, error) {
	buff := &bytes.Buffer{}
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s", "_type" : "%s" } }%s`, metachainTpsDocID, index, "\n"))

	serializedInfo, err := json.Marshal(generalInfo)
	if err != nil {
		return nil, err
	}

	serializedInfo = append(serializedInfo, "\n"...)

	buff.Grow(len(meta) + len(serializedInfo))
	_, err = buff.Write(meta)
	if err != nil {
		return nil, err
	}
	_, err = buff.Write(serializedInfo)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

func serializeShardInfo(buff *bytes.Buffer, shardTPS *data.TPS, index string) error {
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s%d", "_type" : "%s" } }%s`,
		shardTpsDocIDPrefix, shardTPS.ShardID, index, "\n"))

	serializedInfo, err := json.Marshal(shardTPS)
	if err != nil {
		return err
	}

	serializedInfo = append(serializedInfo, "\n"...)

	buff.Grow(len(meta) + len(serializedInfo))
	_, err = buff.Write(meta)
	if err != nil {
		return err
	}
	_, err = buff.Write(serializedInfo)
	if err != nil {
		return err
	}

	return nil
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
