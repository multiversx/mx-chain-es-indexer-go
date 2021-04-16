package accounts

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

// SerializeAccounts will serialize the provided accounts in a way that Elastic Search expects a bulk request
func (ap *accountsProcessor) SerializeAccounts(
	accounts map[string]*data.AccountInfo,
	areESDTAccounts bool,
) ([]*bytes.Buffer, error) {
	buffSlice := data.NewBufferSlice()
	for address, acc := range accounts {
		meta, serializedData, err := prepareSerializedAccountInfo(address, acc, areESDTAccounts)
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

func prepareSerializedAccountInfo(
	address string,
	account *data.AccountInfo,
	isESDTAccount bool,
) ([]byte, []byte, error) {
	id := address
	if isESDTAccount {
		id += fmt.Sprintf("_%s", account.TokenIdentifier)
	}

	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, id, "\n"))
	serializedData, err := json.Marshal(account)
	if err != nil {
		return nil, nil, err
	}

	return meta, serializedData, nil
}

// SerializeAccountsHistory will serialize accounts history in a way that Elastic Search expects a bulk request
func (ap *accountsProcessor) SerializeAccountsHistory(
	accounts map[string]*data.AccountBalanceHistory,
) ([]*bytes.Buffer, error) {
	var err error

	buffSlice := data.NewBufferSlice()
	for _, acc := range accounts {
		meta, serializedData, errPrepareAcc := prepareSerializedAccountBalanceHistory(acc)
		if errPrepareAcc != nil {
			return nil, err
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			return nil, err
		}
	}

	return buffSlice.Buffers(), nil
}

func prepareSerializedAccountBalanceHistory(
	account *data.AccountBalanceHistory,
) ([]byte, []byte, error) {
	// no '_id' is specified because an elastic client would never search after the identifier for this index.
	// this is also an improvement: more details here:
	// https://www.elastic.co/guide/en/elasticsearch/reference/master/tune-for-indexing-speed.html#_use_auto_generated_ids
	meta := []byte(fmt.Sprintf(`{ "index" : { } }%s`, "\n"))

	serializedData, err := json.Marshal(account)
	if err != nil {
		return nil, nil, err
	}

	return meta, serializedData, nil
}
