package logsevents

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/process/tokeninfo"
	"github.com/ElrondNetwork/elrond-go-core/core"
)

// SerializeLogs will serialize the provided logs in a way that Elastic Search expects a bulk request
func (logsAndEventsProcessor) SerializeLogs(logs []*data.Logs, buffSlice *data.BufferSlice, index string) error {
	for _, lg := range logs {
		meta := []byte(fmt.Sprintf(`{ "update" : { "_index":"%s", "_id" : "%s" } }%s`, index, converters.JsonEscape(lg.ID), "\n"))
		serializedData, errMarshal := json.Marshal(lg)
		if errMarshal != nil {
			return errMarshal
		}

		codeToExecute := `
		if ('create' == ctx.op) {
			ctx._source = params.log
		} else {
			if (ctx._source.containsKey('timestamp')) {
				if (ctx._source.timestamp <= params.account.timestamp) {
					ctx._source = params.log
				}
			} else {
				ctx._source = params.log
			}
		}
`
		serializedDataStr := fmt.Sprintf(`{"scripted_upsert": true, "script": {`+
			`"source": "%s",`+
			`"lang": "painless",`+
			`"params": { "log": %s }},`+
			`"upsert": {}}`,
			converters.FormatPainlessSource(codeToExecute), serializedData,
		)

		err := buffSlice.PutData(meta, []byte(serializedDataStr))
		if err != nil {
			return err
		}
	}

	return nil
}

// SerializeSCDeploys will serialize the provided smart contract deploys in a way that Elastic Search expects a bulk request
func (logsAndEventsProcessor) SerializeSCDeploys(deploys map[string]*data.ScDeployInfo, buffSlice *data.BufferSlice, index string) error {
	for scAddr, deployInfo := range deploys {
		meta := []byte(fmt.Sprintf(`{ "update" : { "_index":"%s", "_id" : "%s" } }%s`, index, converters.JsonEscape(scAddr), "\n"))

		serializedData, err := serializeDeploy(deployInfo)
		if err != nil {
			return err
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			return err
		}
	}

	return nil
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

	codeToExecute := `
		if (!ctx._source.containsKey('upgrades')) {
			ctx._source.upgrades = [params.elem];
		} else {
			ctx._source.upgrades.add(params.elem);
		}
`
	serializedDataStr := fmt.Sprintf(`{"script": {`+
		`"source": "%s",`+
		`"lang": "painless",`+
		`"params": {"elem": %s}},`+
		`"upsert": %s}`,
		converters.FormatPainlessSource(codeToExecute), string(upgradeSerialized), string(serializedData))

	return []byte(serializedDataStr), nil
}

// SerializeTokens will serialize the provided tokens' data in a way that Elasticsearch expects a bulk request
func (logsAndEventsProcessor) SerializeTokens(tokens []*data.TokenInfo, updateNFTData []*data.NFTDataUpdate, buffSlice *data.BufferSlice, index string) error {
	for _, tokenData := range tokens {
		meta, serializedData, err := serializeToken(tokenData, index)
		if err != nil {
			return err
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			return err
		}
	}

	return converters.PrepareNFTUpdateData(buffSlice, updateNFTData, false, index)
}

func serializeToken(tokenData *data.TokenInfo, index string) ([]byte, []byte, error) {
	if tokenData.TransferOwnership {
		return serializeTokenTransferOwnership(tokenData, index)
	}

	meta := []byte(fmt.Sprintf(`{ "update" : { "_index":"%s", "_id" : "%s" } }%s`, index, converters.JsonEscape(tokenData.Token), "\n"))
	serializedTokenData, err := json.Marshal(tokenData)
	if err != nil {
		return nil, nil, err
	}

	codeToExecute := `
		if (ctx._source.containsKey('roles')) {
			HashMap roles = ctx._source.roles;
			ctx._source = params.token;
			ctx._source.roles = roles
		}
`
	serializedDataStr := fmt.Sprintf(`{"script": {`+
		`"source": "%s",`+
		`"lang": "painless",`+
		`"params": {"token": %s}},`+
		`"upsert": %s}`,
		converters.FormatPainlessSource(codeToExecute), string(serializedTokenData), string(serializedTokenData))

	return meta, []byte(serializedDataStr), nil
}

func serializeTokenTransferOwnership(tokenData *data.TokenInfo, index string) ([]byte, []byte, error) {
	meta := []byte(fmt.Sprintf(`{ "update" : { "_index":"%s", "_id" : "%s" } }%s`, index, converters.JsonEscape(tokenData.Token), "\n"))
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

	codeToExecute := `
		if (!ctx._source.containsKey('ownersHistory')) {
			ctx._source.ownersHistory = [params.elem]
		} else {
			ctx._source.ownersHistory.add(params.elem)
		}
		ctx._source.currentOwner = params.owner
`
	serializedDataStr := fmt.Sprintf(`{"script": {`+
		`"source": "%s",`+
		`"lang": "painless",`+
		`"params": {"elem": %s, "owner": "%s"}},`+
		`"upsert": %s}`,
		converters.FormatPainlessSource(codeToExecute), string(ownerDataSerialized), converters.JsonEscape(tokenData.CurrentOwner), string(tokenDataSerialized))

	return meta, []byte(serializedDataStr), nil
}

// SerializeDelegators will serialize the provided delegators in a way that Elasticsearch expects a bulk request
func (lep *logsAndEventsProcessor) SerializeDelegators(delegators map[string]*data.Delegator, buffSlice *data.BufferSlice, index string) error {
	for _, delegator := range delegators {
		meta, serializedData, err := lep.prepareSerializedDelegator(delegator, index)
		if err != nil {
			return err
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			return err
		}
	}

	return nil
}

func (lep *logsAndEventsProcessor) prepareSerializedDelegator(delegator *data.Delegator, index string) ([]byte, []byte, error) {
	id := lep.computeDelegatorID(delegator)
	if delegator.ShouldDelete {
		meta := []byte(fmt.Sprintf(`{ "delete" : { "_index": "%s", "_id" : "%s" } }%s`, index, converters.JsonEscape(id), "\n"))
		return meta, nil, nil
	}

	meta := []byte(fmt.Sprintf(`{ "index" : { "_index": "%s", "_id" : "%s" } }%s`, index, converters.JsonEscape(id), "\n"))
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
func (lep *logsAndEventsProcessor) SerializeSupplyData(tokensSupply data.TokensHandler, buffSlice *data.BufferSlice, index string) error {
	for _, supplyData := range tokensSupply.GetAll() {
		if supplyData.Type != core.NonFungibleESDT {
			continue
		}

		meta := []byte(fmt.Sprintf(`{ "delete" : { "_index": "%s", "_id" : "%s" } }%s`, index, converters.JsonEscape(supplyData.Identifier), "\n"))
		err := buffSlice.PutData(meta, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// SerializeRolesData will serialize the provided roles data
func (lep *logsAndEventsProcessor) SerializeRolesData(
	tokenRolesAndProperties *tokeninfo.TokenRolesAndProperties,
	buffSlice *data.BufferSlice,
	index string,
) error {
	for role, roleData := range tokenRolesAndProperties.GetRoles() {
		for _, rd := range roleData {
			err := serializeRoleData(buffSlice, rd, role, index)
			if err != nil {
				return err
			}
		}
	}

	for _, tokenAndProp := range tokenRolesAndProperties.GetAllTokensWithProperties() {
		err := serializePropertiesData(buffSlice, index, tokenAndProp)
		if err != nil {
			return err
		}
	}

	return nil
}

func serializeRoleData(buffSlice *data.BufferSlice, rd *tokeninfo.RoleData, role string, index string) error {
	meta := []byte(fmt.Sprintf(`{ "update" : {"_index": "%s", "_id" : "%s" } }%s`, index, converters.JsonEscape(rd.Token), "\n"))
	var serializedDataStr string
	if rd.Set {
		codeToExecute := `	
			if (!ctx._source.containsKey('roles')) {
				ctx._source.roles = new HashMap();
			}
			if (!ctx._source.roles.containsKey(params.role)) {
				ctx._source.roles.put(params.role, [params.address]);
			} else {
				int i;
				for (i = 0; i < ctx._source.roles.get(params.role).length; i++) {
					if (ctx._source.roles.get(params.role).get(i) == params.address) {
						return;
					}
				}
				ctx._source.roles.get(params.role).add(params.address);
			}
`
		serializedDataStr = fmt.Sprintf(`{"script": {`+
			`"source": "%s",`+
			`"lang": "painless",`+
			`"params": { "role": "%s", "address": "%s"}},`+
			`"upsert": { "roles": {"%s": ["%s"]}}}`,
			converters.FormatPainlessSource(codeToExecute),
			converters.JsonEscape(role),
			converters.JsonEscape(rd.Address),
			converters.JsonEscape(role),
			converters.JsonEscape(rd.Address),
		)
	} else {
		codeToExecute := `
	if (ctx._source.containsKey('roles')) {
		if (ctx._source.roles.containsKey(params.role)) {
			ctx._source.roles.get(params.role).removeIf(p -> p.equals(params.address))
		}
	}
`
		serializedDataStr = fmt.Sprintf(`{"script": {`+
			`"source": "%s",`+
			`"lang": "painless",`+
			`"params": { "role": "%s", "address": "%s" }},`+
			`"upsert": {} }`,
			converters.FormatPainlessSource(codeToExecute),
			converters.JsonEscape(role),
			converters.JsonEscape(rd.Address),
		)
	}

	return buffSlice.PutData(meta, []byte(serializedDataStr))
}

func serializePropertiesData(buffSlice *data.BufferSlice, index string, tokenProp *tokeninfo.PropertiesData) error {
	meta := []byte(fmt.Sprintf(`{ "update" : {"_index": "%s", "_id" : "%s" } }%s`, index, converters.JsonEscape(tokenProp.Token), "\n"))

	propertiesBytes, err := json.Marshal(tokenProp.Properties)
	if err != nil {
		return err
	}

	codeToExecute := `	
			if (!ctx._source.containsKey('properties')) {
				ctx._source.properties = new HashMap();
			}
			params.properties.forEach(
				(key, value) -> ctx._source.properties[key] = value
			);
`
	serializedDataStr := fmt.Sprintf(`{"script": {`+
		`"source": "%s",`+
		`"lang": "painless",`+
		`"params": { "properties": %s}},`+
		`"upsert": {}}}`,
		converters.FormatPainlessSource(codeToExecute), propertiesBytes)

	return buffSlice.PutData(meta, []byte(serializedDataStr))
}
