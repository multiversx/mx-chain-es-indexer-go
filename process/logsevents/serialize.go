package logsevents

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

// SerializeLogs will serialize the provided logs in a way that Elastic Search expects a bulk request
func (logsAndEventsProcessor) SerializeLogs(logs []*data.Logs) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()
	for _, log := range logs {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, log.ID, "\n"))
		serializedData, errMarshal := json.Marshal(log)
		if errMarshal != nil {
			return nil, errMarshal
		}

		err := buffSlice.PutData(meta, serializedData)
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}

func (logsAndEventsProcessor) SerializeSCDeploys(deploys map[string]*data.ScDeployInfo) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()
	for scAddr, deployInfo := range deploys {
		deployInfo.Upgrades = make([]*data.Upgrade, 0)
		serializedData, errPrepareD := json.Marshal(deployInfo)
		if errPrepareD != nil {
			return nil, errPrepareD
		}

		upgradeData := &data.Upgrade{
			TxHash:    deployInfo.TxHash,
			Upgrader:  deployInfo.Creator,
			Timestamp: deployInfo.Timestamp,
		}
		upgradeSerialized, errPrepareU := json.Marshal(upgradeData)
		if errPrepareU != nil {
			return nil, errPrepareU
		}

		meta := []byte(fmt.Sprintf(`{ "update" : { "_id" : "%s", "_type" : "_doc" } }%s`, scAddr, "\n"))
		serializedDataStr := fmt.Sprintf(`{"script": {`+
			`"source": "ctx._source.upgrades.add(params.elem);",`+
			`"lang": "painless",`+
			`"params": {"elem": %s}},`+
			`"upsert": %s}`,
			string(upgradeSerialized), string(serializedData))

		err := buffSlice.PutData(meta, []byte(serializedDataStr))
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}
