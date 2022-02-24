package logsevents

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
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

	meta := []byte(fmt.Sprintf(`{ "update" : { "_id" : "%s", "_type" : "_doc" } }%s`, tokenData.Token, "\n"))
	serializedTokenData, err := json.Marshal(tokenData)
	if err != nil {
		return nil, nil, err
	}

	serializedDataStr := fmt.Sprintf(`{"script": {`+
		`"source": "if (ctx._source.containsKey('roles')) {HashMap roles = ctx._source.roles; ctx._source = params.token; ctx._source.roles = roles}",`+
		`"lang": "painless",`+
		`"params": {"token": %s}},`+
		`"upsert": %s}`,
		string(serializedTokenData), string(serializedTokenData))

	return meta, []byte(serializedDataStr), nil
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

// SerializeSupplyData will serialize the provided supply data
func (lep *logsAndEventsProcessor) SerializeSupplyData(tokensSupply data.TokensHandler) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()
	for _, supplyData := range tokensSupply.GetAll() {
		if supplyData.Type != core.NonFungibleESDT {
			continue
		}

		meta := []byte(fmt.Sprintf(`{ "delete" : { "_id" : "%s" } }%s`, supplyData.Identifier, "\n"))
		err := buffSlice.PutData(meta, nil)
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}

// SerializeRolesData will serialize the provided roles data
func (lep *logsAndEventsProcessor) SerializeRolesData(timestamp uint64, rolesData data.RolesData) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()
	for role, roleData := range rolesData {
		for _, rd := range roleData {
			err := serializeRoleData(buffSlice, rd, timestamp, role)
			if err != nil {
				return nil, err
			}
		}
	}

	return buffSlice.Buffers(), nil
}

func serializeRoleData(buffSlice *data.BufferSlice, rd *data.RoleData, timestamp uint64, role string) error {
	meta := []byte(fmt.Sprintf(`{ "update" : { "_id" : "%s", "_type" : "_doc" } }%s`, rd.Token, "\n"))
	var serializedDataStr string
	if rd.Set {
		serializedDataStr = fmt.Sprintf(`{"script": {`+
			`"source": "if (!ctx._source.containsKey('roles')) { ctx._source.roles =  new HashMap();} if (!ctx._source.roles.containsKey(params.role)) { ctx._source.roles.put(params.role, new HashMap());} ctx._source.roles.get(params.role).put(params.address,params.timestamp) ",`+
			`"lang": "painless",`+
			`"params": { "role": "%s", "address": "%s", "timestamp": %d }},`+
			`"upsert": { "roles": {"%s": {"%s": %d}}}}`,
			role, rd.Address, timestamp, role, rd.Address, timestamp)
	} else {
		serializedDataStr = fmt.Sprintf(`{"script": {`+
			`"source": "if (ctx._source.containsKey('roles')) { if (ctx._source.roles.containsKey(params.role)) { ctx._source.roles.get(params.role).remove(params.address); } } ",`+
			`"lang": "painless",`+
			`"params": { "role": "%s", "address": "%s" }},`+
			`"upsert": {} }`,
			role, rd.Address)
	}

	err := buffSlice.PutData(meta, []byte(serializedDataStr))
	if err != nil {
		return err
	}

	return nil
}
