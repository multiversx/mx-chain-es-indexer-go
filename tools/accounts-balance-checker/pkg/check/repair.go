package check

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/multiversx/mx-chain-es-indexer-go/data"
)

func (bc *balanceChecker) deleteExtraBalance(addr, identifier string, timestamp uint64, index string) error {
	if !bc.doRepair {
		return nil
	}

	id := prepareID(addr, identifier)
	meta := []byte(fmt.Sprintf(`{ "update" : {"_index":"%s", "_id" : "%s" } }%s`, index, id, "\n"))
	serializedDataStr := fmt.Sprintf(`{"scripted_upsert": true, "script": {`+
		`"source": "if ( ctx.op == 'create' )  { ctx.op = 'noop' } else { if (ctx._source.containsKey('timestamp')) { if (ctx._source.timestamp < params.timestamp ) { ctx.op = 'delete'  } } else {  ctx.op = 'delete' } }",`+
		`"lang": "painless",`+
		`"params": {"timestamp": %d}},`+
		`"upsert": {}}`,
		timestamp,
	)
	if timestamp == 0 {
		serializedDataStr = fmt.Sprintf(`{"scripted_upsert": true, "script": {` +
			`"source": "if ( ctx.op == 'create' )  { ctx.op = 'noop' } else { ctx.op = 'delete' }",` +
			`"lang": "painless"},` +
			`"upsert": {}}`,
		)
	}

	buffSlice := data.NewBufferSlice(0)
	_ = buffSlice.PutData(meta, []byte(serializedDataStr))

	err := bc.esClient.DoBulkRequest(buffSlice.Buffers()[0], index)
	if err != nil {
		return err
	}

	log.Info("deleted", "id", id)

	return nil
}

func (bc *balanceChecker) fixWrongBalance(addr, identifier string, timestamp uint64, balanceFromProxy string, index string) error {
	if !bc.doRepair {
		return nil
	}

	balanceBig, _ := big.NewInt(0).SetString(balanceFromProxy, 10)
	balanceFloat := bc.balanceToFloat.ComputeESDTBalanceAsFloat(balanceBig)
	if identifier == "" {
		balanceFloat = bc.balanceToFloat.ComputeBalanceAsFloat(balanceBig)
	}

	id := prepareID(addr, identifier)
	meta := []byte(fmt.Sprintf(`{ "update" : {"_index":"%s", "_id" : "%s" } }%s`, index, id, "\n"))
	serializedDataStr := fmt.Sprintf(`{"scripted_upsert": true, "script": {`+
		`"source": "if (ctx.op == 'create') { ctx.op = 'noop'} else { if (ctx._source.containsKey('timestamp')) { if (ctx._source.timestamp < params.timestamp) {ctx._source.timestamp = params.timestamp;ctx._source.balance = params.balanceStr;ctx._source.balanceNum = params.balanceFloat;}} else {ctx._source.timestamp = params.timestamp; ctx._source.balance = params.balanceStr; ctx._source.balanceNum = params.balanceFloat;}}",`+
		`"lang": "painless",`+
		`"params": {"timestamp": %d, "balanceStr": "%s", "balanceFloat": %.10f}},`+
		`"upsert": {}}`,
		timestamp, balanceFromProxy, balanceFloat,
	)

	buffSlice := data.NewBufferSlice(0)
	_ = buffSlice.PutData(meta, []byte(serializedDataStr))

	err := bc.esClient.DoBulkRequest(buffSlice.Buffers()[0], index)
	if err != nil {
		return err
	}

	log.Info("updated", "id", id)

	return nil
}

func prepareID(addr, identifier string) string {
	id := addr

	if identifier == "" {
		return id

	}
	id += "-" + identifier

	split := strings.Split(identifier, "-")
	if len(split) == 2 {
		id += "-00"
	}

	return id
}
