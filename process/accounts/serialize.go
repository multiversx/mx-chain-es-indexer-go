package accounts

import (
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

// SerializeNFTCreateInfo will serialize the provided nft create information in a way that Elastic Search expects a bulk request
func (ap *accountsProcessor) SerializeNFTCreateInfo(tokensInfo []*data.TokenInfo, buffSlice *data.BufferSlice, index string) error {
	for _, tokenData := range tokensInfo {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_index":"%s", "_id" : "%s" } }%s`, index, tokenData.Identifier, "\n"))
		serializedData, errMarshal := json.Marshal(tokenData)
		if errMarshal != nil {
			return errMarshal
		}

		err := buffSlice.PutData(meta, serializedData)
		if err != nil {
			return err
		}
	}

	return nil
}

// SerializeAccounts will serialize the provided accounts in a way that Elasticsearch expects a bulk request
func (ap *accountsProcessor) SerializeAccounts(accounts map[string]*data.AccountInfo, buffSlice *data.BufferSlice, index string) error {
	for _, acc := range accounts {
		meta, serializedData, err := prepareSerializedAccount(acc, false, index)
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

// SerializeAccountsESDT will serialize the provided accounts and nfts updates in a way that Elasticsearch expects a bulk request
func (ap *accountsProcessor) SerializeAccountsESDT(
	accounts map[string]*data.AccountInfo,
	updateNFTData []*data.NFTDataUpdate,
	buffSlice *data.BufferSlice,
	index string,
) error {
	for _, acc := range accounts {
		meta, serializedData, err := prepareSerializedAccount(acc, true, index)
		if err != nil {
			return err
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			return err
		}
	}

	err := converters.PrepareNFTUpdateData(buffSlice, updateNFTData, true, index)
	if err != nil {
		return err
	}

	return nil
}

func prepareSerializedAccount(acc *data.AccountInfo, isESDT bool, index string) ([]byte, []byte, error) {
	if (acc.Balance == "0" || acc.Balance == "") && isESDT {
		meta, serializedData := prepareDeleteAccountInfo(acc, isESDT, index)
		return meta, serializedData, nil
	}

	return prepareSerializedAccountInfo(acc, isESDT, index)
}

func prepareDeleteAccountInfo(acct *data.AccountInfo, isESDT bool, index string) ([]byte, []byte) {
	id := acct.Address
	if isESDT {
		hexEncodedNonce := converters.EncodeNonceToHex(acct.TokenNonce)
		id += fmt.Sprintf("-%s-%s", acct.TokenName, hexEncodedNonce)
	}

	meta := []byte(fmt.Sprintf(`{ "update" : {"_index":"%s", "_id" : "%s" } }%s`, index, id, "\n"))

	serializedDataStr := fmt.Sprintf(`{"scripted_upsert": true, "script": {`+
		`"source": "if ( ctx.op == 'create' )  { ctx.op = 'noop' } else { if (ctx._source.timestamp < params.timestamp ) { ctx.op = 'delete'  } }",`+
		`"lang": "painless",`+
		`"params": {"timestamp": %d}},`+
		`"upsert": {}}`,
		acct.Timestamp,
	)

	return meta, []byte(serializedDataStr)
}

func prepareSerializedAccountInfo(
	account *data.AccountInfo,
	isESDTAccount bool,
	index string,
) ([]byte, []byte, error) {
	id := account.Address
	if isESDTAccount {
		hexEncodedNonce := converters.EncodeNonceToHex(account.TokenNonce)
		id += fmt.Sprintf("-%s-%s", account.TokenName, hexEncodedNonce)
	}

	serializedAccount, err := json.Marshal(account)
	if err != nil {
		return nil, nil, err
	}

	meta := []byte(fmt.Sprintf(`{ "update" : {"_index": "%s", "_id" : "%s" } }%s`, index, id, "\n"))
	serializedDataStr := fmt.Sprintf(`{"scripted_upsert": true, "script": {`+
		`"source": "if ( ctx.op == 'create' )  { ctx._source = params.account } else { if (ctx._source.timestamp < params.account.timestamp ) { ctx._source = params.account  } }",`+
		`"lang": "painless",`+
		`"params": { "account": %s }},`+
		`"upsert": {}}`,
		serializedAccount,
	)

	return meta, []byte(serializedDataStr), nil
}

// SerializeAccountsHistory will serialize accounts history in a way that Elastic Search expects a bulk request
func (ap *accountsProcessor) SerializeAccountsHistory(
	accounts map[string]*data.AccountBalanceHistory,
	buffSlice *data.BufferSlice,
	index string,
) error {
	var err error

	for _, acc := range accounts {
		meta, serializedData, errPrepareAcc := prepareSerializedAccountBalanceHistory(acc, index)
		if errPrepareAcc != nil {
			return err
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			return err
		}
	}

	return nil
}

func prepareSerializedAccountBalanceHistory(
	account *data.AccountBalanceHistory,
	index string,
) ([]byte, []byte, error) {
	id := account.Address

	isESDT := account.Token != ""
	if isESDT {
		hexEncodedNonce := converters.EncodeNonceToHex(account.TokenNonce)
		id += fmt.Sprintf("-%s-%s", account.Token, hexEncodedNonce)
	}

	id += fmt.Sprintf("-%d", account.Timestamp)
	meta := []byte(fmt.Sprintf(`{ "index" : { "_index":"%s", "_id" : "%s" } }%s`, index, id, "\n"))

	serializedData, err := json.Marshal(account)
	if err != nil {
		return nil, nil, err
	}

	return meta, serializedData, nil
}

// SerializeTypeForProvidedIDs will serialize the type for the provided ids
func (ap *accountsProcessor) SerializeTypeForProvidedIDs(
	ids []string,
	tokenType string,
	buffSlice *data.BufferSlice,
	index string,
) error {
	for _, id := range ids {
		meta := []byte(fmt.Sprintf(`{ "update" : {"_index":"%s", "_id" : "%s" } }%s`, index, id, "\n"))

		serializedDataStr := fmt.Sprintf(`{"scripted_upsert": true, "script": {`+
			`"source": "if ( ctx.op == 'create' )  { ctx.op = 'noop' } else  { ctx._source.type = params.type }",`+
			`"lang": "painless",`+
			`"params": {"type": "%s"}},`+
			`"upsert": {}}`,
			tokenType)

		err := buffSlice.PutData(meta, []byte(serializedDataStr))
		if err != nil {
			return err
		}
	}

	return nil
}
