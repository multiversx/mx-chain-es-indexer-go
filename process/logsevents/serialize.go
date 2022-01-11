package logsevents

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/converters"
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

// SerializeSCDeploys will serialize the provided smart contract deploys in a way that Elastic Search expects a bulk request
func (logsAndEventsProcessor) SerializeSCDeploys(deploys map[string]*data.ScDeployInfo) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()
	for scAddr, deployInfo := range deploys {
		meta := []byte(fmt.Sprintf(`{ "update" : { "_id" : "%s", "_type" : "_doc" } }%s`, scAddr, "\n"))

		serializedData, err := serializeDeploy(deployInfo)
		if err != nil {
			return nil, err
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}

func serializeDeploy(deployInfo *data.ScDeployInfo) ([]byte, error) {
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

	serializedDataStr := fmt.Sprintf(`{"script": {`+
		`"source": "if (!ctx._source.containsKey('upgrades')) { ctx._source.upgrades = [ params.elem ]; } else {  ctx._source.upgrades.add(params.elem); }",`+
		`"lang": "painless",`+
		`"params": {"elem": %s}},`+
		`"upsert": %s}`,
		string(upgradeSerialized), string(serializedData))

	return []byte(serializedDataStr), nil
}

// SerializeTokens will serialize the provided tokens data in a way that Elastic Search expects a bulk request
func (logsAndEventsProcessor) SerializeTokens(tokens []*data.TokenInfo, updateNFTData []*data.NFTDataUpdate) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()
	for _, tokenData := range tokens {
		meta, serializedData, err := serializeToken(tokenData)
		if err != nil {
			return nil, err
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			return nil, err
		}
	}

	err := converters.PrepareNFTUpdateData(buffSlice, updateNFTData, false)
	if err != nil {
		return nil, err
	}

	return buffSlice.Buffers(), nil
}

func serializeToken(tokenData *data.TokenInfo) ([]byte, []byte, error) {
	if tokenData.TransferOwnership {
		return serializeTokenTransferOwnership(tokenData)
	}

	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, tokenData.Token, "\n"))
	serializedData, err := json.Marshal(tokenData)
	if err != nil {
		return nil, nil, err
	}

	return meta, serializedData, nil
}

func serializeTokenTransferOwnership(tokenData *data.TokenInfo) ([]byte, []byte, error) {
	meta := []byte(fmt.Sprintf(`{ "update" : { "_id" : "%s", "_type" : "_doc" } }%s`, tokenData.Token, "\n"))
	tokenDataSerialized, err := json.Marshal(tokenData)
	if err != nil {
		return nil, nil, err
	}

	currentOwnerData := &data.OwnerData{}
	if len(tokenData.OwnersHistory) > 0 {
		currentOwnerData = tokenData.OwnersHistory[0]
	}

	ownerDataSerialized, err := json.Marshal(currentOwnerData)
	if err != nil {
		return nil, nil, err
	}

	serializedDataStr := fmt.Sprintf(`{"script": {`+
		`"source": "if (!ctx._source.containsKey('ownersHistory')) { ctx._source.ownersHistory = [ params.elem ] } else { ctx._source.ownersHistory.add(params.elem) } ctx._source.currentOwner = params.owner ",`+
		`"lang": "painless",`+
		`"params": {"elem": %s, "owner": "%s"}},`+
		`"upsert": %s}`,
		string(ownerDataSerialized), tokenData.CurrentOwner, string(tokenDataSerialized))

	return meta, []byte(serializedDataStr), nil
}

// SerializeDelegators will serialize the provided delegators in a way that Elastic Search expects a bulk request
func (lep *logsAndEventsProcessor) SerializeDelegators(delegators map[string]*data.Delegator) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()
	for _, delegator := range delegators {
		meta, serializedData, err := lep.prepareSerializedDelegator(delegator)
		if err != nil {
			return nil, err
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}

func (lep *logsAndEventsProcessor) prepareSerializedDelegator(delegator *data.Delegator) ([]byte, []byte, error) {
	id := lep.computeDelegatorID(delegator)
	if delegator.ShouldDelete {
		meta := []byte(fmt.Sprintf(`{ "delete" : { "_id" : "%s" } }%s`, id, "\n"))
		return meta, nil, nil
	}

	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, id, "\n"))
	serializedData, errMarshal := json.Marshal(delegator)
	if errMarshal != nil {
		return nil, nil, errMarshal
	}

	return meta, serializedData, nil
}

func (lep *logsAndEventsProcessor) computeDelegatorID(delegator *data.Delegator) string {
	delegatorContract := delegator.Address + delegator.Contract

	hashBytes := lep.hasher.Compute(delegatorContract)

	return base64.StdEncoding.EncodeToString(hashBytes)
}
