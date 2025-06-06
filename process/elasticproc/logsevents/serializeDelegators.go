package logsevents

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
)

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

// PrepareDelegatorsQueryInCaseOfRevert will prepare the delegators query in case of revert
func (lep *logsAndEventsProcessor) PrepareDelegatorsQueryInCaseOfRevert(timestamp uint64) *bytes.Buffer {
	codeToExecute := `
	if ( !ctx._source.containsKey('unDelegateInfo') ) { return } 
	if ( ctx._source.unDelegateInfo.length == 0 ) { return }
	ctx._source.unDelegateInfo.removeIf(info -> info.timestamp.equals(params.timestamp));
`

	query := fmt.Sprintf(`
	{
	  "query": {
		"match": {
		  "timestamp": "%d"
		}
	  },
	  "script": {
		"source": "%s",
		"lang": "painless",
		"params": {"timestamp": %d}
	  }
	}`, timestamp, converters.FormatPainlessSource(codeToExecute), timestamp)

	return bytes.NewBuffer([]byte(query))
}

func (lep *logsAndEventsProcessor) prepareSerializedDelegator(delegator *data.Delegator, index string) ([]byte, []byte, error) {
	id := lep.computeDelegatorID(delegator)
	if delegator.ShouldDelete {
		meta := []byte(fmt.Sprintf(`{ "delete" : { "_index": "%s", "_id" : "%s" } }%s`, index, converters.JsonEscape(id), "\n"))
		return meta, nil, nil
	}

	delegatorSerialized, errMarshal := json.Marshal(delegator)
	if errMarshal != nil {
		return nil, nil, errMarshal
	}

	meta := []byte(fmt.Sprintf(`{ "update" : { "_index":"%s", "_id" : "%s" } }%s`, index, converters.JsonEscape(id), "\n"))
	if delegator.UnDelegateInfo != nil {
		serializedData, err := prepareSerializedDataForUnDelegate(delegator, delegatorSerialized)
		return meta, serializedData, err
	}
	if delegator.WithdrawFundIDs != nil {
		serializedData, err := prepareSerializedDataForWithdrawal(delegator, delegatorSerialized)
		return meta, serializedData, err
	}

	return meta, prepareSerializedDataForDelegator(delegatorSerialized), nil
}

func prepareSerializedDataForDelegator(delegatorSerialized []byte) []byte {
	// TODO add timestampMs
	codeToExecute := `
		if ('create' == ctx.op) {
			ctx._source = params.delegator
		} else {
			ctx._source.activeStake = params.delegator.activeStake;
			ctx._source.activeStakeNum = params.delegator.activeStakeNum;
			ctx._source.timestamp = params.delegator.timestamp;
		}
`
	serializedDataStr := fmt.Sprintf(`{"scripted_upsert": true, "script": {`+
		`"source": "%s",`+
		`"lang": "painless",`+
		`"params": { "delegator": %s }},`+
		`"upsert": {}}`,
		converters.FormatPainlessSource(codeToExecute), string(delegatorSerialized),
	)

	return []byte(serializedDataStr)
}

func prepareSerializedDataForUnDelegate(delegator *data.Delegator, delegatorSerialized []byte) ([]byte, error) {
	unDelegateInfoSerialized, err := json.Marshal(delegator.UnDelegateInfo)
	if err != nil {
		return nil, err
	}

	// TODO add timestampMs
	codeToExecute := `
		if ('create' == ctx.op) {
			ctx._source = params.delegator
		} else {
			if (!ctx._source.containsKey('unDelegateInfo')) {
				ctx._source.unDelegateInfo = [params.unDelegate];
			} else {
				ctx._source.unDelegateInfo.add(params.unDelegate);
			}

			ctx._source.activeStake = params.delegator.activeStake;
			ctx._source.activeStakeNum = params.delegator.activeStakeNum;
			ctx._source.timestamp = params.delegator.timestamp;
		}
`
	serializedDataStr := fmt.Sprintf(`{"scripted_upsert": true, "script": {`+
		`"source": "%s",`+
		`"lang": "painless",`+
		`"params": { "delegator": %s, "unDelegate": %s }},`+
		`"upsert": {}}`,
		converters.FormatPainlessSource(codeToExecute), string(delegatorSerialized), string(unDelegateInfoSerialized),
	)

	return []byte(serializedDataStr), nil
}

func prepareSerializedDataForWithdrawal(delegator *data.Delegator, delegatorSerialized []byte) ([]byte, error) {
	withdrawFundIDsSerialized, err := json.Marshal(delegator.WithdrawFundIDs)
	if err != nil {
		return nil, err
	}

	codeToExecute := `
		if ('create' == ctx.op) {
			ctx._source = params.delegator
		} else {
			if (!ctx._source.containsKey('unDelegateInfo')) {
				return
			} else {
				Iterator itr = ctx._source.unDelegateInfo.iterator();
				while (itr.hasNext()) {
					HashMap unDelegate = itr.next();
					for (int j = 0; j < params.withdrawIds.length; j++) {
					  if (unDelegate.id == params.withdrawIds[j]) {
						itr.remove();
					  }
					}
				}
				if (ctx._source.unDelegateInfo.length == 0) {ctx._source.remove('unDelegateInfo')}
			}

			ctx._source.activeStake = params.delegator.activeStake;
			ctx._source.activeStakeNum = params.delegator.activeStakeNum;
			ctx._source.timestamp = params.delegator.timestamp;
		}
`
	serializedDataStr := fmt.Sprintf(`{"scripted_upsert": true, "script": {`+
		`"source": "%s",`+
		`"lang": "painless",`+
		`"params": { "delegator": %s, "withdrawIds": %s }},`+
		`"upsert": {}}`,
		converters.FormatPainlessSource(codeToExecute), string(delegatorSerialized), string(withdrawFundIDsSerialized),
	)

	return []byte(serializedDataStr), nil
}

func (lep *logsAndEventsProcessor) computeDelegatorID(delegator *data.Delegator) string {
	delegatorContract := delegator.Address + delegator.Contract

	hashBytes := lep.hasher.Compute(delegatorContract)

	return base64.StdEncoding.EncodeToString(hashBytes)
}
